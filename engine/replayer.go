package engine

import (
	"github.com/bcspragu/Gobots/botapi"
)

type Playback struct {
	replay botapi.Replay
}

func NewPlayback(r botapi.Replay) *Playback {
	return &Playback{r}
}

func (p *Playback) Board(round int) (*Board, error) {
	// TODO: Stop ignoring errors
	if round == 0 {
		w, err := p.replay.InitialBoard()
		if err != nil {
			return nil, err
		}
		b, err := boardFromWire(w)
		if err != nil {
			return nil, err
		}
		return b, nil
	} else {
		rs, err := p.replay.Rounds()
		if err != nil {
			return nil, err
		}
		w, err := rs.At(round - 1).EndBoard()
		if err != nil {
			return nil, err
		}
		b, err := boardFromWire(w)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}

// boardFromWire converts the wire representation to the board
func boardFromWire(wire botapi.Board) (*Board, error) {
	b := EmptyBoard(int(wire.Width()), int(wire.Height()))
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
