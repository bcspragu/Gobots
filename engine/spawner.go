package engine

import "math/rand"

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

type RandomSpawn struct {
	// 1 in n chance  of spawning
	N int
}

func (r *RandomSpawn) Spawn(locs []Loc) (res []Loc) {
	for _, loc := range locs {
		if rand.Intn(r.N) == 0 {
			res = append(res, loc)
		}
	}
	return
}
