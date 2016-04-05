package main

import (
	"math/rand"
	"time"

	"github.com/bcspragu/Gobots/game"
)

type random struct{}

func (random) Act(b *game.Board, r *game.Robot) game.Action {
	ds := []game.Direction{
		game.North,
		game.South,
		game.East,
		game.West,
	}

	as := []game.ActionKind{
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Move,
		game.Attack,
		game.Guard,
	}

	var a game.Action
	rand.Seed(time.Now().UnixNano())
	aa, bb := rand.Intn(len(as)), rand.Intn(len(ds))
	ak := as[aa]
	a.Kind = ak
	if ak == game.Move || ak == game.Attack {
		a.Direction = ds[bb]
	}
	return a
}
