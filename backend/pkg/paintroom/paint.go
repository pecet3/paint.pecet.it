package paintroom

import (
	"context"
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

func (p *Paint) Run(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.mu.Lock()
				if len(p.pixelFrameBuf) > 0 {
					event := Event{
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
			case <-ctx.Done():
				return
			}

		}
	}()
}
