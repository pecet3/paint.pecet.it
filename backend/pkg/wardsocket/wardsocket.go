package wardsocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Client  *Client
}
type ByteEvent struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type Upgrader struct {
	HandshakeTimeout  time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	CheckOrigin       func(r *http.Request) bool
	EnableCompression bool
}

type WardSocket struct {
	upgrader *websocket.Upgrader
}

func New(upgrader *Upgrader) *WardSocket {
	if upgrader == nil {
		upgrader = &Upgrader{
			HandshakeTimeout:  5 * time.Second,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			CheckOrigin:       func(r *http.Request) bool { return true },
			EnableCompression: false,
		}
	}
	ws := &WardSocket{
		upgrader: &websocket.Upgrader{
			HandshakeTimeout:  upgrader.HandshakeTimeout,
			ReadBufferSize:    upgrader.ReadBufferSize,
			WriteBufferSize:   upgrader.WriteBufferSize,
			CheckOrigin:       upgrader.CheckOrigin,
			EnableCompression: upgrader.EnableCompression,
		},
	}
	return ws
}

func (ws *WardSocket) UpgradeRequest(w http.ResponseWriter, r *http.Request) (*Client, error) {
	ug := &websocket.Upgrader{
		HandshakeTimeout:  ws.upgrader.HandshakeTimeout,
		ReadBufferSize:    ws.upgrader.ReadBufferSize,
		WriteBufferSize:   ws.upgrader.WriteBufferSize,
		CheckOrigin:       ws.upgrader.CheckOrigin,
		EnableCompression: ws.upgrader.EnableCompression,
	}
	conn, err := ug.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}
