package main

import (
	"net/http"
	"strconv"
	"strings"
)

// I'm sure I'll need this eventually, and am definitely not prematurely
// optimizing.
type context struct {
	magicToken string
	r          *http.Request
	w          http.ResponseWriter

	p *player
}

func newContext(w http.ResponseWriter, r *http.Request) context {
	return context{
		w: w,
		r: r,
	}
}

type player struct {
	Name      string   // GitHub username, probably
	Endpoints []string // Servers they have hosting AI
}

func (c *context) gameID() gameID {
	return gameID(strings.Split(c.r.URL.Path, "/")[2])
}

func (c *context) roundNumber() int {
	parts := strings.Split(c.r.URL.Path, "/")
	if len(parts) < 4 {
		return -1
	}
	if r, err := strconv.Atoi(parts[3]); err == nil {
		return r
	}
	return -1
}
