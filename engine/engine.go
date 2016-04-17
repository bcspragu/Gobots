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
	// Cell holds info about a which robot is in a cell and what type of cell
	// it is
	Cell struct {
		Bot  *Robot
		Type CellType
	}

	CellType   int
	DamageType int

	collisionMap map[Loc][]*botMove
	moveMap      map[RobotID]*botMove

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
	Cells [][]Cell
	Bots  map[RobotID]Loc

	Round  int
	NextID RobotID

	config   BoardConfig
	p1Spawns []Loc
}

type BoardConfig struct {
	Size      Loc
	Spawner   Spawner
	CellTyper Typer
	NumRounds int
}

var DefaultConfig = BoardConfig{
	Size:      Loc{X: 17, Y: 17},
	Spawner:   NewRandomSpawn(2),
	CellTyper: NewLineSpawn(Loc{X: 17, Y: 17}),
	NumRounds: 100,
}

func (b *Board) BotCount(faction int) (n int) {
	for _, loc := range b.Bots {
		bot := b.Cells[loc.X][loc.Y].Bot
		if bot.Faction == faction {
			n++
		}
	}
	return
}

func (b *Board) CellsJS() [][]Cell {
	return b.Cells
}

// EmptyBoard creates an empty board of the given size.
func EmptyBoard(bc BoardConfig) *Board {
	b := &Board{
		Bots:   make(map[RobotID]Loc),
		Cells:  make([][]Cell, bc.Size.X),
		config: bc,
	}

	// TODO: Use other, more efficient allocation method
	for i := 0; i < b.Width(); i++ {
		b.Cells[i] = make([]Cell, b.Height())
	}

	return b
}

func (b *Board) Width() int {
	return b.config.Size.X
}

func (b *Board) Height() int {
	return b.config.Size.Y
}

func (b *Board) InitBoard() {
	for x := 0; x < b.Width(); x++ {
		for y := 0; y < b.Height(); y++ {
			t := b.config.CellTyper.Type(x, y)
			b.Cells[x][y].Type = t
			// TODO: New scheme to allow for bots spawning differently
			if t == Spawn && x < b.Width()/2 {
				b.p1Spawns = append(b.p1Spawns, Loc{x, y})
			}
		}
	}
	b.spawnBots()
}

func (b *Board) newID() RobotID {
	b.NextID++
	return b.NextID
}

func (b *Board) addBot(bot *Robot, l Loc) {
	if !b.inBounds(l) {
		return
	}

	b.Bots[bot.ID] = l
	b.Cells[l.X][l.Y].Bot = bot
}

func (b *Board) removeBot(l Loc) {
	if !b.inBounds(l) {
		return
	}

	if bot := b.Cells[l.X][l.Y].Bot; bot != nil {
		delete(b.Bots, bot.ID)
		b.Cells[l.X][l.Y].Bot = nil
	}
}

func (b *Board) spawnBots() {
	// Clear out the spawn zone
	for _, locA := range b.p1Spawns {
		locB := Loc{b.Width() - 1 - locA.X, locA.Y}
		b.removeBot(locA)
		b.removeBot(locB)
	}

	// Spawn() returns the list of locations to spawn bots at
	for _, locA := range b.config.Spawner.Spawn(b.p1Spawns) {
		locB := Loc{b.Width() - 1 - locA.X, locA.Y}

		b.addBot(&Robot{
			ID:      b.newID(),
			Health:  InitialHealth,
			Faction: P1Faction,
		}, locA)

		b.addBot(&Robot{
			ID:      b.newID(),
			Health:  InitialHealth,
			Faction: P2Faction,
		}, locB)
	}
}

func (b *Board) Update(ta, tb botapi.Turn_List) {
	// Put all the moves and bots into a map
	moves := make(moveMap)
	for i := 0; i < ta.Len(); i++ {
		t := ta.At(i)
		id := RobotID(t.Id())
		loc, bot := b.fromID(id)

		moves[id] = &botMove{
			Bot:      bot,
			Turn:     t,
			Location: loc,
		}
	}

	for i := 0; i < tb.Len(); i++ {
		t := tb.At(i)
		id := RobotID(t.Id())
		loc, bot := b.fromID(id)

		moves[id] = &botMove{
			Bot:      bot,
			Turn:     t,
			Location: loc,
		}
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

	if b.Round%NewBotsSpacing == 0 && b.Round < b.config.NumRounds {
		b.spawnBots()
	}
}

func (b *Board) issueAttacks(moves moveMap) {
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
		if victim := b.Cells[attackLoc.X][attackLoc.Y].Bot; victim != nil {
			b.hurtBot(moves[victim.ID], Attack)
		}
	}
}

func (b *Board) issueSelfDestructs(moves moveMap) {
	for _, move := range moves {
		if move.Turn.Which() != botapi.Turn_Which_selfDestruct {
			continue
		}

		// They're Metro-booming on production:
		// (https://www.youtube.com/watch?v=NiM5ARaexPE)
		for _, boomLoc := range b.surrounding(move.Location) {
			// If there's a bot in the blast radius
			if victim := b.Cells[boomLoc.X][boomLoc.Y].Bot; victim != nil {
				b.hurtBot(moves[victim.ID], Destruct)
			}
		}

		// Kill 'em
		b.hurtBot(move, Self)
	}
}

func (b *Board) fromID(id RobotID) (Loc, *Robot) {
	if loc, ex := b.Bots[id]; ex {
		return loc, b.Cells[loc.X][loc.Y].Bot
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

func (b *Board) addCollisions(c collisionMap, moves moveMap) {
	for _, move := range moves {
		nextLoc := b.nextLoc(move)
		// Add where they want to move
		c[nextLoc] = append(c[nextLoc], move)
	}
}

func (b *Board) hurtBot(move *botMove, dt DamageType) {
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
	for _, loc := range b.Bots {
		bot := b.Cells[loc.X][loc.Y].Bot
		// Smite them
		if bot.Health <= 0 {
			b.removeBot(loc)
		}
	}
}

func (b *Board) moveBot(bot *Robot, loc Loc) error {
	oldLoc := b.robotLoc(bot)
	if manhattanDistance(oldLoc, loc) > 1 {
		return errors.New("Teleporting or some ish")
	}
	b.Cells[oldLoc.X][oldLoc.Y].Bot = nil
	b.Cells[loc.X][loc.Y].Bot = bot
	b.Bots[bot.ID] = loc
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

func (b *Board) nextLoc(move *botMove) Loc {
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
	switch dir {
	case botapi.Direction_north:
		y = -1
	case botapi.Direction_south:
		y = 1
	case botapi.Direction_east:
		x = 1
	case botapi.Direction_west:
		x = -1
	}
	return
}

func (b *Board) robotLoc(r *Robot) Loc {
	return b.Bots[r.ID]
}

// IsFinished reports whether the game is finished.
func (b *Board) IsFinished() bool {
	return b.Round >= 100
}

func (b *Board) At(l Loc) *Robot {
	if !b.inBounds(l) {
		return nil
	}
	return b.Cells[l.X][l.Y].Bot
}

func (b *Board) inBounds(loc Loc) bool {
	return loc.X < b.Width() && loc.X >= 0 && loc.Y >= 0 && loc.Y < b.Height()
}

func (b *Board) isValidLoc(loc Loc) bool {
	return b.inBounds(loc) &&
		(b.Cells[loc.X][loc.Y].Type == Valid || b.Cells[loc.X][loc.Y].Type == Spawn)
}

// ToWire converts the board to the wire representation with respect to the
// given faction (since the wire factions are us vs. them).
func (b *Board) ToWire(out botapi.Board, faction int) error {
	out.SetWidth(uint16(b.Width()))
	out.SetHeight(uint16(b.Height()))
	out.SetRound(int32(b.Round))

	robots, err := botapi.NewRobot_List(out.Segment(), int32(len(b.Bots)))
	if err != nil {
		return err
	}
	if err = out.SetRobots(robots); err != nil {
		return err
	}

	n := 0
	for _, loc := range b.Bots {
		r := b.Cells[loc.X][loc.Y].Bot
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

	cells, err := botapi.NewCellType_List(out.Segment(), int32(b.Width()*b.Height()))
	if err != nil {
		return err
	}

	if err = out.SetCells(cells); err != nil {
		return err
	}

	for x, col := range b.Cells {
		for y, cell := range col {
			cells.Set(x+y*b.Width(), cellToWire[cell.Type])
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
