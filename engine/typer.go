package engine

import "math"

type Typer interface {
	// Return the type of cell those x, y coordinates should be
	Type(x, y int) CellType
}

type baseCircle struct {
	size Loc
	c    [][]CellType
}

// circleSpawn has a ring of valid spawn points
type circleSpawn struct{ *baseCircle }

// lineSpawn has a vertical line of spawns on either side
type lineSpawn struct{ *baseCircle }

func newCircleSpawn(size Loc) *circleSpawn {
	cs := &circleSpawn{newBaseCircle(size)}

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			if cs.shouldBeSpawn(x, y) {
				cs.c[x][y] = Spawn
			}
		}
	}

	return cs
}

func (c *circleSpawn) shouldBeSpawn(x, y int) bool {
	if c.c[x][y] != Valid {
		return false
	}

	surrounding := []struct{ x, y int }{
		{x - 1, y},
		{x + 1, y},
		{x, y - 1},
		{x, y + 1},
	}
	for _, s := range surrounding {
		if s.x < 0 || s.x >= c.size.X || s.y < 0 || s.y >= c.size.Y {
			return true
		}
		if c.c[s.x][s.y] == Invalid {
			return true
		}
	}
	return false
}

func NewLineSpawn(size Loc) *lineSpawn {
	ls := &lineSpawn{newBaseCircle(size)}

	for y := 0; y < size.Y; y++ {
		if ls.c[0][y] == Valid {
			ls.c[0][y] = Spawn
			ls.c[size.X-1][y] = Spawn
		}
	}
	return ls
}

func newBaseCircle(size Loc) *baseCircle {
	ct := &baseCircle{
		c:    make([][]CellType, size.X),
		size: size,
	}
	for i := 0; i < size.X; i++ {
		ct.c[i] = make([]CellType, size.Y)
	}

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			ct.c[x][y] = ct.firstPass(x, y)
		}
	}

	return ct
}

func (c *baseCircle) firstPass(x, y int) CellType {
	x += 1
	y += 1
	dx, dy := absFloat(float64(c.size.X+1)/2.0-float64(x)), absFloat(float64(c.size.Y+1)/2.0-float64(y))
	d := math.Sqrt(dx*dx + dy*dy)
	if d < float64(c.size.X+1)/2.0 {
		return Valid
	} else {
		return Invalid
	}
}

func (c *baseCircle) Type(x, y int) CellType {
	return c.c[x][y]
}
