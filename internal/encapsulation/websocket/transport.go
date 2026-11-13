package websocket

import (
	"net/http"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
)

const (
	WriteWait      = 10 * time.Second
	PongWait       = 90 * time.Second
	PingPeriod     = 30 * time.Second
	MaxMessageSize = 4 << 20
	BufferSize     = 32 * 1024
)

func NewUpgrader() gorilla.Upgrader {
	return gorilla.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == ""
		},
	}
}

func NewDialer() gorilla.Dialer {
	return gorilla.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  15 * time.Second,
		EnableCompression: false,
	}
}

func ConfigureRead(conn *gorilla.Conn) {
	conn.SetReadLimit(MaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(PongWait))
	})
}

func WriteBinary(conn *gorilla.Conn, payload []byte) error {
	_ = conn.SetWriteDeadline(time.Now().Add(WriteWait))
	return conn.WriteMessage(gorilla.BinaryMessage, payload)
}

func StartKeepalive(conn *gorilla.Conn, stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(PingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if err := conn.WriteControl(gorilla.PingMessage, nil, time.Now().Add(WriteWait)); err != nil {
					return
				}
			}
		}
	}()
	return done
}

func CloseNormally(conn *gorilla.Conn) {
	_ = conn.WriteControl(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""), time.Now().Add(WriteWait))
}
