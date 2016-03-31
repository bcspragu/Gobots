package main

import "github.com/bcspragu/Gobots/game"

type pathfinder struct {
	targets map[uint32]uint32
}

func (pf *pathfinder) Act(b *game.Board, r *game.Robot) game.Action {
	// Immediate surrounding attacks
	ds := []game.Direction{
		game.North,
		game.South,
		game.East,
		game.West,
	}
	for _, d := range ds {
		loc := r.Loc.Add(d)
		if pf.opponentAt(b, loc) {
			return game.Action{
				Kind:      game.Attack,
				Direction: d,
			}
		}
	}

	// Acquire target
	tgt, ok := pf.targets[r.ID]
	var opp *game.Robot
	if ok {
		opp = b.Find(func(q *game.Robot) bool {
			return q.ID == tgt
		})
	}
	if !ok || opp == nil {
		if pf.targets == nil {
			pf.targets = make(map[uint32]uint32)
		}
		opp = nearestOpponent(b, r.Loc)
		if opp == nil {
			return game.Action{Kind: game.Wait}
		}
		pf.targets[r.ID] = opp.ID
	}

	// Move to target.
	// Don't worry about collisions, since we already shot at all neighbors.
	// TODO: but what about friends?
	// TODO: and why don't you compute the vector angle?
	switch {
	case opp.Loc.X < r.Loc.X:
		return game.Action{
			Kind:      game.Move,
			Direction: game.West,
		}
	case opp.Loc.X > r.Loc.X:
		return game.Action{
			Kind:      game.Move,
			Direction: game.East,
		}
	case opp.Loc.Y < r.Loc.Y:
		return game.Action{
			Kind:      game.Move,
			Direction: game.North,
		}
	case opp.Loc.Y > r.Loc.Y:
		return game.Action{
			Kind:      game.Move,
			Direction: game.South,
		}
	}
	// TODO: impossibru?
	return game.Action{Kind: game.Wait}
}

func nearestOpponent(b *game.Board, loc game.Loc) *game.Robot {
	// Probably faster ways of doing this.. traversing outward
	var closest *game.Robot
	var closestDist int
	for y, row := range b.Cells {
		for x, r := range row {
			curr := game.Loc{x, y}
			if r == nil || r.Faction != game.OpponentFaction {
				continue
			}
			d := game.Distance(loc, curr)
			if closest == nil || d < closestDist {
				closest, closestDist = r, d
			}
		}
	}
	return closest
}

func (pf *pathfinder) opponentAt(b *game.Board, loc game.Loc) bool {
	if !b.IsInside(loc) {
		return false
	}
	r := b.At(loc)
	if r == nil {
		return false
	}
	return r.Faction == game.OpponentFaction
}
