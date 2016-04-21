package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"

	gocontext "golang.org/x/net/context"

	"github.com/bcspragu/Gobots/engine"
	"github.com/gorilla/securecookie"
)

var (
	addr      = flag.String("addr", ":8000", "HTTP server address")
	apiAddr   = flag.String("api_addr", ":8001", "RPC server address")
	templates = tmpl{template.Must(template.ParseGlob("templates/*.html"))}

	db               datastore
	s                *securecookie.SecureCookie
	globalAIEndpoint *aiEndpoint
)

func main() {
	flag.Parse()
	var err error

	if db, err = initDB("gobots.db"); err != nil {
		log.Fatal("Couldn't open the database, SHUT IT DOWN")
	}

	if s, err = initKeys(); err != nil {
		log.Fatal("Can't encrypt the cookies! WHATEVER WILL WE DO")
	}

	http.HandleFunc("/", baseWrapper(serveIndex))
	http.HandleFunc("/createUser", baseWrapper(createUserHandler))
	http.HandleFunc("/login", baseWrapper(loginHandler))
	http.HandleFunc("/logout", baseWrapper(logoutHandler))
	http.HandleFunc("/bots", baseWrapper(botsHandler))
	http.HandleFunc("/bot/", baseWrapper(botHandler))
	http.HandleFunc("/game/", noUserWrapper(serveGame))
	http.HandleFunc("/startMatch", baseWrapper(requireLogin(startMatch)))

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

func serveIndex(c context) error {
	data := tmplData{
		Data: map[string]interface{}{
			"Bots": globalAIEndpoint.listOnlineAIs(),
		},
		Scripts: []template.URL{
			"/js/main.js",
		},
	}

	return templates.ExecuteTemplate(c, "index.html", data)
}

func serveGame(c context) error {
	replay, err := db.lookupGame(c.gameID())
	if err != nil {
		if err == errGameNotFound {
			data := tmplData{
				Data: map[string]interface{}{
					"GameID": c.gameID(),
					"Exists": false,
				},
			}
			return templates.ExecuteTemplate(c, "game.html", data)
		}
		return err
	}

	p, err := engine.NewPlayback(replay)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(p)
	if err != nil {
		return err
	}

	var dat bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &dat)
	enc.Write(buf.Bytes())
	enc.Close()

	gInfo, err := db.lookupGameInfo(c.gameID())
	if err != nil {
		return err
	}

	data := tmplData{
		Data: map[string]interface{}{
			"GameID":   c.gameID(),
			"Exists":   true,
			"Playback": dat.String(),
			"P1Name":   gInfo.AI1.Name,
			"P2Name":   gInfo.AI2.Name,
		},
	}
	return templates.ExecuteTemplate(c, "game.html", data)
}

func serveError(w http.ResponseWriter, err error) {
	w.Write([]byte("Internal Server Error"))
	log.Printf("Error: %v\n", err)
}

func startMatch(c context) error {
	ai1, _ := db.lookupAI(aiID(c.r.FormValue("ai1")))
	ai2, _ := db.lookupAI(aiID(c.r.FormValue("ai2")))

	online := globalAIEndpoint.listOnlineAIs()
	var o1, o2 *onlineAI
	for _, v := range online {
		// Because we can't rely on the address of v
		vv := v
		if v.Info.ID == ai1.ID {
			o1 = &vv
		}
		if v.Info.ID == ai2.ID {
			o2 = &vv
		}
	}

	gidCh := make(chan gameID)
	matchDone := make(chan struct{})
	go func() {
		// TODO: Have the user choose the config
		err := runMatch(gidCh, gocontext.TODO(), db, o1, o2, engine.DefaultConfig)
		close(gidCh)
		if err != nil {
			log.Println("runMatch:", err)
		}
		matchDone <- struct{}{}
	}()
	gid := <-gidCh
	<-matchDone
	if gid == "" {
		return errors.New("game couldn't start")
	} else {
		http.Redirect(c.w, c.r, "/game/"+string(gid), http.StatusFound)
	}
	return nil
}

func loginHandler(c context) error {
	if c.r.Method != "POST" {
		return errors.New("Wrong method")
	}

	if c.u != nil {
		c.Write("You're already logged in")
		return nil
	}

	token := c.r.PostFormValue("accessToken")
	if u, err := db.loadUser(accessToken(token)); u == nil {
		c.Write("User doesn't exist")
		return nil
	} else if err != nil {
		return err
	}

	nutFact := nutritionFacts{
		AccessToken: token,
	}

	if encoded, err := s.Encode("info", nutFact); err == nil {
		cookie := &http.Cookie{
			Name:  "info",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(c.w, cookie)
	} else {
		return err
	}

	c.Write("Yes.")
	return nil
}

func logoutHandler(c context) error {
	// Clear token
	nutFact := nutritionFacts{
		AccessToken: "",
	}

	if encoded, err := s.Encode("info", nutFact); err == nil {
		cookie := &http.Cookie{
			Name:  "info",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(c.w, cookie)
	} else {
		return err
	}

	http.Redirect(c.w, c.r, "/", http.StatusFound)
	return nil
}

func createUserHandler(c context) error {
	if c.r.Method != "POST" {
		return errors.New("Wrong method")
	}

	if c.u != nil {
		c.Write("You're already logged in.")
		return nil
	}

	name := c.r.PostFormValue("userName")
	if exists, err := db.userExists(name); exists {
		c.Write("User already exists")
		return nil
	} else if err != nil {
		return err
	}

	token := genName(25)
	_, err := db.createUser(&userInfo{
		Name:  name,
		Token: accessToken(token),
	})

	if err != nil {
		return err
	}

	nutFact := nutritionFacts{
		AccessToken: token,
	}

	if encoded, err := s.Encode("info", nutFact); err == nil {
		cookie := &http.Cookie{
			Name:  "info",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(c.w, cookie)
	}

	c.Write("Your access token is: " + token + ". You'll use this to login and to get your bot on the server.")
	return nil
}

func botsHandler(c context) error {
	dir, err := db.loadDirectory()
	if err != nil {
		return err
	}

	isOn := make(map[aiID]bool)

	for _, on := range globalAIEndpoint.listOnlineAIs() {
		isOn[on.Info.ID] = true
	}

	data := tmplData{
		Data: map[string]interface{}{
			"Directory": dir,
			"Online":    isOn,
		},
	}
	return templates.ExecuteTemplate(c, "bots.html", data)
}

func botHandler(c context) error {
	dir, err := db.loadDirectory()
	if err != nil {
		return err
	}

	matches, err := db.matchHistory(c.botID())
	if err != nil {
		return err
	}

	data := tmplData{
		Data: map[string]interface{}{
			"History":   matches,
			"Directory": dir,
			"ID":        c.botID(),
		},
	}
	return templates.ExecuteTemplate(c, "bot.html", data)
}
