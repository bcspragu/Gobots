package game

import "math"

// Distance returns the Manhattan distance between two locations.
func Distance(a, b Loc) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// LineDistance returns the straight line distance between two locations.
func LineDistance(a, b Loc) float64 {
	dx := math.Pow(float64(b.X-a.X), 2)
	dy := math.Pow(float64(b.Y-a.Y), 2)
	return math.Sqrt(dx + dy)
}

// Find finds a robot on the board that matches the given function.
func (b *Board) Find(f func(*Robot) bool) *Robot {
	for _, row := range b.Cells {
		for _, r := range row {
			if r != nil && f(r) {
				return r
			}
		}
	}
	return nil
}

// LocType returns the type of cell at location loc. It is either Invalid,
// Valid, or Spawn
func (b *Board) LocType(loc Loc) LocType {
	if loc.X >= 0 && loc.Y >= 0 && loc.X < b.Size.X && loc.Y < b.Size.Y {
		return b.LType[loc.X][loc.Y]
	}
	return Invalid
}

// LocsAround returns the locations surrounding the given location, as long as
// they're inside the game board
func (b *Board) LocsAround(loc Loc) []Loc {
	x, y := loc.X, loc.Y
	surrounding := []Loc{
		{x - 1, y},
		{x + 1, y},
		{x, y - 1},
		{x, y + 1},
	}

	var locs []Loc
	for _, l := range surrounding {
		if b.IsInside(l) {
			locs = append(locs, l)
		}
	}
	return locs
}

// Towards takes a current location and a destination, and returns the
// direction to travel to reach it
func Towards(curr, dest Loc) Direction {
	if curr == dest {
		return None
	}

	xD, yD := dest.X-curr.X, dest.Y-curr.Y

	if abs(xD) > abs(yD) {
		if xD > 0 {
			return East
		} else {
			return West
		}
	} else {
		if yD > 0 {
			return South
		} else {
			return North
		}
	}
}

// Center returns the center point of the board. If the board has an even
// dimension, Center() will return the left/top-more location.
func (b *Board) Center() Loc {
	return Loc{
		X: b.Size.X / 2,
		Y: b.Size.Y / 2,
	}
}

// Bots returns a slice of all the bots of a given faction
func (b *Board) Bots(f Faction) []*Robot {
	var bots []*Robot
	for _, col := range b.Cells {
		for _, bot := range col {
			if bot != nil && bot.Faction == f {
				bots = append(bots, bot)
			}
		}
	}
	return bots
}

// IsInside reports whether loc is inside the board bounds.
func (b *Board) IsInside(loc Loc) bool {
	return loc.X >= 0 && loc.X < b.Size.X && loc.Y >= 0 && loc.Y < b.Size.Y
}

// At returns the robot at a particular cell or nil if none is present.
func (b *Board) At(loc Loc) *Robot {
	return b.Cells[loc.X][loc.Y]
}

// Add returns a the current location moved in the direction provided.
func (loc Loc) Add(d Direction) Loc {
	switch d {
	case North:
		return Loc{X: loc.X, Y: loc.Y - 1}
	case South:
		return Loc{X: loc.X, Y: loc.Y + 1}
	case West:
		return Loc{X: loc.X - 1, Y: loc.Y}
	case East:
		return Loc{X: loc.X + 1, Y: loc.Y}
	default:
		return loc
	}
}

func ToFactory(ai AI) Factory {
	return func(gameID string) AI {
		return ai
	}
}

func abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}
