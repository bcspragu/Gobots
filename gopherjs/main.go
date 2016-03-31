package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/gopherjs/gopherjs/js"

	"github.com/bcspragu/Gobots/engine"
)

func main() {
	// TODO(bsprague): Add things to this handy global JS object when you get
	// even remotely that far
	js.Global.Set("Gobot", map[string]interface{}{
		"GetBoard": GetBoard,
	})
}

func GetBoard(base64Board string) *js.Object {
	var board engine.Board
	buf := bytes.NewReader([]byte(base64Board))
	r := base64.NewDecoder(base64.StdEncoding, buf)
	gob.NewDecoder(r).Decode(&board)
	return js.MakeWrapper(&board)
}
