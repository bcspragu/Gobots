package main

import (
	cryptorand "crypto/rand"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/securecookie"
)

// I was going to call it cookieData, but data about a cookie is just the
// nutrition facts #HowDidIGetAProfessionalCodingJob
type nutritionFacts struct {
	AccessToken string
}

func requireLogin(handler func(c context) error) func(c context) error {
	return func(c context) error {
		// Do some best-effort context-filling
		if nutFact, err := loadCookie(c.r); err == nil {
			c.token = accessToken(nutFact.AccessToken)
			if nutFact.AccessToken != "" {
				if u, err := db.loadUser(accessToken(nutFact.AccessToken)); err != nil {
					http.Redirect(c.w, c.r, "/", http.StatusFound)
					return err
				} else if u == nil {
					log.Printf("Error: User tried to access %s without being logged in\n", c.r.URL.Path)
					http.Redirect(c.w, c.r, "/", http.StatusFound)
					return nil
				}
			}
		}
		return handler(c)
	}
}

func baseWrapper(handler func(c context) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := newContext(w, r)
		if nutFact, err := loadCookie(c.r); err == nil {
			c.token = accessToken(nutFact.AccessToken)
			if nutFact.AccessToken != "" {
				if u, err := db.loadUser(accessToken(nutFact.AccessToken)); err == nil {
					c.u = u
				} else {
					log.Printf("Error: %v\n", err)
				}
			}
		}
		if err := handler(c); err != nil {
			serveError(c.w, err)
		}
	}
}

func noUserWrapper(handler func(c context) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := newContext(w, r)
		if err := handler(c); err != nil {
			serveError(c.w, err)
		}
	}
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func genName(n int) string {
	b := make([]byte, n)
	r := rand.New(cryptoRandSource{})
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func initKeys(hashPath, blockPath string) (*securecookie.SecureCookie, error) {
	var hashKey []byte
	var blockKey []byte

	if dat, err := loadOrGenKey(hashPath); err != nil {
		return nil, err
	} else {
		hashKey = dat
	}

	if dat, err := loadOrGenKey(blockPath); err != nil {
		return nil, err
	} else {
		blockKey = dat
	}

	return securecookie.New(hashKey, blockKey), nil
}

func loadOrGenKey(name string) ([]byte, error) {
	if f, err := ioutil.ReadFile(name); err != nil {
		if dat := securecookie.GenerateRandomKey(32); dat != nil {
			if err := ioutil.WriteFile(name, dat, 0777); err == nil {
				return dat, nil
			}
			return nil, errors.New("Error writing file")
		}
		return nil, errors.New("Failed to generate key")
	} else {
		return f, nil
	}
}

func loadCookie(r *http.Request) (nutritionFacts, error) {
	if cookie, err := r.Cookie("info"); err == nil {
		value := nutritionFacts{}
		if err = s.Decode("info", cookie.Value, &value); err != nil {
			return nutritionFacts{}, err
		}
		return value, nil
	} else {
		return nutritionFacts{}, err
	}
}

type cryptoRandSource struct{}

func (cryptoRandSource) Int63() int64 {
	var buf [8]byte
	_, err := cryptorand.Read(buf[:])
	if err != nil {
		panic(err)
	}
	return int64(buf[0]) |
		int64(buf[1])<<8 |
		int64(buf[2])<<16 |
		int64(buf[3])<<24 |
		int64(buf[4])<<32 |
		int64(buf[5])<<40 |
		int64(buf[6])<<48 |
		int64(buf[7]&0x7f)<<56
}

func (cryptoRandSource) Seed(int64) {}
