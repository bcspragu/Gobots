# Gobots

A programmatic robot-fighting game, **heavily** inspired by [Robot
Game](http://robotgame.net). As of the time of this writing, it isn't online
yet, but it will be at GobotGame.com (hopefully) soon.

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

Explaining what this code does:

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

If you're building this as part of the workshop at UMass, you can use the [soon
to be] provided script for deploying your bot to Google Cloud Platform.
