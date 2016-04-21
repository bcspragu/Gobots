package engine

import "strconv"

type Faction int

const (
	P1Faction Faction = iota + 1
	P2Faction
)

// A Robot is a single piece on a board.
type Robot struct {
	ID      RobotID
	Health  int
	Faction Faction
}

// ByID implements sort.Interface for a slice of robots
type ByID []*Robot

func (b ByID) Len() int           { return len(b) }
func (b ByID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByID) Less(i, j int) bool { return b[i].ID < b[j].ID }

// RobotID uniquely identifies a robot within a game
type RobotID uint32

func (id RobotID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id RobotID) GoString() string {
	return id.String()
}
