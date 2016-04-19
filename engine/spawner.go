package engine

import "math/rand"

type Spawner interface {
	// Given a list of possible spawn locations for the first player, return a
	// list of locations to spawn players. The program will automatically
	// mirror the spawn locations for the other faction.
	Spawn(locs []Loc) []Loc
}

type SpawnType int

var (
	AllSpawn        Spawner = allSpawn{}
	EveryOtherSpawn Spawner = everyOtherSpawn{}
)

type noSpawn struct{}

func (noSpawn) Spawn(locs []Loc) []Loc { return []Loc{} }

type allSpawn struct{}

func (allSpawn) Spawn(locs []Loc) []Loc { return locs }

type everyOtherSpawn struct{}

func (everyOtherSpawn) Spawn(locs []Loc) (res []Loc) {
	i := 0
	for _, loc := range locs {
		if i%2 == 0 {
			res = append(res, loc)
		}
		i++
	}
	return
}

func NewRandomSpawn(n int) Spawner {
	return &randomSpawn{
		n: n,
	}
}

type randomSpawn struct {
	// 1 in n chance  of spawning
	n int
}

func (r *randomSpawn) Spawn(locs []Loc) (res []Loc) {
	for _, loc := range locs {
		if rand.Intn(r.n) == 0 {
			res = append(res, loc)
		}
	}
	return
}
