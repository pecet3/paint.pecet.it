package wardsocket

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

type Upgrader struct {
	HandshakeTimeout  time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	CheckOrigin       func(r *http.Request) bool
	EnableCompression bool
}

type WardSocket struct {
	rooms map[string]*Room
	mu    sync.Mutex

	upgrader *websocket.Upgrader
}

//	if upgrader == nil {
//			upgrader = &Upgrader{
//				ReadBufferSize:  1024,
//				WriteBufferSize: 1024,
//				CheckOrigin: func(r *http.Request) bool {
//					return true
//				},
//			}
//		}
func New(w *ward.Ward, upgrader *Upgrader) *WardSocket {
	if upgrader == nil {
		upgrader = &Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}
	gorillaUpgrader := &websocket.Upgrader{
		HandshakeTimeout:  upgrader.HandshakeTimeout,
		ReadBufferSize:    upgrader.ReadBufferSize,
		WriteBufferSize:   upgrader.WriteBufferSize,
		CheckOrigin:       upgrader.CheckOrigin,
		EnableCompression: upgrader.EnableCompression,
	}
	return &WardSocket{
		rooms:    make(map[string]*Room),
		upgrader: gorillaUpgrader,
	}
}

func (m *WardSocket) GetRoom(roomIdent string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[roomIdent]
	return r, ok
}

func (m *WardSocket) AddRoom(room *Room) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[room.Ident] = room
}

func (m *WardSocket) DeleteRoom(roomIdent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if room, ok := m.rooms[roomIdent]; ok {
		room.Close()
		delete(m.rooms, roomIdent)
		return nil
	}
	return errors.New("room not found: " + roomIdent)
}

func (w *WardSocket) AssignRequestToRoom(wreq *ward.Request, room *Room) error {
	conn, err := w.upgrader.Upgrade(wreq.ResponseWriter, wreq.Http, nil)
	if err != nil {
		return err
	}

	room.JoinConn(conn, wreq)
	return nil
}
