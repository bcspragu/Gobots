package main

import (
	"flag"

	"github.com/bcspragu/Gobots/game"
)

var botName = flag.String("bot_name", "aggro", "which bot to use")

func main() {
	flag.Parse()
	code := "IAqDpTlpWCfYnRvUflMpIZdwU"
	var g game.AI

	switch *botName {
	case "aggro":
		g = aggro{}
	case "random":
		g = random{}
	case "pathfinder":
		g = &pathfinder{}
	}
	game.StartServerForBot(*botName, code, g)
}
