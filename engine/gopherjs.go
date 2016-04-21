package engine

// At returns the robot at a location or nil if not found.
func (b *Board) AtXY(x, y int) Cell {
	return b.Cells[x][y]
}
