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

func (b *Board) LocType(loc Loc) LocType {
	if loc.X >= 0 && loc.Y >= 0 && loc.X < b.Size.X && loc.Y < b.Size.Y {
		return b.LType[loc.X][loc.Y]
	}
	return Invalid
}

// IsInside reports whether loc is inside the board bounds.
func (b *Board) IsInside(loc Loc) bool {
	return loc.X >= 0 && loc.X < len(b.Cells[0]) && loc.Y >= 0 && loc.Y < len(b.Cells)
}

// At returns the robot at a particular cell or nil if none is present.
func (b *Board) At(loc Loc) *Robot {
	return b.Cells[loc.Y][loc.X]
}

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
