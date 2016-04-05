package engine

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bcspragu/Gobots/botapi"
)

const (
	P1Faction = 1
	P2Faction = 2

	InitialHealth = 50

	CollisionDamage = 5
	AttackDamage    = 10
	DestructDamage  = 15
	SelfDamage      = 1000 // Make them super dead

	NewBotsSpacing = 10 // Number of rounds to wait before spawning new bots
)

type (
	// CellInfo holds info about a which robot is in a cell and what type of cell
	// it is
	CellInfo struct {
		Bot      *Robot
		CellType CellType
	}

	CellType   int
	DamageType int

	collisionMap map[Loc][]botMove

	botMove struct {
		Bot      *Robot
		Location Loc
		Turn     botapi.Turn
	}
)

const (
	UnknownDamageType DamageType = iota
	Collision
	Attack
	Destruct
	Self
)

const (
	Invalid CellType = iota
	Valid
	Spawn
)

var (
	damageMap = map[DamageType]int{
		Collision: CollisionDamage,
		Attack:    AttackDamage,
		Destruct:  DestructDamage,
	}

	cellToWire = map[CellType]botapi.CellType{
		Invalid: botapi.CellType_invalid,
		Valid:   botapi.CellType_valid,
		Spawn:   botapi.CellType_spawn,
	}

	cellFromWire = map[botapi.CellType]CellType{
		botapi.CellType_invalid: Invalid,
		botapi.CellType_valid:   Valid,
		botapi.CellType_spawn:   Spawn,
	}
)

type Board struct {
	Locs map[Loc]*Robot

	Cells [][]CellType
	Size  Loc
	Round int

	NextID RobotID

	s Spawner
	c Typer

	leftSpawns []Loc
}

type BoardConfig struct {
	Size      Loc
	Spawner   Spawner
	CellTyper Typer
}

var DefaultConfig = BoardConfig{
	Size:      Loc{X: 17, Y: 17},
	Spawner:   NewRandomSpawn(2),
	CellTyper: NewLineSpawn(Loc{X: 17, Y: 17}),
}

func (b *Board) BotCount(faction int) (n int) {
	for _, bot := range b.Locs {
		if bot.Faction == faction {
			n++
		}
	}
	return
}

func (b *Board) CellsJS() [][]CellType {
	return b.Cells
}

// EmptyBoard creates an empty board of the given size.
func EmptyBoard(bc BoardConfig) *Board {
	b := &Board{
		Locs:  make(map[Loc]*Robot),
		Size:  bc.Size,
		Cells: make([][]CellType, bc.Size.X),
	}

	for i := 0; i < bc.Size.X; i++ {
		b.Cells[i] = make([]CellType, bc.Size.Y)
	}

	return b
}

func (b *Board) InitBoard(bc BoardConfig) {
	b.s = bc.Spawner
	b.c = bc.CellTyper

	for x := 0; x < b.Size.X; x++ {
		for y := 0; y < b.Size.Y; y++ {
			t := b.c.Type(x, y)
			b.Cells[x][y] = t
			if t == Spawn && x < b.Size.X/2 {
				b.leftSpawns = append(b.leftSpawns, Loc{x, y})
			}
		}
	}
	b.spawnBots()
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

func (b *Board) spawnBots() {
	// Clear out the spawn zone
	for _, locA := range b.leftSpawns {
		locB := Loc{b.Size.X - 1 - locA.X, locA.Y}
		delete(b.Locs, locA)
		delete(b.Locs, locB)
	}

	// Spawn() returns the list of locations to spawn bots at
	for _, locA := range b.s.Spawn(b.leftSpawns) {
		locB := Loc{b.Size.X - 1 - locA.X, locA.Y}
		b.Locs[locA] = &Robot{
			ID:      b.newID(),
			Health:  InitialHealth,
			Faction: P1Faction,
		}
		b.Locs[locB] = &Robot{
			ID:      b.newID(),
			Health:  InitialHealth,
			Faction: P2Faction,
		}
	}
}

func (b *Board) Update(ta, tb botapi.Turn_List) {
	// Put all the moves and bots into a list
	moves := make([]botMove, ta.Len()+tb.Len())
	for i := 0; i < ta.Len(); i++ {
		t := ta.At(i)
		loc, bot := b.fromID(RobotID(t.Id()))
		moves[i].Bot = bot
		moves[i].Turn = t
		moves[i].Location = loc
	}
	for i, l := 0, ta.Len(); i < tb.Len(); i++ {
		t := tb.At(i)
		_, bot := b.fromID(RobotID(t.Id()))
		moves[i+l].Bot = bot
		moves[i+l].Turn = t
	}
	c := make(collisionMap)
	b.addCollisions(c, moves)

	// Move the bots to their new locations, unless they collide with something,
	// in which case just subtract 1 from their health and don't move them.

	// TODO: This allows bots to swap places, which isn't allowed in the original
	// game.
	for loc, ms := range c {
		// If there's only one bot trying to get somewhere, just move them there
		if len(ms) == 1 {
			b.moveBot(ms[0].Bot, loc)
		} else {
			// Multiple bots, hurt 'em
			for _, m := range ms {
				b.hurtBot(m, Collision)
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
	b.issueAttacks(moves)

	// Get rid of anyone who was viciously murdered
	b.clearTheDead()

	// Boom goes the dynamite
	b.issueSelfDestructs(moves)

	// Get rid of anyone killed in some kamikaze-shenanigans
	b.clearTheDead()

	b.Round++

	if b.Round%NewBotsSpacing == 0 {
		b.spawnBots()
	}
}

func (b *Board) issueAttacks(moves []botMove) {
	for _, move := range moves {
		if move.Turn.Which() != botapi.Turn_Which_attack {
			continue
		}

		// They're attacking
		xOff, yOff := directionOffsets(move.Turn.Attack())
		attackLoc := Loc{
			X: move.Location.X + xOff,
			Y: move.Location.Y + yOff,
		}

		// If there's a bot at the attack location, make them sad
		// You *can* attack your own robots
		victim := b.Locs[attackLoc]
		if victim != nil {
			for _, m := range moves {
				if m.Bot.ID == victim.ID {
					b.hurtBot(m, Attack)
					break
				}
			}
		}
	}
}

func (b *Board) issueSelfDestructs(moves []botMove) {
	for _, move := range moves {
		if move.Turn.Which() != botapi.Turn_Which_selfDestruct {
			continue
		}

		// They're Metro-booming on production:
		// (https://www.youtube.com/watch?v=NiM5ARaexPE)
		for _, boomLoc := range b.surrounding(move.Location) {
			// If there's a bot in the blast radius
			victim := b.Locs[boomLoc]
			if victim != nil {
				for _, m := range moves {
					if m.Bot.ID == victim.ID {
						b.hurtBot(m, Destruct)
						break
					}
				}
			}
		}

		// Kill 'em
		b.hurtBot(move, Self)
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

func (b *Board) addCollisions(c collisionMap, moves []botMove) {
	for _, move := range moves {
		nextLoc := b.nextLoc(move)
		// Add where they want to move
		c[nextLoc] = append(c[nextLoc], move)
	}
}

func (b *Board) hurtBot(move botMove, dt DamageType) {
	switch dt {
	case Self:
		move.Bot.Health = 0
	case Attack, Destruct:
		// If they are guarding, they take half damage
		if move.Turn.Which() == botapi.Turn_Which_guard {
			move.Bot.Health -= damageMap[dt] / 2
		} else {
			move.Bot.Health -= damageMap[dt]
		}
	case Collision:
		// If they aren't guarding, they take damage
		if move.Turn.Which() != botapi.Turn_Which_guard {
			move.Bot.Health -= damageMap[dt]
		}
	}
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

func (b *Board) nextLoc(move botMove) Loc {
	currentLoc := b.robotLoc(move.Bot)
	// If they aren't moving, return their current loc
	if move.Turn.Which() != botapi.Turn_Which_move {
		return currentLoc
	}

	// They're moving, return where they're going

	xOff, yOff := directionOffsets(move.Turn.Move())
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
			CellType: Invalid,
		}
	}

	return CellInfo{
		Bot:      b.Locs[Loc{X: x, Y: y}],
		CellType: b.Cells[x][y],
	}
}

func (b *Board) isValidLoc(loc Loc) bool {
	if loc.X >= b.Size.X || loc.X < 0 || loc.Y < 0 || loc.Y >= b.Size.Y {
		return false
	}
	return b.Cells[loc.X][loc.Y] == Valid || b.Cells[loc.X][loc.Y] == Spawn
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

// ToWireWithInitial converts the board to the wire representation with respect
// to the given faction (since the wire factions are us vs. them), including
// information about which cells are which type.
func (b *Board) ToWireWithInitial(out botapi.InitialBoard, faction int) error {
	wireBoard, err := out.NewBoard()
	b.ToWire(wireBoard, faction)

	cells, err := botapi.NewCellType_List(out.Segment(), int32(b.Size.X*b.Size.Y))
	if err != nil {
		return err
	}

	if err = out.SetCells(cells); err != nil {
		return err
	}

	for x, col := range b.Cells {
		for y, cell := range col {
			cells.Set(x+y*b.Size.X, cellToWire[cell])
		}
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

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
