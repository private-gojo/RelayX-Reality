package server

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
	"github.com/private-gojo/relayx-reality/internal/bridge/connection"
	wsencap "github.com/private-gojo/relayx-reality/internal/encapsulation/websocket"
)

type Server struct {
	upstream string
	path     string
	token    string
	limit    chan struct{}
	upgrader gorilla.Upgrader
	mu       sync.Mutex
	active   map[*gorilla.Conn]struct{}
}

func NewServer(upstream, path, token string, maxSessions int) *Server {
	return &Server{
		upstream: upstream,
		path:     path,
		token:    token,
		limit:    make(chan struct{}, maxSessions),
		active:   make(map[*gorilla.Conn]struct{}),
		upgrader: wsencap.NewUpgrader(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc(s.path, s.handleTunnel)
	return mux
}

func (s *Server) handleTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !authorized(r.Header.Get("Authorization"), s.token) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	select {
	case s.limit <- struct{}{}:
		defer func() { <-s.limit }()
	default:
		http.Error(w, "session capacity reached", http.StatusServiceUnavailable)
		return
	}

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.track(ws, true)
	defer s.track(ws, false)
	defer ws.Close()

	dialer := net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	tcp, err := dialer.DialContext(r.Context(), "tcp", s.upstream)
	if err != nil {
		_ = ws.WriteControl(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseTryAgainLater, "upstream unavailable"), time.Now().Add(wsencap.WriteWait))
		return
	}
	defer tcp.Close()
	connection.Relay(tcp, ws)
}

func (s *Server) track(conn *gorilla.Conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if add {
		s.active[conn] = struct{}{}
		return
	}
	delete(s.active, conn)
}

func (s *Server) CloseAll() {
	s.mu.Lock()
	connections := make([]*gorilla.Conn, 0, len(s.active))
	for conn := range s.active {
		connections = append(connections, conn)
	}
	s.mu.Unlock()
	for _, conn := range connections {
		_ = conn.Close()
	}
}

func authorized(header, token string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	presented := strings.TrimPrefix(header, prefix)
	return len(presented) == len(token) &&
		subtle.ConstantTimeCompare([]byte(presented), []byte(token)) == 1
}
