package connection

import (
	"io"
	"net"
	"sync"

	gorilla "github.com/gorilla/websocket"
	wsencap "github.com/private-gojo/relayx-reality/internal/encapsulation/websocket"
)

func Relay(tcp net.Conn, ws *gorilla.Conn) {
	wsencap.ConfigureRead(ws)

	stopPing := make(chan struct{})
	pingDone := wsencap.StartKeepalive(ws, stopPing)

	done := make(chan struct{}, 2)
	go func() {
		buffer := make([]byte, wsencap.BufferSize)
		for {
			n, err := tcp.Read(buffer)
			if n > 0 {
				if writeErr := wsencap.WriteBinary(ws, buffer[:n]); writeErr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}()

	go func() {
		for {
			messageType, payload, err := ws.ReadMessage()
			if err != nil || messageType != gorilla.BinaryMessage {
				break
			}
			if err := writeAll(tcp, payload); err != nil {
				break
			}
		}
		done <- struct{}{}
	}()

	<-done
	_ = tcp.Close()
	wsencap.CloseNormally(ws)
	_ = ws.Close()
	<-done
	close(stopPing)
	<-pingDone
}

func writeAll(writer net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

var _ sync.Locker
