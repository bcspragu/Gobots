package main

import (
	"flag"
	"time"

	"github.com/bcspragu/Gobots/game"
)

var token string
var botName = flag.String("bot_name", "aggro", "which bot to use")

func main() {
	flag.StringVar(&token, "token", "", "which token to connect to the server with")
	flag.Parse()

	var g game.AI
	switch *botName {
	case "aggro":
		g = aggro{}
	case "random":
		g = random{}
	case "pathfinder":
		g = &pathfinder{}
	case "sunguard":
		g = sunguard{}
	}
	game.Connect(*botName, token, game.ToFactory(g), &game.ServerConfig{
		ServerAddress: "localhost:8081",
		RetryInterval: 10 * time.Second,
	})
}
