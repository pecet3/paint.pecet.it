package paint

import (
	"context"
	"fmt"
	"sync"

	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

type Paint struct {
	rooms map[string]*PaintRoom
	mu    sync.RWMutex

	ws *wardsocket.WardSocket
}

func New(ws *wardsocket.WardSocket) *Paint {
	return &Paint{
		rooms: make(map[string]*PaintRoom),
		ws:    ws,
	}
}

func (p *Paint) GetRoom(roomIdent string) *PaintRoom {
	p.mu.RLock()
	defer p.mu.RUnlock()

	room, exists := p.rooms[roomIdent]
	if !exists {
		return nil
	}
	return room
}

func (p *Paint) CreateRoom(cfg *RoomConfig) {
	ctx := context.Background()
	channel, ctx := wardsocket.NewChannel().WithCancelContext(ctx)
	room := newPaintRoom(cfg, channel)
	room.RegisterHandlers()
	room.Run(p, ctx)
	channel.Run(ctx)
	p.mu.Lock()
	p.rooms[cfg.Name] = room
	p.mu.Unlock()

}

func (p *Paint) DeleteRoom(roomIdent string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, exists := p.rooms[roomIdent]
	if exists {
		delete(p.rooms, roomIdent)
	}

}

func (p *Paint) ListRooms() []PaintRoomInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	rooms := make([]PaintRoomInfo, 0, len(p.rooms))
	for _, room := range p.rooms {
		rooms = append(rooms, room.Info())
	}
	return rooms
}

func (p *Paint) AssignRequestToRoom(wreq *ward.Request, roomName string) error {
	room := p.GetRoom(roomName)
	if room == nil {
		return fmt.Errorf("room not found: %s", roomName)
	}
	return p.ws.AssignRequestToChannel(wreq, room.Channel)
}
