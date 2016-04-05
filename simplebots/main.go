package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/bcspragu/Gobots/game"
)

var token string
var flags = flag.NewFlagSet("program flags", flag.ContinueOnError)
var botName = flags.String("bot_name", "aggro", "which bot to use")

func main() {
	flags.StringVar(&token, "token", "", "which token to connect to the server with")
	flags.SetOutput(ioutil.Discard)
	flags.Parse(os.Args[1:])
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
	game.StartServerForFactory(*botName, token, game.ToFactory(g))

	//res := game.FightBotsN(game.ToFactory(g), game.ToFactory(&pathfinder{}), 1)
	//for _, m := range res {
	//fmt.Println(m.String())
	//}
}
