package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/gopherjs/gopherjs/js"

	"github.com/bcspragu/Gobots/engine"
)

func main() {
	js.Global.Set("Gobot", map[string]interface{}{
		"GetPlayback": GetPlayback,
	})
}

func GetPlayback(base64Playback string) *js.Object {
	var p engine.Playback
	buf := bytes.NewReader([]byte(base64Playback))
	r := base64.NewDecoder(base64.StdEncoding, buf)
	gob.NewDecoder(r).Decode(&p)
	return js.MakeWrapper(&p)
}
