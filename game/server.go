package game

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bcspragu/Gobots/botapi"
	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2/rpc"
)

// Client represents a connection to the game server.
type Client struct {
	conn      *rpc.Conn
	connector botapi.AiConnector
}

// Dial connects to a server at the given TCP address.
func Dial(addr string) (*Client, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn := rpc.NewConn(rpc.StreamTransport(c))
	return &Client{
		conn:      conn,
		connector: botapi.AiConnector{Client: conn.Bootstrap(context.TODO())},
	}, nil
}

// Close terminates the connection to the server.
func (c *Client) Close() error {
	return c.conn.Close()
}

// RegisterAI adds an AI implementation for the token given by the website.
// The AI factory function will be called for each new game encountered.
func (c *Client) RegisterAI(name, token string, factory Factory) error {
	a := botapi.Ai_ServerToClient(&aiAdapter{
		factory: factory,
		games:   make(map[string]AI),
	})
	_, err := c.connector.Connect(context.TODO(), func(r botapi.ConnectRequest) error {
		creds, err := r.NewCredentials()
		if err != nil {
			return err
		}
		err = creds.SetBotName(name)
		if err != nil {
			return err
		}
		err = creds.SetSecretToken(token)
		if err != nil {
			return err
		}
		r.SetAi(a)
		return nil
	}).Struct()
	return err
}

func StartServerForFactory(name, token string, factory Factory) {
	c, err := Dial("localhost:8001")
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial:", err)
		os.Exit(exitFail)
	}
	if err = c.RegisterAI(name, token, factory); err != nil {
		fmt.Fprintln(os.Stderr, "register:", err)
		os.Exit(exitFail)
	}
	fmt.Fprintf(os.Stderr, "Connected bot %s. Ctrl-C or send SIGINT to disconnect.", name)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT)
	<-sig
	signal.Stop(sig)
	fmt.Fprintln(os.Stderr, "Interrupted. Quitting...")
	if err := c.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "close:", err)
		os.Exit(exitFail)
	}
}

func StartServerForBot(name, token string, ai AI) {
	factory := func(gameID string) AI {
		return ai
	}
	StartServerForFactory(name, token, factory)
}
