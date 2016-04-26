package engine

import (
	"strconv"
	"strings"
	"testing"

	"github.com/bcspragu/Gobots/botapi"
	"zombiezen.com/go/capnproto2"
)

var testConfig = BoardConfig{
	Size: Loc{X: 3, Y: 5},
}

func TestEmptyBoardIsEmpty(t *testing.T) {
	b := EmptyBoard(testConfig)
	if b.Width() != 3 {
		t.Errorf("b.Width() = %d; want 3", b.Width())
	}
	if b.Height() != 5 {
		t.Errorf("b.Height() = %d; want 5", b.Height())
	}
	for y := 0; y < 5; y++ {
		for x := 0; x < 3; x++ {
			loc := Loc{x, y}
			if r := b.At(loc); r != nil {
				t.Errorf("b.At(%v) = %#v; want nil", loc, r)
			}
		}
	}
}

func TestBoardAddBot(t *testing.T) {
	b := EmptyBoard(testConfig)
	loc := Loc{1, 2}
	b.addBot(&Robot{
		ID:      1234,
		Health:  50,
		Faction: 3,
	}, loc)
	if r := b.At(loc); r != nil {
		if r.ID != 1234 {
			t.Errorf("b.At(%v).ID = %d; want 1234", loc, r.ID)
		}
		if r.Health != 50 {
			t.Errorf("b.At(%v).Health = %d; want 50", loc, r.Health)
		}
		if r.Faction != 3 {
			t.Errorf("b.At(%v).Faction = %d; want 3", loc, r.Faction)
		}
	} else {
		t.Errorf("b.At(%v) = nil", r)
	}
}

func TestBoardUpdate(t *testing.T) {
	tests := []struct {
		name      string
		config    BoardConfig
		init      map[Loc]*Robot
		initRound int

		want      map[Loc]*Robot
		wantRound int
		turnList1 string
		turnList2 string
	}{
		{
			name: "This is no-op update change detector test case",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 10, Faction: 0},
				Loc{2, 2}: &Robot{ID: 456, Health: 10, Faction: 1},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 10, Faction: 0},
				Loc{2, 2}: &Robot{ID: 456, Health: 10, Faction: 1},
			},
			wantRound: 1,
			turnList1: "123:W",
			turnList2: "456:W",
		},
		{
			name: "This checks a simple collision",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{0, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{0, 1}: &Robot{ID: 123, Health: 15, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 15, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:ME",
			turnList2: "456:MW",
		},
		{
			name: "This checks bots trying to go out of bounds. They shouldn't move, and they should lose health",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{0, 0}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{4, 4}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{0, 0}: &Robot{ID: 123, Health: 15, Faction: P1Faction},
				Loc{4, 4}: &Robot{ID: 456, Health: 15, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:MW",
			turnList2: "456:MS",
		},
		{
			name: "This checks a complicated dependency chain that needs to be unravelled",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 1, Health: 20, Faction: P1Faction},
				Loc{3, 1}: &Robot{ID: 2, Health: 20, Faction: P1Faction},
				Loc{2, 0}: &Robot{ID: 3, Health: 20, Faction: P1Faction},
				Loc{1, 2}: &Robot{ID: 4, Health: 20, Faction: P1Faction},
				Loc{0, 2}: &Robot{ID: 5, Health: 20, Faction: P1Faction},
				Loc{0, 1}: &Robot{ID: 6, Health: 20, Faction: P1Faction},
				Loc{3, 2}: &Robot{ID: 7, Health: 20, Faction: P1Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 1, Health: 15, Faction: P1Faction},
				Loc{3, 1}: &Robot{ID: 2, Health: 15, Faction: P1Faction},
				Loc{2, 0}: &Robot{ID: 3, Health: 15, Faction: P1Faction},
				Loc{1, 2}: &Robot{ID: 4, Health: 15, Faction: P1Faction},
				Loc{0, 2}: &Robot{ID: 5, Health: 15, Faction: P1Faction},
				Loc{0, 1}: &Robot{ID: 6, Health: 15, Faction: P1Faction},
				Loc{4, 2}: &Robot{ID: 7, Health: 20, Faction: P1Faction},
			},
			wantRound: 1,
			turnList1: "1:ME,2:MW,3:MS,4:MN,5:ME,6:MS,7:ME",
			turnList2: "",
		},
		{
			name: "This checks a conga line of bots around the outside the map windmilling",
			config: BoardConfig{
				Size: Loc{X: 4, Y: 4},
			},
			init: map[Loc]*Robot{
				Loc{0, 0}: &Robot{ID: 1, Health: 20, Faction: P1Faction},
				Loc{0, 1}: &Robot{ID: 2, Health: 20, Faction: P1Faction},
				Loc{0, 2}: &Robot{ID: 3, Health: 20, Faction: P1Faction},
				Loc{0, 3}: &Robot{ID: 4, Health: 20, Faction: P1Faction},
				Loc{1, 3}: &Robot{ID: 5, Health: 20, Faction: P1Faction},
				Loc{2, 3}: &Robot{ID: 6, Health: 20, Faction: P1Faction},
				Loc{3, 3}: &Robot{ID: 7, Health: 20, Faction: P1Faction},
				Loc{3, 2}: &Robot{ID: 8, Health: 20, Faction: P1Faction},
				Loc{3, 1}: &Robot{ID: 9, Health: 20, Faction: P1Faction},
				Loc{3, 0}: &Robot{ID: 10, Health: 20, Faction: P1Faction},
				Loc{2, 0}: &Robot{ID: 11, Health: 20, Faction: P1Faction},
				Loc{1, 0}: &Robot{ID: 12, Health: 20, Faction: P1Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{0, 1}: &Robot{ID: 1, Health: 20, Faction: P1Faction},
				Loc{0, 2}: &Robot{ID: 2, Health: 20, Faction: P1Faction},
				Loc{0, 3}: &Robot{ID: 3, Health: 20, Faction: P1Faction},
				Loc{1, 3}: &Robot{ID: 4, Health: 20, Faction: P1Faction},
				Loc{2, 3}: &Robot{ID: 5, Health: 20, Faction: P1Faction},
				Loc{3, 3}: &Robot{ID: 6, Health: 20, Faction: P1Faction},
				Loc{3, 2}: &Robot{ID: 7, Health: 20, Faction: P1Faction},
				Loc{3, 1}: &Robot{ID: 8, Health: 20, Faction: P1Faction},
				Loc{3, 0}: &Robot{ID: 9, Health: 20, Faction: P1Faction},
				Loc{2, 0}: &Robot{ID: 10, Health: 20, Faction: P1Faction},
				Loc{1, 0}: &Robot{ID: 11, Health: 20, Faction: P1Faction},
				Loc{0, 0}: &Robot{ID: 12, Health: 20, Faction: P1Faction},
			},
			wantRound: 1,
			turnList1: "1:MS,2:MS,3:MS,4:ME,5:ME,6:ME,7:MN,8:MN,9:MN,10:MW,11:MW,12:MW",
			turnList2: "",
		},
		{
			name: "This checks four bots moving in a windmill, none should collide",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{0, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{1, 0}: &Robot{ID: 124, Health: 20, Faction: P1Faction},
				Loc{1, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
				Loc{0, 0}: &Robot{ID: 457, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{0, 0}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{1, 1}: &Robot{ID: 124, Health: 20, Faction: P1Faction},
				Loc{0, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
				Loc{1, 0}: &Robot{ID: 457, Health: 20, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:MN,124:MS",
			turnList2: "456:MW,457:ME",
		},
		{
			name: "This checks two bots trying to swap places, which isn't allowed. They collide instead.",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 15, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 15, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:ME",
			turnList2: "456:MW",
		},
		{
			name: "This checks basic attacking",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 10, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:W",
			turnList2: "456:AW",
		},
		{
			name: "This checks guarding from an attack",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 15, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:G",
			turnList2: "456:AW",
		},
		{
			name: "This checks guarding from a self-destruct",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 13, Faction: P1Faction},
			},
			wantRound: 1,
			turnList1: "123:G",
			turnList2: "456:D",
		},
		{
			name: "This checks guarding from a collision",
			config: BoardConfig{
				Size:      Loc{X: 5, Y: 5},
				CellTyper: allValid{},
				Spawner:   noSpawn{},
			},
			init: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{1, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				Loc{2, 1}: &Robot{ID: 456, Health: 15, Faction: P2Faction},
			},
			wantRound: 1,
			turnList1: "123:G",
			turnList2: "456:MW",
		},
		{
			name: "This checks self-destructing",
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]*Robot{
				Loc{4, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction}, // Out of blast radius
				Loc{2, 1}: &Robot{ID: 456, Health: 20, Faction: P2Faction}, // Exploder
				// In blast radius
				Loc{1, 0}: &Robot{ID: 124, Health: 20, Faction: P1Faction},
				Loc{2, 0}: &Robot{ID: 125, Health: 20, Faction: P1Faction},
				Loc{3, 0}: &Robot{ID: 126, Health: 20, Faction: P1Faction},
				Loc{1, 1}: &Robot{ID: 127, Health: 20, Faction: P1Faction},
				Loc{3, 1}: &Robot{ID: 128, Health: 20, Faction: P1Faction},
				Loc{1, 2}: &Robot{ID: 129, Health: 20, Faction: P1Faction},
				Loc{2, 2}: &Robot{ID: 130, Health: 20, Faction: P1Faction},
				Loc{3, 2}: &Robot{ID: 131, Health: 20, Faction: P1Faction},
			},
			initRound: 0,
			want: map[Loc]*Robot{
				Loc{4, 1}: &Robot{ID: 123, Health: 20, Faction: P1Faction},
				// In blast radius
				Loc{1, 0}: &Robot{ID: 124, Health: 5, Faction: P1Faction},
				Loc{2, 0}: &Robot{ID: 125, Health: 5, Faction: P1Faction},
				Loc{3, 0}: &Robot{ID: 126, Health: 5, Faction: P1Faction},
				Loc{1, 1}: &Robot{ID: 127, Health: 5, Faction: P1Faction},
				Loc{3, 1}: &Robot{ID: 128, Health: 5, Faction: P1Faction},
				Loc{1, 2}: &Robot{ID: 129, Health: 5, Faction: P1Faction},
				Loc{2, 2}: &Robot{ID: 130, Health: 5, Faction: P1Faction},
				Loc{3, 2}: &Robot{ID: 131, Health: 5, Faction: P1Faction},
			},
			wantRound: 1,
			turnList1: "123:W,124:W,125:W,126:W,127:W,128:W,129:W,130:W,131:W",
			turnList2: "456:D",
		},
	}
	for _, test := range tests {
		if test.config.CellTyper == nil {
			test.config.CellTyper = allValid{}
		}
		if test.config.Spawner == nil {
			test.config.Spawner = noSpawn{}
		}
		b := EmptyBoard(test.config)
		b.InitBoard()
		b.Round = test.initRound
		for l, r := range test.init {
			b.addBot(r, l)
		}

		ta, err := turnList(test.turnList1)
		if err != nil {
			t.Fatal("turnList:", err)
		}
		tb, err := turnList(test.turnList2)
		if err != nil {
			t.Fatal("turnList:", err)
		}
		b.Update(ta, tb)

		if b.Round != test.wantRound {
			t.Errorf("b.Round = %d; want %d", b.Round, test.wantRound)
		}

		var displayedDescription bool
		for l, bot := range test.want {
			r := b.At(l)
			if *r != *bot {
				if !displayedDescription {
					t.Errorf("Failed on test case: %s", test.name)
					displayedDescription = true
				}
				t.Errorf("b.At(%v) = %#v; want %#v", l, *r, *bot)
			}
		}
	}
}

func TestToWire(t *testing.T) {
	b := EmptyBoard(BoardConfig{Size: Loc{X: 4, Y: 6}})
	b.addBot(&Robot{ID: 254, Health: 50, Faction: 0}, Loc{1, 2})
	b.addBot(&Robot{ID: 973, Health: 12, Faction: 1}, Loc{3, 4})
	b.Round = 42

	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		t.Fatal("capnp.NewMessage:", err)
	}
	wb, err := botapi.NewRootBoard(seg)
	if err != nil {
		t.Fatal("botapi.NewRootBoard:", err)
	}

	err = b.ToWire(wb, 1)
	if err != nil {
		t.Fatal("b.ToWire:", err)
	}

	if wb.Width() != 4 {
		t.Errorf("width = %d; want 4", wb.Width())
	}
	if wb.Height() != 6 {
		t.Errorf("height = %d; want 6", wb.Height())
	}
	if wb.Round() != 42 {
		t.Errorf("round = %d; want 42", wb.Round())
	}
	if robots, err := wb.Robots(); err == nil {
		if robots.Len() == 2 {
			if robots.At(0).Id() != 254 {
				t.Errorf("robots[0].id = %d; want 254", robots.At(0).Id())
			}
			if robots.At(0).X() != 1 || robots.At(0).Y() != 2 {
				t.Errorf("robots[0].x,y = %d,%d; want 1,2", robots.At(0).X(), robots.At(0).Y())
			}
			if robots.At(0).Health() != 50 {
				t.Errorf("robots[0].health = %d; want 50", robots.At(0).Health())
			}
			if robots.At(0).Faction() != botapi.Faction_opponent {
				t.Errorf("robots[0].faction = %v; want opponent", robots.At(0).Faction())
			}

			if robots.At(1).Id() != 973 {
				t.Errorf("robots[1].id = %d; want 973", robots.At(1).Id())
			}
			if robots.At(1).X() != 3 || robots.At(1).Y() != 4 {
				t.Errorf("robots[1].x,y = %d,%d; want 3,4", robots.At(1).X(), robots.At(1).Y())
			}
			if robots.At(1).Health() != 12 {
				t.Errorf("robots[1].health = %d; want 12", robots.At(1).Health())
			}
			if robots.At(1).Faction() != botapi.Faction_mine {
				t.Errorf("robots[1].faction = %v; want mine", robots.At(1).Faction())
			}
		} else {
			t.Errorf("len(robots) = %d; want 2", robots.Len())
		}
	} else {
		t.Errorf("robots error: %v", err)
	}
}

var dMap = map[byte]botapi.Direction{
	'N': botapi.Direction_north,
	'S': botapi.Direction_south,
	'E': botapi.Direction_east,
	'W': botapi.Direction_west,
}

func turnList(t string) (botapi.Turn_List, error) {
	_, s, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return botapi.Turn_List{}, err
	}

	moves := strings.Split(t, ",")
	if t == "" {
		moves = []string{}
	}

	l, err := botapi.NewTurn_List(s, int32(len(moves)))
	if err != nil {
		return l, err
	}
	for i, m := range moves {
		t := l.At(i)
		p := strings.Split(m, ":")
		id, err := strconv.Atoi(p[0])
		if err != nil {
			return l, err
		}
		t.SetId(uint32(id))
		move := p[1]
		switch move[0] {
		// Wait
		case 'W':
			t.SetWait()
		// Move
		case 'M':
			t.SetMove(dMap[move[1]])
		// Attack
		case 'A':
			t.SetAttack(dMap[move[1]])
		// Destruct
		case 'D':
			t.SetSelfDestruct()
		// Guard
		case 'G':
			t.SetGuard()
		}
	}
	return l, nil
}
