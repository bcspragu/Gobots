package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/bcspragu/Gobots/game"
)

var flags = flag.NewFlagSet("program flags", flag.ContinueOnError)
var botName = flags.String("bot_name", "aggro", "which bot to use")

func main() {
	flags.SetOutput(ioutil.Discard)
	flags.Parse(os.Args[1:])
	code := "HAXWUumlfiJekSXfwAHQlttHH"
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
	game.StartServerForBot(*botName, code, g)
}
