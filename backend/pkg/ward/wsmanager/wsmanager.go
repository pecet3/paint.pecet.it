package wsmanager

import (
	"errors"
	"sync"
)

type WsManager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func New() *WsManager {
	return &WsManager{
		rooms: make(map[string]*Room),
	}
}

func (m *WsManager) GetRoom(roomIdent string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[roomIdent]
	return r, ok
}

func (m *WsManager) SetRoom(room *Room, roomIdent string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[roomIdent] = room
}

func (m *WsManager) DeleteRoom(roomIdent string) error {
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
