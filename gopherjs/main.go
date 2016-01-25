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

func GetBoard(jsonBoard string) *js.Object {
	var board *engine.Board
	json.Unmarshal([]byte(jsonBoard), &board)
	return js.MakeWrapper(board)
}
