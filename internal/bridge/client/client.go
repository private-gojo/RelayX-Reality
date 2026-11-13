package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	gorilla "github.com/gorilla/websocket"
	"github.com/private-gojo/relayx-reality/internal/bridge/connection"
	wsencap "github.com/private-gojo/relayx-reality/internal/encapsulation/websocket"
)

type Client struct {
	localAddr string
	remoteURL string
	token     string
	limit     chan struct{}
	dialer    gorilla.Dialer
}

func NewClient(localAddr, remoteURL, token string, maxSessions int) (*Client, error) {
	u, err := url.Parse(remoteURL)
	if err != nil || u.Scheme != "wss" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("REMOTE_WS_URL must be a wss:// URL without userinfo or fragment")
	}
	host, _, err := net.SplitHostPort(localAddr)
	if err != nil || host == "" {
		return nil, errors.New("LOCAL_ADDR must include a host and port")
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, errors.New("LOCAL_ADDR must bind to loopback only")
	}
	if len(token) < 32 {
		return nil, errors.New("BRIDGE_TOKEN must be at least 32 characters")
	}
	if maxSessions < 1 || maxSessions > 1000 {
		return nil, errors.New("MAX_SESSIONS must be from 1 to 1000")
	}
	return &Client{
		localAddr: localAddr,
		remoteURL: remoteURL,
		token:     token,
		limit:     make(chan struct{}, maxSessions),
		dialer:    wsencap.NewDialer(),
	}, nil
}

func (c *Client) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", c.localAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", c.localAddr, err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	log.Printf("[CLIENT] listening=%s remote=%s", c.localAddr, c.remoteURL)
	for {
		localConn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept local connection: %w", err)
		}
		select {
		case c.limit <- struct{}{}:
			go func() {
				defer func() { <-c.limit }()
				c.handleLocal(localConn)
			}()
		default:
			_ = localConn.Close()
		}
	}
}

func (c *Client) handleLocal(localConn net.Conn) {
	defer localConn.Close()
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)
	ws, response, err := c.dialer.Dial(c.remoteURL, header)
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
	if err != nil {
		log.Printf("[CLIENT] websocket connect failed: %v", err)
		return
	}
	defer ws.Close()
	connection.Relay(localConn, ws)
}

