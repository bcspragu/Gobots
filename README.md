# Gobots

A programmatic robot-fighting game, **heavily** inspired by [Robot
Game](http://robotgame.net). As of the time of this writing, it isn't online
yet, but it will be at GobotGame.com (hopefully) soon.

[Related Go Talk Slides](https://docs.google.com/a/google.com/presentation/d/1XCBCgk5l17PItL9w_1m9zs1UN2VStiNE_R-H8D4P8VY/edit?usp=sharing)

## Installing Go

Instructions for installing Go can be found
[here](https://golang.org/doc/install). Once Go is installed on your system and
your GOPATH is set properly, run `go get -u github.com/bcspragu/Gobots/...` to
install the relevant packages for Gobots.

## Creating an account on Gobot Game

1. Go to [Gobot Game](http://gobotgame.com) and enter a user name. It'll return
   a unique token that you'll input into `game.StartServerForFactory` which
   authenticates you and identifies you to the server.
2. Your bot will show up in the list of Online Bots, you can start a match with
   another online bot (or yourself), by selecting "Fight Bots".
3. Watch the match play out, tweak bot, repeat!

## Developing a Bot

Several example robots are provided in the `simplebots` subdirectory, though
none of them are even remotely good. Building a bot and connecting it to the
server can be done with the following snippet.

```go

package main

import "github.com/bcspragu/Gobots/game"

// Bot moves to the center and does nothing else
type bot struct{}

func (bot) Act(b *game.Board, r *game.Robot) game.Action {
  return game.Action{
    Kind:      game.Move,
    Direction: game.Towards(r.Loc, b.Center()),
  }
}

func main() {
	game.StartServerForFactory("MyBot", "MyAccessToken", game.ToFactory(bot{}))
}
```

The `game.StartServerForFactory` call will take in an instance of your bot,
connect to the GobotGame server, authenticate your bot, and play matches as
they're requested through the website. It'll automatically attempt to reconnect
to the server if the connection is lost, and will return error messages if the
AccessToken is invalid or there's something wrong with the bot. To disconnect
your bot from the server, type Ctrl-Z from the terminal.

### Defining your bot

A bot is anything that implements the [game.AI
interface](https://godoc.org/github.com/bcspragu/Gobots/game#AI). If you make a
fancy bot that can keep memory about the history about the current match, you
might want to implement the
[Factory](https://godoc.org/github.com/bcspragu/Gobots/game#Factory) function
on your own, which will be called every time a new game is created, and passed
the game ID.

All of the connecting to the server is handled by `game.StartServerForFactory`,
which takes three parameters.

"MyBot" is the name of your robot. Pick something unique and intimidating.

"MyAccessToken" can be gotten from GobotGame.com once it's online. The rules
are nearly identical to [the rules for RobotGame](https://robotgame.net/rules).

The final field is a Factory wrapping your bot. If you haven't implemented the
factory on your own, you can use the `game.ToFactory` utility method.

## Testing your Bot

The `game` package contains a function
[FightBots](https://godoc.org/github.com/bcspragu/Gobots/game#FightBots) for
fighting two bots against each other and observing the outcome. To actually
view the contents of a match, connect both of the bots to the server and fight
them on there.

## Deploying your Bot

I have a bunch of Google Cloud credits for anyone who wants to serve there bot
on Compute Engine. There's a guide for getting started with Go on Compute
Engine
[here](https://cloud.google.com/go/getting-started/run-on-compute-engine), and
there's also the [cloudlaunch](https://godoc.org/go4.org/cloud/cloudlaunch)
package for easily deploying binaries to Cloud Engine, if you're on a 64-bit
Linux distro.
