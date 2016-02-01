package engine

import (
	"math"
)

// Ignore how weird the math is, I had to fiddle a bit to get the grids to come out how I wanted them
func (b *Board) cellType(x, y int) CellType {
	x += 1
	y += 1
	dx, dy := absFloat(float64(b.Size.X+1)/2.0-float64(x)), absFloat(float64(b.Size.Y+1)/2.0-float64(y))
	d := math.Sqrt(dx*dx + dy*dy)
	if d < float64(b.Size.X+1)/2.0 {
		return Valid
	} else {
		return Invalid
	}
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
