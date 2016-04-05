package game

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bcspragu/Gobots/botapi"
	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2/rpc"
)

var (
	flags         = flag.NewFlagSet("game flags", flag.ContinueOnError)
	serverAddress = flags.String("server_address", "localhost:8001", "address of API server")
	retryInterval = flags.Duration("retry_interval", 10*time.Second, "how often (in seconds) to retry connecting to the server after losing a connection")
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
		games:   make(map[string]gameState),
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

// connect wraps the Dial and RegisterAI functionality
func connect(name, token string, factory Factory) (*Client, error) {
	c, err := Dial(*serverAddress)
	if err != nil {
		return c, fmt.Errorf("Failed to connect to server: %v", err)
	}
	if err = c.RegisterAI(name, token, factory); err != nil {
		return c, fmt.Errorf("Failed to register bot: %v", err)
	}
	return c, err
}

// StartServerForFactory connects to the server with the given robot name and
// user token, and registers the robot provided by the factory function
func StartServerForFactory(name, token string, factory Factory) {
	flags.SetOutput(ioutil.Discard)
	flags.Parse(os.Args[1:])

	c, err := connect(name, token, factory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
	fmt.Fprintf(os.Stderr, "Connected bot %s. Ctrl-C or send SIGINT to disconnect.\n", name)
	retryChan := make(chan time.Time)
	cWait := func() {
		c.conn.Wait()
		retryChan <- time.Now()
	}
	go cWait()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT)

	// Wait for either our connection with the server to terminate, or the user
	// to mercilessly silence their bot.
loop:
	for {
		select {
		case <-retryChan:
			fmt.Fprintln(os.Stderr, "Lost connection to server, trying to reconnect...")
			c, err = connect(name, token, factory)
			if err == nil {
				fmt.Fprintf(os.Stderr, "Reconnected bot %s successfully!\n", name)
				// Wait until we lose the connection again
				go cWait()
			} else {
				if strings.Contains(err.Error(), "connection refused") {
					// If we can't connect, just wait ten (or --retry_interval) seconds and try again
					go func() {
						retryChan <- <-time.After(*retryInterval)
					}()
				} else {
					// Fail on all other errors, like the server saying you have an invalid token
					fmt.Fprintln(os.Stderr, err)
					os.Exit(exitFail)
				}
			}
		case <-sig:
			signal.Stop(sig)
			break loop
		}
	}

	signal.Stop(sig)
	fmt.Fprintln(os.Stderr, "Interrupted. Quitting...")
	if c != nil {
		if err := c.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing our connection:", err)
			os.Exit(exitFail)
		}
	}
}
