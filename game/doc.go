package game

/*
Package game is used for developing robot AIs to compete on GobotGame.com. An
example AI is included below:

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

The winner of a game is determined by who has more robots on the board after
100 rounds, with robots spawning at the left and right edges of the board every
10 turns. The rules are nearly identical to those of RobotGame, available at
https://robotgame.net/rules.
*/
