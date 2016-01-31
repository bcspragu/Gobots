package engine

import (
	"math"
)

func (b *Board) cellType(x, y int) CellType {
	dx, dy := abs(b.Size.X/2.0-x), abs(b.Size.Y/2.0-y)
	d := math.Sqrt(dx*dx + dy*dy)
	if d < dx {
		return Valid
	} else {
		return Invalid
	}
}
