package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	gocontext "golang.org/x/net/context"

	"github.com/bcspragu/Gobots/engine"
	"github.com/gorilla/securecookie"
)

var (
	addr      = flag.String("addr", ":8000", "HTTP server address")
	apiAddr   = flag.String("api_addr", ":8001", "RPC server address")
	templates = tmpl{template.Must(template.ParseGlob("templates/*.html"))}

	db               datastore
	secretz          string
	s                *securecookie.SecureCookie
	globalAIEndpoint *aiEndpoint
)

const (
	clientId = "07ef388cb32ffbbd5146"
)

func main() {
	flag.Parse()
	var err error

	if db, err = initDB("gobots.db"); err != nil {
		log.Fatal("Couldn't open the database, SHUT IT DOWN")
	}

	if secretz, err = initSecretz(); err != nil {
		log.Fatal("Ain't got no GitHub client secret!!")
	}

	if s, err = initKeys(); err != nil {
		log.Fatal("Can't encrypt the cookies! WHATEVER WILL WE DO")
	}

	http.HandleFunc("/", withLogin(serveIndex))
	http.HandleFunc("/game/", withLogin(serveGame))
	http.HandleFunc("/auth", withLogin(serveAuth))
	http.HandleFunc("/loadBots", withLogin(loadBots))
	http.HandleFunc("/startMatch", withLogin(startMatch))

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))

	globalAIEndpoint, err = startAIEndpoint(*apiAddr, db)
	if err != nil {
		log.Fatal("AI RPC endpoint failed to start:", err)
	}

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("Yeah...so about that whole server thing: ", err)
	}
}

func serveIndex(c context) {
	data := tmplData{
		Data: map[string]interface{}{
			"Bots": globalAIEndpoint.listOnlineAIs(),
		},
		Scripts: []template.URL{
			"/js/main.js",
		},
	}

	if err := templates.ExecuteTemplate(c, "index.html", data); err != nil {
		serveError(c.w, err)
	}
}

func serveGame(c context) {
	replay, err := db.lookupGame(c.gameID())
	if c.roundNumber() == -1 {
		data := tmplData{
			Data: map[string]interface{}{
				"GameID": c.gameID(),
				"Exists": err != errDatastoreNotFound,
			},
		}
		if err := templates.ExecuteTemplate(c, "game.html", data); err != nil {
			serveError(c.w, err)
		}
		return
	}

	// If we're here, they're looking for a single boards encoding
	board, err := engine.NewPlayback(replay).Board(c.roundNumber())
	if err != nil {
		serveError(c.w, err)
	}
	d, err := json.Marshal(board.ToJSONBoard())
	if err != nil {
		serveError(c.w, err)
	}
	c.w.Write(d)
}

func serveError(w http.ResponseWriter, err error) {
	w.Write([]byte("Internal Server Error"))
	log.Printf("Error: %v\n", err)
}

func startMatch(c context) {
	//TODO DOIAFJHJKSHLAJSDLKJASLKDJ
	ai1, _ := db.lookupAI(aiID(c.r.FormValue("ai1")))
	ai2, _ := db.lookupAI(aiID(c.r.FormValue("ai2")))
	online := globalAIEndpoint.listOnlineAIs()
	var o1, o2 *onlineAI
	for _, v := range online {
		if v.Info.ID == ai1.ID {
			o1 = &v
		} else if v.Info.ID == ai2.ID {
			o2 = &v
		}
	}

	gidCh := make(chan gameID)
	go func() {
		err := runMatch(gidCh, gocontext.TODO(), db, o1, o2)
		close(gidCh)
		if err != nil {
			log.Println("runMatch:", err)
		}
	}()
	gid := <-gidCh
	if gid == "" {
		serveError(c.w, errors.New("game couldn't start"))
	} else {
		http.Redirect(c.w, c.r, "/game/"+string(gid), http.StatusFound)
	}
}

func loadBots(c context) {
	uid := userID(c.p.Name)
	_, token, err := db.createAI(uid, &aiInfo{Nick: c.r.PostFormValue("nick")})
	if err != nil {
		serveError(c.w, err)
		return
	}
	fmt.Fprintln(c.w, "Congrats, your token is:", token)
}

func serveAuth(c context) {
	if c.r.FormValue("state") != c.magicToken {
		log.Println("They're spoofing GitHub's API. I AM THE ONE WHO KNOCKS (on GitHub's API server)")
		return
	}

	resp, err := http.PostForm("https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {clientId},
		"client_secret": {secretz},
		"code":          {c.r.FormValue("code")},
	})
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	sRep := string(d)
	// http://i.imgur.com/c4jt321.png
	is, ie := strings.Index(sRep, "=")+1, strings.Index(sRep, "&")
	accessToken := sRep[is:ie]

	nutFact, err := loadCookie(c.r)
	if err != nil {
		log.Println(err)
		// This is weird, they must have cookies off, which makes it hard for us to
		// validate them. Screw 'em for now
		return
	}

	// Set their access token to what we just got, and create a user with that token
	if nutFact.AccessToken == "" {
		nutFact.AccessToken = accessToken
		go db.createUser(userID(accessToken))
	}

	if encoded, err := s.Encode("info", nutFact); err == nil {
		cookie := &http.Cookie{
			Name:  "info",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(c.w, cookie)
	}

	http.Redirect(c.w, c.r, "/", http.StatusFound)
}

func username(uID userID) string {
	resp, err := http.Get("https://api.github.com/user?access_token=" + string(uID))
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Println(err)
		return ""
	}

	return data["login"].(string)
}
