/*
Package game is used for developing robot AIs to compete on GobotGame.com. An
example AI is included below:

		package main

		import "github.com/bcspragu/Gobots/game"

		// Bot moves to the center and does nothing else
		type bot struct{}

		func (bot) Act(b *game.Board, r *game.Robot) game.Action {
			return game.Action{
				Kind:      game.Move,
				Direction: game.Towards(r.Loc, b.Center()),
			}
		}

		func main() {
			game.StartServerForFactory("MyBot", "MyAccessToken", game.ToFactory(bot{}))
		}

The winner of a game is determined by who has more robots on the board after
100 rounds, with robots spawning at the left and right edges of the board every
10 turns. The rules are nearly identical to those of RobotGame, available at
https://robotgame.net/rules.
*/
package game

import "github.com/bcspragu/Gobots/botapi"

// Board represents the state of the board in a round.
type Board struct {
	Round int
	Size  Loc
	Cells [][]*Robot
	LType [][]LocType
}

// A Robot is a piece on the board.
type Robot struct {
	ID      uint32
	Loc     Loc
	Faction Faction
	Health  int
}

// Faction identifies who owns a robot.
type Faction int

const (
	MyFaction Faction = iota + 1
	OpponentFaction
)

// LocType identifies what properties (invalid, valid, spawn) the location has
type LocType int

const (
	Invalid LocType = iota
	Valid
	Spawn
)

// An AI is an algorithm that makes moves for a particular game.
type AI interface {
	Act(board *Board, r *Robot) Action
}

// Loc is a coordinate pair.
type Loc struct {
	X, Y int
}

var locFromWire = map[botapi.CellType]LocType{
	botapi.CellType_invalid: Invalid,
	botapi.CellType_valid:   Valid,
	botapi.CellType_spawn:   Spawn,
}

// A Action represents what a robot will do.  The zero value waits the turn.
type Action struct {
	Kind      ActionKind
	Direction Direction
}

func (a Action) toWire(id uint32, wire botapi.Turn) {
	wire.SetId(id)
	switch a.Kind {
	case Wait:
		wire.SetWait()
	case Move:
		wire.SetMove(a.Direction.toWire())
	case Attack:
		wire.SetAttack(a.Direction.toWire())
	case SelfDestruct:
		wire.SetSelfDestruct()
	case Guard:
		wire.SetGuard()
	}
}

// ActionKind is an enumeration of the kinds of turns.
type ActionKind int

// Kinds of actions.
const (
	Wait ActionKind = iota
	Move
	Attack
	SelfDestruct
	Guard
)

// Direction is a cardinal direction.
type Direction int

// The defined directions.
const (
	None  = Direction(-1)
	North = Direction(botapi.Direction_north)
	South = Direction(botapi.Direction_south)
	East  = Direction(botapi.Direction_east)
	West  = Direction(botapi.Direction_west)
)

func (d Direction) toWire() botapi.Direction {
	return botapi.Direction(d)
}

// Factory is a function that creates an AI per game.
type Factory func(gameID string) AI

type gameState struct {
	ai   AI
	locs [][]LocType
}

// aiAdapter is a type that implements botapi.Ai by mapping turns to
// games and calling the AI interface methods.
//
// Note: since TakeTurn does not call server.Ack, the Cap'n Proto
// concurrency model guarantees that each call to TakeTurn happens after
// the previous return. Thus, we don't need to add any additional locks.
type aiAdapter struct {
	factory Factory
	games   map[string]gameState
}

func (a *aiAdapter) TakeTurn(call botapi.Ai_takeTurn) error {
	ib, err := call.Params.Board()
	if err != nil {
		return err
	}
	board, err := ib.Board()
	if err != nil {
		return err
	}
	gameID, err := board.GameId()
	if err != nil {
		return err
	}

	// Convert the board to the game representation
	b, robots, err := convertBoard(board)
	if err != nil {
		return err
	}

	// Load the AI for this game, or create a new one
	ai := a.games[gameID].ai
	if ai == nil {
		ai = a.factory(gameID)

		// Load the cells for the board
		cells, err := ib.Cells()
		if err != nil {
			return nil
		}
		locs := convertLocs(cells, len(b.Cells), len(b.Cells[0]))
		a.games[gameID] = gameState{
			ai:   ai,
			locs: locs,
		}
	}
	b.LType = a.games[gameID].locs
	turns, err := botapi.NewTurn_List(call.Results.Segment(), int32(len(robots)))
	if err != nil {
		return err
	}
	for i, r := range robots {
		t := ai.Act(b, r)
		t.toWire(r.ID, turns.At(i))
	}
	call.Results.SetTurns(turns)
	return nil
}

func convertLocs(wireLocs botapi.CellType_List, w, h int) [][]LocType {
	locs := make([]LocType, w*h)
	cols := make([][]LocType, w)
	for x := range cols {
		cols[x] = locs[x*h : (x+1)*h]
	}
	for i := 0; i < wireLocs.Len(); i++ {
		x, y := i%w, i/w
		cols[x][y] = locFromWire[wireLocs.At(i)]
	}
	return cols
}

func convertBoard(wire botapi.Board) (b *Board, playerBots []*Robot, err error) {
	w, h := int(wire.Width()), int(wire.Height())
	cells := make([]*Robot, w*h)
	cols := make([][]*Robot, w)
	for x := range cols {
		cols[x] = cells[x*h : (x+1)*h]
	}
	robots, err := wire.Robots()
	if err != nil {
		return nil, nil, err
	}
	playerBots = make([]*Robot, 0, robots.Len())
	for i, n := 0, robots.Len(); i < n; i++ {
		r := robots.At(i)
		l := Loc{X: int(r.X()), Y: int(r.Y())}
		rr := &Robot{
			ID:     r.Id(),
			Loc:    l,
			Health: int(r.Health()),
		}
		switch r.Faction() {
		case botapi.Faction_mine:
			rr.Faction = MyFaction
			playerBots = append(playerBots, rr)
		case botapi.Faction_opponent:
			fallthrough
		default:
			rr.Faction = OpponentFaction
		}
		cols[l.X][l.Y] = rr
	}
	return &Board{
		Size:  Loc{w, h},
		Round: int(wire.Round()),
		Cells: cols,
	}, playerBots, nil
}

const (
	exitFail  = 1
	exitUsage = 64
)
