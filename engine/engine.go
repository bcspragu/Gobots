package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/bcspragu/Gobots/botapi"
)

const (
	P1Faction = 1
	P2Faction = 2

	InitialHealth = 5

	CollisionDamage = 1
	AttackDamage    = 2
	DestructDamage  = 2
	SelfDamage      = 1000 // Make them super dead

	NewBotsSpacing = 5 // Number of rounds to wait before spawning new bots
)

type CellInfo struct {
	Bot      *Robot
	CellType CellType
}

type CellType int

type SpawnFunc func(Loc) bool

var all SpawnFunc = func(loc Loc) bool {
	return true
}

var everyOther SpawnFunc = func(loc Loc) bool {
	return loc.Y%2 == 0
}

var randomSpawnFuncGen = func(b *Board) SpawnFunc {
	r := rand.Perm(b.Size.Y)
	return func(loc Loc) bool {
		return r[loc.Y]%2 == 0
	}
}

const (
	UnknownCellType CellType = iota
	Invalid
	Valid
)

type collisionMap map[Loc][]*Robot

type LocPair struct {
	L Loc
	B *Robot
}

type Board struct {
	Locs map[Loc]*Robot

	Cells [][]CellType
	Size  Loc
	Round int

	NextID RobotID
}

func (b *Board) CellsJS() [][]CellType {
	return b.Cells
}

type JSONBoard struct {
	Pairs []LocPair

	Cells [][]CellType
	Size  Loc
	Round int

	NextID RobotID
}

func (b *Board) ToJSONBoard() *JSONBoard {
	j := &JSONBoard{
		Size:   b.Size,
		Round:  b.Round,
		NextID: b.NextID,
		Cells:  b.Cells,
	}

	j.Pairs = make([]LocPair, len(b.Locs))
	i := 0
	for loc, bot := range b.Locs {
		j.Pairs[i] = LocPair{
			L: loc,
			B: bot,
		}
		i++
	}
	return j
}

func (j *JSONBoard) ToBoard() *Board {
	b := &Board{
		Locs:   make(map[Loc]*Robot),
		Size:   j.Size,
		Round:  j.Round,
		NextID: j.NextID,
		Cells:  j.Cells,
	}

	for _, pair := range j.Pairs {
		b.Locs[pair.L] = pair.B
	}
	return b
}

// EmptyBoard creates an empty board of the given size.
func EmptyBoard(w, h int) *Board {
	b := &Board{
		Locs:  make(map[Loc]*Robot),
		Size:  Loc{w, h},
		Cells: make([][]CellType, w),
	}

	for i := 0; i < w; i++ {
		b.Cells[i] = make([]CellType, h)
	}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			b.Cells[x][y] = b.cellType(x, y)
		}
	}

	return b
}

func (b *Board) Width() int {
	return b.Size.X
}

func (b *Board) Height() int {
	return b.Size.Y
}

func (b *Board) newID() RobotID {
	b.NextID++
	return b.NextID
}

// NewBoard creates an initialized game board for two factions.
func NewBoard(w, h int) *Board {
	b := EmptyBoard(w, h)
	b.SpawnBots(everyOther)
	return b
}

func (b *Board) SpawnBots(spawnFunc SpawnFunc) {
	// Just line the ends with robots
	for i := 0; i < b.Size.Y; i++ {
		la, lb := Loc{0, i}, Loc{b.Size.X - 1, i}
		if b.isValidLoc(la) && spawnFunc(la) {
			b.Locs[la] = &Robot{
				ID:      b.newID(),
				Health:  InitialHealth,
				Faction: P1Faction,
			}
		}

		if b.isValidLoc(lb) && spawnFunc(lb) {
			b.Locs[lb] = &Robot{
				ID:      b.newID(),
				Health:  InitialHealth,
				Faction: P2Faction,
			}
		}
	}
}

func (b *Board) Update(ta, tb botapi.Turn_List) {
	c := make(collisionMap)
	b.addCollisions(c, ta)
	b.addCollisions(c, tb)

	// Move the bots to their new locations, unless they collide with something,
	// in which case just subtract 1 from their health and don't move them.

	for loc, bots := range c {
		// If there's only one bot trying to get somewhere, just move them there
		if len(bots) == 1 {
			b.moveBot(bots[0], loc)
		} else {
			// Multiple bots, hurt 'em
			for _, bot := range bots {
				b.hurtBot(bot, CollisionDamage)
			}
		}

	}
	// Get rid of anyone who died in a collision
	b.clearTheDead()

	// Ok, we've moved everyone into place and hurt them for bumping into each
	// other, now we issue attacks
	// We issue attacks first, because I don't like the idea of robots
	// self-destructing when someone could have killed them, it makes for better
	// strategy this way

	// Allow all attacks to be issued before removing bots, because there's no
	// good, sensical way to order attacks. They all happen simultaneously
	b.issueAttacks(ta)
	b.issueAttacks(tb)

	// Get rid of anyone who was viciously murdered
	b.clearTheDead()

	// Boom goes the dynamite
	b.issueSelfDestructs(ta)
	b.issueSelfDestructs(tb)

	// Get rid of anyone killed in some kamikaze-shenanigans
	b.clearTheDead()

	b.Round++

	if b.Round%NewBotsSpacing == 0 {
		b.SpawnBots(randomSpawnFuncGen(b))
	}
}

func (b *Board) issueAttacks(ts botapi.Turn_List) {
	for i := 0; i < ts.Len(); i++ {
		t := ts.At(i)
		if t.Which() != botapi.Turn_Which_attack {
			continue
		}

		// They're attacking
		loc, _ := b.fromID(RobotID(t.Id()))
		xOff, yOff := directionOffsets(t.Attack())
		attackLoc := Loc{
			X: loc.X + xOff,
			Y: loc.Y + yOff,
		}

		// If there's a bot at the attack location, make them sad
		// You *can* hurt attack your own robots
		victim := b.Locs[attackLoc]
		if victim != nil {
			b.hurtBot(victim, AttackDamage)
		}
	}
}

func (b *Board) issueSelfDestructs(ts botapi.Turn_List) {
	for i := 0; i < ts.Len(); i++ {
		t := ts.At(i)
		if t.Which() != botapi.Turn_Which_selfDestruct {
			continue
		}

		// They're Metro-booming on production:
		// (https://www.youtube.com/watch?v=NiM5ARaexPE)
		loc, bomber := b.fromID(RobotID(t.Id()))
		for _, boomLoc := range b.surrounding(loc) {
			// If there's a bot in the blast radius
			victim := b.Locs[boomLoc]
			if victim != nil {
				b.hurtBot(victim, DestructDamage)
			}
		}

		// Kill 'em
		b.hurtBot(bomber, SelfDamage)
	}
}

func (b *Board) fromID(id RobotID) (Loc, *Robot) {
	for loc, bot := range b.Locs {
		if bot.ID == id {
			return loc, bot
		}
	}
	return Loc{}, nil
}

func (b *Board) surrounding(loc Loc) []Loc {
	offs := []int{-1, 0, 1}

	// At most 8 surrounding locations
	vLocs := make([]Loc, 0, 8)
	for _, ox := range offs {
		for _, oy := range offs {
			// Skip the explosion location
			if ox == 0 && oy == 0 {
				continue
			}
			l := Loc{
				X: loc.X + ox,
				Y: loc.Y + oy,
			}
			if b.isValidLoc(l) {
				vLocs = append(vLocs, l)
			}
		}
	}
	return vLocs
}

func (b *Board) addCollisions(c collisionMap, ts botapi.Turn_List) {
	for i := 0; i < ts.Len(); i++ {
		t := ts.At(i)
		_, bot := b.fromID(RobotID(t.Id()))
		nextLoc := b.nextLoc(bot, t)
		// Add where they want to move
		c[nextLoc] = append(c[nextLoc], bot)
	}
}

func (b *Board) hurtBot(r *Robot, damage int) {
	r.Health -= damage
}

func (b *Board) clearTheDead() {
	var killKeys []Loc
	for loc, bot := range b.Locs {

		// Smite them
		if bot.Health <= 0 {
			killKeys = append(killKeys, loc)
		}
	}

	for _, loc := range killKeys {
		delete(b.Locs, loc)
	}
}

func (b *Board) moveBot(bot *Robot, loc Loc) error {
	oldLoc := b.robotLoc(bot)
	if manhattanDistance(oldLoc, loc) > 1 {
		return errors.New("Teleporting or some ish")
	}
	delete(b.Locs, oldLoc)
	b.Locs[loc] = bot
	return nil
}

func manhattanDistance(loc1, loc2 Loc) int {
	return abs(loc1.X-loc2.X) + abs(loc1.Y-loc2.Y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (b *Board) nextLoc(bot *Robot, t botapi.Turn) Loc {
	currentLoc := b.robotLoc(bot)
	// If they aren't moving, return their current loc
	if t.Which() != botapi.Turn_Which_move {
		return currentLoc
	}

	// They're moving, return where they're going

	xOff, yOff := directionOffsets(t.Move())
	nextLoc := Loc{
		X: currentLoc.X + xOff,
		Y: currentLoc.Y + yOff,
	}

	if b.isValidLoc(nextLoc) {
		return nextLoc
	}

	// TODO: Penalize people for creating incompetent bots that like travelling
	// to invalid locations, which is the case if we've reached here.
	return currentLoc
}

func directionOffsets(dir botapi.Direction) (x, y int) {
	var xOff, yOff int
	switch dir {
	case botapi.Direction_north:
		yOff = -1
	case botapi.Direction_south:
		yOff = 1
	case botapi.Direction_east:
		xOff = 1
	case botapi.Direction_west:
		xOff = -1
	}
	return xOff, yOff
}

func (b *Board) robotLoc(r *Robot) Loc {
	for loc, bot := range b.Locs {
		if bot.ID == r.ID {
			return loc
		}
	}
	return Loc{}
}

// IsFinished reports whether the game is finished.
func (b *Board) IsFinished() bool {
	return b.Round >= 100
}

// At returns the robot at a location or nil if not found.
func (b *Board) AtXY(x, y int) CellInfo {
	if !b.isValidLoc(Loc{x, y}) {
		return CellInfo{
			Bot:      nil,
			CellType: b.Cells[x][y],
		}

	}
	return CellInfo{
		Bot:      b.Locs[Loc{X: x, Y: y}],
		CellType: b.Cells[x][y],
	}
}

func (b *Board) isValidLoc(loc Loc) bool {
	return b.Cells[loc.X][loc.Y] == Valid
}

// ToWire converts the board to the wire representation with respect to the
// given faction (since the wire factions are us vs. them).
func (b *Board) ToWire(out botapi.Board, faction int) error {
	out.SetWidth(uint16(b.Size.X))
	out.SetHeight(uint16(b.Size.Y))
	out.SetRound(int32(b.Round))

	robots, err := botapi.NewRobot_List(out.Segment(), int32(len(b.Locs)))
	if err != nil {
		return err
	}
	if err = out.SetRobots(robots); err != nil {
		return err
	}

	n := 0
	for loc, r := range b.Locs {
		outr := robots.At(n)
		outr.SetId(uint32(r.ID))
		outr.SetX(uint16(loc.X))
		outr.SetY(uint16(loc.Y))
		outr.SetHealth(int16(r.Health))
		if r.Faction == faction {
			outr.SetFaction(botapi.Faction_mine)
		} else {
			outr.SetFaction(botapi.Faction_opponent)
		}
		n++
	}
	return nil
}

// A Robot is a single piece on a board.
type Robot struct {
	ID      RobotID
	Health  int
	Faction int
}

type RobotID uint32

func (id RobotID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id RobotID) GoString() string {
	return id.String()
}

// Loc is a position on a board.
type Loc struct {
	X, Y int
}

func (loc Loc) String() string {
	return fmt.Sprintf("(%d, %d)", loc.X, loc.Y)
}
