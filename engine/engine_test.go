package engine

import (
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

func TestBoard_Set(t *testing.T) {
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

func TestUpdate(t *testing.T) {
	tests := []struct {
		config    BoardConfig
		init      map[Loc]Robot
		initRound int

		// TODO: parameters for round

		want         map[Loc]Robot
		wantRound    int
		turnListFunc func(*capnp.Segment) (botapi.Turn_List, botapi.Turn_List)
	}{
		// TODO: This is no-op update change detector test.
		{
			config: BoardConfig{
				Size: Loc{X: 5, Y: 5},
			},
			init: map[Loc]Robot{
				Loc{1, 1}: Robot{ID: 123, Health: 10, Faction: 0},
				Loc{2, 2}: Robot{ID: 456, Health: 10, Faction: 1},
			},
			initRound: 0,
			want: map[Loc]Robot{
				Loc{1, 1}: Robot{ID: 123, Health: 10, Faction: 0},
				Loc{2, 2}: Robot{ID: 456, Health: 10, Faction: 1},
			},
			wantRound: 1,
			turnListFunc: func(s *capnp.Segment) (botapi.Turn_List, botapi.Turn_List) {
				l, err := botapi.NewTurn_List(s, 0)
				if err != nil {
					t.Fatal("botapi.NewTurn_List:", err)
				}
				return l, l
			},
		},
	}
	for i, test := range tests {
		t.Logf("tests[%d], size = %v, round = %d", i, test.config, test.initRound)
		b := EmptyBoard(test.config)
		b.Round = test.initRound
		for l, r := range test.init {
			t.Logf("  -> set %v to %#v", l, r)
			rr := new(Robot)
			b.addBot(rr, l)
		}

		_, s, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			t.Fatal("capnp.NewMessage:", err)
		}
		ta, tb := test.turnListFunc(s)
		b.Update(ta, tb)

		if b.Round != test.wantRound {
			t.Errorf("  !! b.Round = %d; want %d", b.Round, test.wantRound)
		}

		for y := 0; y < test.config.Size.Y; y++ {
			for x := 0; x < test.config.Size.X; x++ {
				loc := Loc{x, y}
				r := b.At(loc)
				want, ok := test.want[loc]
				if (r != nil) != ok {
					if ok {
						t.Errorf("  !! b.At(%v) = nil; want %#v", loc, want)
					} else {
						t.Errorf("  !! b.At(%v) = %#v; want nil", loc, r)
					}
					continue
				}
				if !ok {
					continue
				}
				if *r != want {
					t.Errorf("  !! b.At(%v) = %#v; want %#v", loc, r, want)
				}
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
