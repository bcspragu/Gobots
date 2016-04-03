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
	game.StartServerForBot("MyBot", "MyAccessToken", bot{})
}
```

"MyAccessToken" can be gotten from GobotGame.com once it's online. The rules
are nearly identical to [the rules for RobotGame](https://robotgame.net/rules).

## Testing your Bot

Currently, the only way to test your bot is by running a local version of the
server and fighting it against other bots there, though the game package will
include functions for running many matches of a bot against another bot soon.

## Deploying your Bot

If you're building this as part of the workshop at UMass, you can use the
provided script for deploying your bot to Google Cloud Platform.
