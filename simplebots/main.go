package main

import (
	"flag"
	"time"

	"github.com/bcspragu/Gobots/game"
)

var (
	token   = flag.String("token", "", "which token to connect to the server with")
	addr    = flag.String("addr", "localhost:8001", "The address of the game server")
	botName = flag.String("bot_name", "aggro", "which bot to use")
)

func main() {
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
	game.Connect(*botName, *token, game.ToFactory(g), &game.ServerConfig{
		ServerAddress: *addr,
		RetryInterval: 10 * time.Second,
	})
}
