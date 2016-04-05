package game

import (
	"fmt"
	"net"
	"time"

	"github.com/bcspragu/Gobots/botapi"
	"github.com/bcspragu/Gobots/engine"
	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

type localAI struct {
	botapi.Ai
}

type MatchResult struct {
	P1Score int
	P2Score int
}

func (m *MatchResult) String() string {
	outcome := "Tie"
	if m.P1Score > m.P2Score {
		outcome = "Player 1 wins"
	} else if m.P1Score < m.P2Score {
		outcome = "Player 2 wins"
	}
	return fmt.Sprintf("P1: %d P2: %d - %s", m.P1Score, m.P2Score, outcome)
}

func FightBots(f1, f2 Factory, n int) MatchResult {
	return fightN(f1, f2, 1)[0]
}

// FightBotsN plays a match between the two bots and returns the result.
func FightBotsN(f1, f2 Factory, n int) []MatchResult {
	return fightN(f1, f2, n)
}

func fightN(f1, f2 Factory, n int) []MatchResult {
	aiA := &aiAdapter{factory: f1, games: make(map[string]gameState)}
	aiB := &aiAdapter{factory: f2, games: make(map[string]gameState)}

	a1, a2 := net.Pipe()
	b1, b2 := net.Pipe()

	ca1, ca2 := rpc.StreamTransport(a1), rpc.StreamTransport(a2)
	cb1, cb2 := rpc.StreamTransport(b1), rpc.StreamTransport(b2)

	// Server-side
	srvA := botapi.Ai_ServerToClient(aiA)
	srvB := botapi.Ai_ServerToClient(aiB)
	serverConnA := rpc.NewConn(ca1, rpc.MainInterface(srvA.Client))
	serverConnB := rpc.NewConn(cb1, rpc.MainInterface(srvB.Client))
	defer serverConnA.Wait()
	defer serverConnB.Wait()

	// Client-side
	ctx := context.Background()

	clientConnA := rpc.NewConn(ca2)
	clientConnB := rpc.NewConn(cb2)
	defer clientConnA.Close()
	defer clientConnB.Close()

	clientA := localAI{botapi.Ai{Client: clientConnA.Bootstrap(ctx)}}
	clientB := localAI{botapi.Ai{Client: clientConnB.Bootstrap(ctx)}}

	matchRes := make([]MatchResult, n)
	// Run the game
	for i := 0; i < n; i++ {
		b := engine.EmptyBoard(engine.DefaultConfig)
		b.InitBoard(engine.DefaultConfig)
		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		wb, _ := botapi.NewRootInitialBoard(seg)
		b.ToWireWithInitial(wb, engine.P1Faction)

		for !b.IsFinished() {
			turnCtx, _ := context.WithTimeout(ctx, 30*time.Second)
			resA, _ := clientA.takeTurn(turnCtx, "0", b, engine.P1Faction)
			resB, _ := clientB.takeTurn(turnCtx, "0", b, engine.P2Faction)
			b.Update(resA, resB)
		}
		matchRes[i] = MatchResult{
			P1Score: b.BotCount(1),
			P2Score: b.BotCount(2),
		}
	}
	return matchRes
}

func (la *localAI) takeTurn(ctx context.Context, gid string, b *engine.Board, faction int) (botapi.Turn_List, error) {
	// TODO Probably don't ignore errors
	res, _ := la.TakeTurn(ctx, func(p botapi.Ai_takeTurn_Params) error {
		iwb, err := p.NewBoard()
		if err != nil {
			return err
		}
		wb, err := iwb.NewBoard()
		if err != nil {
			return err
		}
		wb.SetGameId(gid)
		return b.ToWireWithInitial(iwb, faction)
	}).Struct()
	return res.Turns()
}
