package paintroom

import (
	"encoding/json"
	"image"
	"sync"
	"time"

	"paint.pecet.it/pkg/ward/wardsocket"
)

const (
	width  = 800
	height = 600
)

type UpdateEvent struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type Paint struct {
	Room   *wardsocket.Room
	Canvas *image.RGBA
	mu     sync.Mutex

	pixelFrameBuf []byte
}

func New(room *wardsocket.Room) *Paint {
	p := &Paint{
		Room:          room,
		Canvas:        image.NewRGBA(image.Rect(0, 0, width, height)),
		pixelFrameBuf: make([]byte, 0),
	}
	p.Room.RegisterEventHandler("canvas_pixel_update", p.handlePixelUpdate)
	p.Room.RegisterJoinHandler(p.handleJoinEvent)
	return p
}

func (p *Paint) handleJoinEvent(c *wardsocket.Client) {
	c.Log("joined to room")

	event := UpdateEvent{
		Type:    "canvas_pixel_update",
		Payload: p.getAllPixelFrames(),
	}

	data, err := json.Marshal(event)
	if err == nil {
		c.Send(data)
	}
}

func (p *Paint) handlePixelUpdate(evt *wardsocket.Event) {
	var data []byte
	err := json.Unmarshal(evt.Payload, &data)
	if err != nil {
		evt.Client.Log("unmarshal err: ", err)
		return
	}
	if len(data) == 0 || len(data)%8 != 0 {
		return
	}
	p.setPixelFramesToBuffers(data)
}

func (p *Paint) Run() {
	go func() {
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			p.mu.Lock()
			if len(p.pixelFrameBuf) > 0 {
				event := UpdateEvent{
					Type:    "canvas_pixel_update",
					Payload: p.pixelFrameBuf,
				}

				data, err := json.Marshal(event)
				if err == nil {
					p.Room.Broadcast(data)
				}

				p.pixelFrameBuf = p.pixelFrameBuf[:0]
			}

			p.mu.Unlock()
		}
	}()
}
