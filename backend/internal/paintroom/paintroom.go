package paintroom

import (
	"context"
	"encoding/json"
	"image"
	"log"
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

	streamBuf []byte
	saveBuf   []byte
}

func New(room *wardsocket.Room) *Paint {
	p := &Paint{
		Room:      room,
		Canvas:    image.NewRGBA(image.Rect(0, 0, width, height)),
		streamBuf: make([]byte, 0),
		saveBuf:   make([]byte, 0),
	}
	return p
}
func (p *Paint) RegisterHandlers() {
	p.Room.RegisterEventHandler("canvas_pixel_update", p.handlePixelUpdate)
	p.Room.RegisterJoinHandler(p.handleJoinEvent)
}

func (p *Paint) Run(ctx context.Context) {
	go func() {
		streamTicker := time.NewTicker(30 * time.Millisecond)
		saveTicker := time.NewTicker(2000 * time.Millisecond)
		syncTicker := time.NewTicker(10000 * time.Millisecond)
		defer func() {
			streamTicker.Stop()
			saveTicker.Stop()
			syncTicker.Stop()
		}()
		for {
			select {
			case <-streamTicker.C:
				go func() {
					p.mu.Lock()
					if len(p.streamBuf) > 0 {
						event := Event{
							Type:    "canvas_pixel_update",
							Payload: p.streamBuf,
						}

						data, err := json.Marshal(event)
						if err == nil {
							p.Room.Broadcast(data)
						}
						p.streamBuf = p.streamBuf[:0]
					}
					p.mu.Unlock()
				}()
			case <-saveTicker.C:
				go func() {
					p.Room.Log("paint save ticker")
					if len(p.saveBuf) == 0 {
						return
					}
					start := time.Now()
					p.mu.Lock()
					p.saveCanvasBytes()
					p.mu.Unlock()
					log.Println(time.Since(start))
				}()
			case <-syncTicker.C:
				go func() {
					p.Room.Log("paint sync ticker")

					start := time.Now()

					p.mu.Lock()
					payload := p.getCanvasBytes()
					event := Event{
						Type:    "canvas_pixel_update",
						Payload: payload,
					}

					data, err := json.Marshal(event)
					if err == nil {
						p.Room.Broadcast(data)
					}
					p.mu.Unlock()

					log.Println(time.Since(start))

				}()
			case <-ctx.Done():
				return
			}

		}
	}()
}
