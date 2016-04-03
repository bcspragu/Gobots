package main

import (
	"net/http"
	"strings"
)

// I'm sure I'll need this eventually, and am definitely not prematurely
// optimizing.
type context struct {
	token accessToken
	r     *http.Request
	w     http.ResponseWriter

	u *userInfo
}

func newContext(w http.ResponseWriter, r *http.Request) context {
	return context{
		w: w,
		r: r,
	}
}

func (c *context) Write(s string) {
	c.w.Write([]byte(s))
}

func (c *context) gameID() gameID {
	return gameID(strings.Split(c.r.URL.Path, "/")[2])
}
