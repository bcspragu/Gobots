package engine

import (
	"errors"
	"fmt"
	"sort"

	"github.com/bcspragu/Gobots/botapi"
)

const (
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

	moveStatus int
	CellType   int
	DamageType int

	collisionMap    map[Loc][]*botMove
	intersectionMap map[Loc][]*botMove
	moveMap         map[RobotID]*botMove

	botMove struct {
		Bot     *Robot
		Current Loc
		Next    Loc
		Turn    botapi.Turn
		Status  moveStatus
	}
)

const (
	Pending moveStatus = iota
	Failed
	Successful
	Checking
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

	OutOfBounds     = errors.New("Out of Bounds")
	AlreadyOccupied = errors.New("Already Occupied")
)

type LocPair struct {
	Old Loc
	New Loc
}

type Board struct {
	Cells [][]Cell
	Bots  map[RobotID]Loc

	Round  int
	NextID RobotID

	Config   BoardConfig
	p1Spawns []Loc

	// Internal move things
	moves         moveMap
	collisions    collisionMap
	intersections intersectionMap
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

func (b *Board) BotCount(faction Faction) (n int) {
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
		Config: bc,
	}

	// TODO: Use other, more efficient allocation method
	for i := 0; i < b.Width(); i++ {
		b.Cells[i] = make([]Cell, b.Height())
	}

	return b
}

func (b *Board) Width() int {
	return b.Config.Size.X
}

func (b *Board) Height() int {
	return b.Config.Size.Y
}

func (b *Board) InitBoard() {
	for x := 0; x < b.Width(); x++ {
		for y := 0; y < b.Height(); y++ {
			t := b.Config.CellTyper.Type(x, y)
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

// Assume the caller knows what it's doing don't check the cell we're moving to
func (b *Board) moveBot(bot *Robot, l LocPair) {
	if !b.inBounds(l.Old) || !b.inBounds(l.New) {
		return
	}

	if l.Old == l.New {
		return
	}

	b.Cells[l.New.X][l.New.Y].Bot = bot
	b.Bots[bot.ID] = l.New

	// Check the old locations and see if something new has moved in. If not,
	// clear it out.
	if oldBot := b.Cells[l.Old.X][l.Old.Y].Bot; oldBot != nil && oldBot.ID == bot.ID {
		b.Cells[l.Old.X][l.Old.Y].Bot = nil
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
	for _, locA := range b.Config.Spawner.Spawn(b.p1Spawns) {
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
	b.moves = make(moveMap)
	b.addMoves(ta, P1Faction)
	b.addMoves(tb, P2Faction)
	b.moveBots()
	// Get rid of anyone who died in a collision
	b.clearTheDead()

	// Ok, we've moved everyone into place and hurt them for bumping into each
	// other, now we issue attacks
	// We issue attacks first, because I don't like the idea of robots
	// self-destructing when someone could have killed them, it makes for better
	// strategy this way

	// Allow all attacks to be issued before removing bots, because there's no
	// good, sensical way to order attacks. They all happen simultaneously
	b.issueAttacks()

	// Get rid of anyone who was viciously murdered
	b.clearTheDead()

	// Boom goes the dynamite
	b.issueSelfDestructs()

	// Get rid of anyone killed in some kamikaze-shenanigans
	b.clearTheDead()

	fmt.Println("Finished Round ", b.Round)
	b.Round++
	if b.Round%NewBotsSpacing == 0 && b.Round < b.Config.NumRounds {
		b.spawnBots()
	}
}

// TODO: Don't trust the client so much, someone not using the `game` package
// could easily crash any of this. Problems:
// - We don't check if robot IDS
// - We assume that each list contains a move for each robot

func (b *Board) addMoves(tl botapi.Turn_List, f Faction) {
	for i := 0; i < tl.Len(); i++ {
		t := tl.At(i)
		id := RobotID(t.Id())
		loc, bot := b.fromID(id)
		// Ignore the move if that bot doesn't exist or isn't owned by the right
		// player
		if bot == nil || (bot != nil && bot.Faction != f) {
			continue
		}

		b.moves[id] = &botMove{
			Bot:     bot,
			Turn:    t,
			Current: loc,
			Next:    b.nextLoc(bot, t),
		}
	}
}

func (b *Board) setBotStatus(bot *Robot) moveStatus {
	move := b.moves[bot.ID]
	// We set our status to checking so we know which bot we're working on
	move.Status = Checking
	// If they don't plan on going anywhere (wait, attack, guard, etc), that's
	// fine with us. We consider not moving to be failed because that helps us
	// unravel our dependency chain
	if move.Current == move.Next {
		move.Status = Failed
		return move.Status
	}

	cols := b.collisions[move.Next]
	if len(cols) > 1 {
		move.Status = Failed
		return move.Status
	}

	ints := b.intersections[move.intersection()]
	if len(ints) > 1 {
		move.Status = Failed
		return move.Status
	}

	// If we're still here we can continue checking to see if their move is
	// valid. Next step: check if there's someone where they want to be
	move.Status = b.unravel(move)
	return move.Status
}

func (b *Board) unravel(startMove *botMove) moveStatus {
	nBot := b.At(startMove.Next)
	if nBot != nil {
		// Someone is in the spot we're heading, check if they've been greenlighted to move
		nMove := b.moves[nBot.ID]
		switch nMove.Status {
		case Successful, Failed:
			// If we know the person in our spot has succeeded or failed, that
			// determines our success as it is.
			return nMove.Status
		case Checking:
			// Similarly, if we've made it back to ourself (Checking), then everyone
			// can move
			return Successful
		case Pending:
			// If we don't know their status, we'll have to check them. Incoming recursion
			return b.setBotStatus(nBot)
		}
	}
	// If there's nobody there, go for it
	return Successful
}

func (b *Board) moveBots() {
	b.addCollisions()

	for _, move := range b.moves {
		// Pick a random bot in Pending
		if move.Status == Pending {
			// This is where all the magic happens
			b.setBotStatus(move.Bot)
		}
	}

	for _, move := range b.moves {
		if move.Status == Successful {
			b.moveBot(move.Bot, LocPair{Old: move.Current, New: move.Next})
		} else if move.Status == Failed && move.Turn.Which() == botapi.Turn_Which_move {
			b.hurtBot(move, Collision)
		}
	}
}

func (b *Board) issueAttacks() {
	for _, move := range b.moves {
		if move.Turn.Which() != botapi.Turn_Which_attack {
			continue
		}

		// They're attacking
		xOff, yOff := directionOffsets(move.Turn.Attack())
		attackLoc := Loc{
			X: move.Current.X + xOff,
			Y: move.Current.Y + yOff,
		}

		// If there's a bot at the attack location, make them sad
		// You *can* attack your own robots
		if victim := b.Cells[attackLoc.X][attackLoc.Y].Bot; victim != nil {
			b.hurtBot(b.moves[victim.ID], Attack)
		}
	}
}

func (b *Board) issueSelfDestructs() {
	for _, move := range b.moves {
		if move.Turn.Which() != botapi.Turn_Which_selfDestruct {
			continue
		}

		// They're Metro-booming on production:
		// (https://www.youtube.com/watch?v=NiM5ARaexPE)
		for _, boomLoc := range b.surrounding(move.Current) {
			// If there's a bot in the blast radius
			if victim := b.Cells[boomLoc.X][boomLoc.Y].Bot; victim != nil {
				b.hurtBot(b.moves[victim.ID], Destruct)
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

func (b *Board) addCollisions() {
	b.collisions = make(collisionMap)
	b.intersections = make(intersectionMap)
	for _, move := range b.moves {
		in := move.intersection()
		b.intersections[in] = append(b.intersections[in], move)

		b.collisions[move.Next] = append(b.collisions[move.Next], move)
	}
}

func (bm *botMove) intersection() Loc {
	return Loc{
		X: bm.Current.X + bm.Next.X,
		Y: bm.Current.Y + bm.Next.Y,
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
	for i, loc := range b.Bots {
		bot := b.Cells[loc.X][loc.Y].Bot
		if bot == nil {
			fmt.Printf("b.Bots thinks %d: %#v, but b.Cells thinks it's nil\n", i, loc)
		}
		// Smite them
		if bot.Health <= 0 {
			fmt.Printf("Killing %d because their health is %d\n", bot.ID, bot.Health)
			b.removeBot(loc)
		}
	}
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

func (b *Board) nextLoc(bot *Robot, turn botapi.Turn) Loc {
	currentLoc := b.robotLoc(bot)
	// If they aren't moving, return their current loc
	if turn.Which() != botapi.Turn_Which_move {
		return currentLoc
	}

	// They're moving, return where they're going

	xOff, yOff := directionOffsets(turn.Move())
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
func (b *Board) ToWire(out botapi.Board, faction Faction) error {
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

	bots := make([]*Robot, len(b.Bots))

	n := 0
	for _, loc := range b.Bots {
		bots[n] = b.Cells[loc.X][loc.Y].Bot
		n++
	}

	sort.Sort(ByID(bots))

	for i, r := range bots {
		loc := b.Bots[r.ID]
		outr := robots.At(i)
		outr.SetId(uint32(r.ID))
		outr.SetX(uint16(loc.X))
		outr.SetY(uint16(loc.Y))
		outr.SetHealth(int16(r.Health))
		if r.Faction == faction {
			outr.SetFaction(botapi.Faction_mine)
		} else {
			outr.SetFaction(botapi.Faction_opponent)
		}
	}
	return nil
}

// ToWireWithInitial converts the board to the wire representation with respect
// to the given faction (since the wire factions are us vs. them), including
// information about which cells are which type.
func (b *Board) ToWireWithInitial(out botapi.InitialBoard, faction Faction) error {
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

// Loc is a position on a board.
type Loc struct {
	X, Y int
}

// Intersection is the line between two positions, multipled by two
type Intersection struct {
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

func (b *Board) printBoard() {
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if b.Cells[x][y].Type == Invalid {
				fmt.Print("[x]")
			} else if b.Cells[x][y].Bot == nil {
				fmt.Print("[ ]")
			} else {
				fmt.Print("[", b.Cells[x][y].Bot.ID, "]")
			}
		}
		fmt.Println("")
	}
}
