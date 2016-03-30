package main

import (
	"encoding/json"

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

// I think it might make sense to ditch jBoard and encode/decode to/from gob/base64
func GetBoard(jsonBoard string) *js.Object {
	var jBoard *engine.JSONBoard
	json.Unmarshal([]byte(jsonBoard), &jBoard)
	return js.MakeWrapper(jBoard.ToBoard())
}
