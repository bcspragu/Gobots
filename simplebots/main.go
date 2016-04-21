package main

import (
	"flag"

	"github.com/bcspragu/Gobots/game"
)

var token string
var botName = flag.String("bot_name", "aggro", "which bot to use")

func main() {
	flag.StringVar(&token, "token", "", "which token to connect to the server with")
	flag.Parse()
	if token == "" {
		token = "HAXWUumlfiJekSXfwAHQlttHH"
	}
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
	game.StartServerForFactoryWithConfig(*botName, token, game.ToFactory(g), game.ServerConfig{
		Addr: "localhost:8001",
	})

	//res := game.FightBotsN(game.ToFactory(g), game.ToFactory(&pathfinder{}), 1)
	//for _, m := range res {
	//fmt.Println(m.String())
	//}
}
