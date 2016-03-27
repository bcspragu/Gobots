package main

import "github.com/bcspragu/Gobots/game"

type aggro struct{}

func (aggro) Action(b *game.Board, r *game.Robot) game.Turn {
	ds := []game.Direction{
		game.North,
		game.South,
		game.East,
		game.West,
	}
	for _, d := range ds {
		loc := r.Loc.Add(d)
		if opponentAt(b, loc) {
			return game.Turn{
				Kind:      game.Attack,
				Direction: d,
			}
		}
	}
	return game.Turn{Kind: game.Wait}
}

func opponentAt(b *game.Board, loc game.Loc) bool {
	if !b.IsInside(loc) {
		return false
	}
	r := b.At(loc)
	if r == nil {
		return false
	}
	return r.Faction == game.OpponentFaction
}
