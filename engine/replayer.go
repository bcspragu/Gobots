package engine

import (
	"github.com/bcspragu/Gobots/botapi"
	"github.com/gopherjs/gopherjs/js"
)

type Playback struct {
	Boards []*Board
}

func (p *Playback) Board(i int) *js.Object {
	return js.MakeWrapper(p.Boards[i])
}

func (p *Playback) NumBoards() int {
	return len(p.Boards)
}

func (p *Playback) BoardsJS() []*Board {
	return p.Boards
}

func NewPlayback(r botapi.Replay) (*Playback, error) {
	if bs, err := boards(r); err == nil {
		return &Playback{
			Boards: bs,
		}, nil
	} else {
		return nil, err
	}
}

func boards(replay botapi.Replay) ([]*Board, error) {
	var cells [][]CellType
	w, err := replay.Initial()
	if err != nil {
		return nil, err
	}
	ib, err := boardFromWireWithInitial(w)
	if err != nil {
		return nil, err
	}
	cells = ib.Cells

	// After 0
	rs, err := replay.Rounds()
	if err != nil {
		return nil, err
	}

	bs := make([]*Board, rs.Len()+1)
	bs[0] = ib
	for i := 0; i < rs.Len(); i++ {
		w, err := rs.At(i).EndBoard()
		if err != nil {
			return nil, err
		}
		b, err := boardFromWire(w)
		if err != nil {
			return nil, err
		}
		b.Cells = cells
		bs[i+1] = b
	}
	return bs, nil
}

// boardFromWire converts the wire representation to the board
func boardFromWire(wire botapi.Board) (*Board, error) {
	b := EmptyBoard(BoardConfig{
		Size: Loc{X: int(wire.Width()), Y: int(wire.Height())},
	})
	b.Round = int(wire.Round())

	bots, err := wire.Robots()
	if err != nil {
		return b, err
	}

	for i := 0; i < bots.Len(); i++ {
		bot := bots.At(i)
		loc := Loc{
			X: int(bot.X()),
			Y: int(bot.Y()),
		}
		b.Locs[loc] = robotFromWire(bot)
	}

	return b, nil
}

// boardFromWireWithInitial converts the wire representation to the board
func boardFromWireWithInitial(wire botapi.InitialBoard) (*Board, error) {
	wb, err := wire.Board()
	if err != nil {
		return new(Board), err
	}

	w, h := int(wb.Width()), int(wb.Height())
	b := EmptyBoard(BoardConfig{
		Size: Loc{X: w, Y: h},
	})
	b.Round = int(wb.Round())

	bots, err := wb.Robots()
	if err != nil {
		return b, err
	}

	for i := 0; i < bots.Len(); i++ {
		bot := bots.At(i)
		loc := Loc{
			X: int(bot.X()),
			Y: int(bot.Y()),
		}
		b.Locs[loc] = robotFromWire(bot)
	}

	cells, err := wire.Cells()
	if err != nil {
		return b, err
	}

	for i := 0; i < cells.Len(); i++ {
		cell := cells.At(i)
		x, y := i%w, i/w
		b.Cells[x][y] = cellFromWire[cell]
	}

	return b, nil
}

func robotFromWire(wire botapi.Robot) *Robot {
	var faction int
	if wire.Faction() == botapi.Faction_mine {
		faction = P1Faction
	} else {
		faction = P2Faction
	}

	return &Robot{
		ID:      RobotID(wire.Id()),
		Health:  int(wire.Health()),
		Faction: faction,
	}
}
