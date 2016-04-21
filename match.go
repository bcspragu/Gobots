package main

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bcspragu/Gobots/botapi"
	"github.com/bcspragu/Gobots/engine"
	gocontext "golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

type aiEndpoint struct {
	ds datastore

	// fields below are protected by mu
	mu     sync.Mutex
	online map[aiID]botapi.Ai
}

const (
	BoardSize = 17
)

func startAIEndpoint(addr string, ds datastore) (*aiEndpoint, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	e := &aiEndpoint{
		ds:     ds,
		online: make(map[aiID]botapi.Ai),
	}
	go e.listen(l)
	return e, nil
}

// listen runs in its own goroutine, listening for connections.
func (e *aiEndpoint) listen(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("ai endpoint: accept:", err)
			return
		}
		go e.handleConn(c)
	}
}

// handleConn runs in its own goroutine, started by listen.
func (e *aiEndpoint) handleConn(c net.Conn) {
	aic := &aiConnector{e: e}
	rc := rpc.NewConn(rpc.StreamTransport(c), rpc.MainInterface(botapi.AiConnector_ServerToClient(aic).Client))
	rc.Wait()
	aic.drop()
}

// listOnlineAIs lists the active AIs connected to the server right now.  The AIs can be passed over to startMatch.
func (e *aiEndpoint) listOnlineAIs() []onlineAI {
	e.mu.Lock()
	defer e.mu.Unlock()
	online := make([]onlineAI, 0, len(e.online))
	for id, client := range e.online {
		info, err := e.ds.lookupAI(id)
		if err != nil {
			log.Printf("Failed to lookup AI %s: %v", id, err)
			continue
		}
		online = append(online, onlineAI{
			Info:   *info,
			client: client,
		})
	}
	return online
}

// connect adds an online AI
func (e *aiEndpoint) connect(name, token string, ai botapi.Ai) (aiID, error) {
	infos, err := e.ds.listAIsForUser(accessToken(token))
	if err != nil {
		return "", err
	}
	var id aiID
	for _, info := range infos {
		if info.Name == name {
			id = info.ID
		}
	}
	// No bot was found with that name
	if id == aiID("") {
		id, err = e.ds.createAI(&aiInfo{
			Name: name,
		}, accessToken(token))
		if err != nil {
			return "", err
		}
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.online[id]; exists {
		return "", errors.New("That bot is already connected. Choose a new name.")
	} else {
		e.online[id] = ai
	}

	return id, nil
}

// removeAIs drops AIs from online, usually via disconnection.
func (e *aiEndpoint) removeAIs(ids []aiID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, i := range ids {
		delete(e.online, i)
	}
}

type aiConnector struct {
	e   *aiEndpoint
	ais []aiID
}

func (aic *aiConnector) Connect(call botapi.AiConnector_connect) error {
	creds, _ := call.Params.Credentials()
	tok, _ := creds.SecretToken()
	name, _ := creds.BotName()
	id, err := aic.e.connect(name, tok, call.Params.Ai())
	if err != nil {
		return err
	}
	aic.ais = append(aic.ais, id)
	return nil
}

func (aic *aiConnector) drop() {
	aic.e.removeAIs(aic.ais)
}

func runMatch(gidCh chan<- gameID, ctx gocontext.Context, ds datastore, aiA, aiB *onlineAI, bc engine.BoardConfig) error {
	sTime := time.Now()
	// Create new board and store it.
	b := engine.EmptyBoard(bc)
	b.InitBoard()
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	wb, _ := botapi.NewRootInitialBoard(seg)
	b.ToWireWithInitial(wb, engine.P1Faction)
	gid, err := ds.startGame(aiA.Info.ID, aiB.Info.ID, wb)
	if err != nil {
		return err
	}
	gidCh <- gid

	// Run the game
	for !b.IsFinished() {
		turnCtx, _ := gocontext.WithTimeout(ctx, 30*time.Second)
		chA, chB := make(chan turnResult), make(chan turnResult)
		go aiA.takeTurn(turnCtx, gid, b, engine.P1Faction, chA)
		go aiB.takeTurn(turnCtx, gid, b, engine.P2Faction, chB)
		ra, rb := <-chA, <-chB
		if ra.err.HasError() {
			log.Printf("Errors from AI ID %s: %v", aiA.Info.ID, ra.err)
		}
		if rb.err.HasError() {
			log.Printf("Errors from AI ID %s: %v", aiB.Info.ID, rb.err)
		}
		b.Update(ra.results, rb.results)
		_, s, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return err
		}
		r, err := botapi.NewRootReplay_Round(s)
		if err != nil {
			return err
		}

		wireBoard, err := r.NewEndBoard()
		if err != nil {
			return err
		}
		b.ToWire(wireBoard, engine.P1Faction)

		ral, rbl := ra.results.Len(), rb.results.Len()
		turns, err := botapi.NewTurn_List(r.Segment(), int32(ral+rbl))
		if err != nil {
			return err
		}
		for i := 0; i < ral; i++ {
			t := ra.results.At(i)
			if err := turns.Set(i, t); err != nil {
				return err
			}
		}
		for i := 0; i < rbl; i++ {
			t := rb.results.At(i)
			if err := turns.Set(i+ral, t); err != nil {
				return err
			}
		}
		r.SetMoves(turns)
		db.addRound(gid, r)
	}

	gInfo := &gameInfo{
		ID:        gid,
		AI1:       &aiA.Info,
		AI2:       &aiB.Info,
		AI1Score:  b.BotCount(1),
		AI2Score:  b.BotCount(2),
		StartTime: sTime,
		EndTime:   time.Now(),
	}
	return db.finishGame(gid, &aiA.Info, &aiB.Info, gInfo)
}

type onlineAI struct {
	Info   aiInfo
	client botapi.Ai
}

type turnResult struct {
	results botapi.Turn_List
	err     turnError
}

func (oa *onlineAI) takeTurn(ctx gocontext.Context, gid gameID, b *engine.Board, faction engine.Faction, ch chan<- turnResult) {
	results, err := oa.client.TakeTurn(ctx, func(p botapi.Ai_takeTurn_Params) error {
		iwb, err := p.NewBoard()
		if err != nil {
			return err
		}
		wb, err := iwb.NewBoard()
		if err != nil {
			return err
		}
		wb.SetGameId(string(gid))
		return b.ToWireWithInitial(iwb, faction)
	}).Struct()
	var te turnError
	if err != nil {
		te = append(te, err)
	}

	tl, err := results.Turns()
	if err != nil {
		te = append(te, err)
	}
	ch <- turnResult{tl, te}
}

type turnError []error

func (t turnError) Error() string {
	var e string
	for _, err := range t {
		e += err.Error()
	}
	return e
}

func (t turnError) HasError() bool {
	return len(t) > 0
}
