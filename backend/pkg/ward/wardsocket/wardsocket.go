package wardsocket

import (
	"errors"
	"sync"

	"paint.pecet.it/pkg/ward"
)

type WardSocket struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func New(w *ward.Ward) *WardSocket {
	return &WardSocket{
		rooms: make(map[string]*Room),
	}
}

func (m *WardSocket) GetRoom(roomIdent string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[roomIdent]
	return r, ok
}

func (m *WardSocket) SetRoom(room *Room, roomIdent string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[roomIdent] = room
}

func (m *WardSocket) DeleteRoom(roomIdent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if room, ok := m.rooms[roomIdent]; ok {
		room.cMu.Lock()
		for _, _ = range room.clients {

		}
		// TODO
		delete(m.rooms, roomIdent)
		return nil
	}
	return errors.New("room not found: " + roomIdent)
}
