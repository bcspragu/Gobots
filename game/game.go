// Package game provides an idiomatic Go wrapper around the bot API.
package game

import (
	"log"

	"github.com/bcspragu/Gobots/botapi"
)

// Board represents the state of the board in a round.
type Board struct {
	Round    int
	Cells    [][]*Robot
	loctypes [][]LocType
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
	InvalidLoc LocType = iota
	ValidLoc
	SpawnLoc
)

// An AI is an algorithm that makes moves for a particular game.
type AI interface {
	Act(board *Board, r *Robot) Action
}

// Loc is a coordinate pair.
type Loc struct {
	X, Y int
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

// aiAdapter is a type that implements botapi.Ai by mapping turns to
// games and calling the AI interface methods.
//
// Note: since TakeTurn does not call server.Ack, the Cap'n Proto
// concurrency model guarantees that each call to TakeTurn happens after
// the previous return. Thus, we don't need to add any additional locks.
type aiAdapter struct {
	factory Factory
	games   map[string]AI
}

func (a *aiAdapter) TakeTurn(call botapi.Ai_takeTurn) error {
	board, err := call.Params.Board()
	if err != nil {
		return err
	}
	gameID, err := board.GameId()
	if err != nil {
		return err
	}
	ai := a.games[gameID]
	if ai == nil {
		ai = a.factory(gameID)
		a.games[gameID] = ai
	}

	b, robots, err := convertBoard(board)
	if err != nil {
		return err
	}
	turns, err := botapi.NewTurn_List(call.Results.Segment(), int32(len(robots)))
	if err != nil {
		return err
	}
	for i, r := range robots {
		t := ai.Act(b, r)
		log.Printf("Robot %v making move %v", r, t)
		t.toWire(r.ID, turns.At(i))
	}
	call.Results.SetTurns(turns)
	return nil
}

func convertBoard(wire botapi.Board) (b *Board, playerBots []*Robot, err error) {
	w, h := int(wire.Width()), int(wire.Height())
	cells := make([]*Robot, w*h)
	rows := make([][]*Robot, h)
	for y := range rows {
		rows[y] = cells[y*w : (y+1)*w]
	}
	robots, err := wire.Robots()
	if err != nil {
		return nil, nil, err
	}
	playerBots = make([]*Robot, 0, robots.Len())
	for i, n := 0, robots.Len(); i < n; i++ {
		r := robots.At(i)
		// TODO(light): check for negative (x,y)
		rr := &Robot{
			ID:     r.Id(),
			Loc:    Loc{int(r.X()), int(r.Y())},
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
		log.Printf("Robot #%d: %+v", i, rr)
		rows[rr.Loc.Y][rr.Loc.X] = rr
	}
	return &Board{
		Round: int(wire.Round()),
		Cells: rows,
	}, playerBots, nil
}

const (
	exitFail  = 1
	exitUsage = 64
)
