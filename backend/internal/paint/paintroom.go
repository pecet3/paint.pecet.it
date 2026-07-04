package paint

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

type PaintRoomInfo struct {
	Name        string `json:"name"`
	IsTemporary bool   `json:"is_temporary"`
	OnlineUsers int    `json:"online_users"`
	IsPassword  bool   `json:"is_password"`
}
type RoomConfig struct {
	Name        string `json:"name"`
	IsTemporary bool   `json:"is_temporary"`
	Password    string `json:"password"`
}
type PaintRoom struct {
	Channel   *wardsocket.Channel
	cfg       *RoomConfig
	Canvas    *image.RGBA
	cMu       sync.RWMutex
	streamBuf []byte
	saveBuf   []byte

	uMu         sync.RWMutex
	users       map[string]*User
	chatHistory []ChatMessage
}

func newPaintRoom(cfg *RoomConfig, channel *wardsocket.Channel) *PaintRoom {
	p := &PaintRoom{
		cfg:         cfg,
		Channel:     channel,
		Canvas:      image.NewRGBA(image.Rect(0, 0, width, height)),
		streamBuf:   make([]byte, 0),
		saveBuf:     make([]byte, 0),
		users:       make(map[string]*User),
		chatHistory: make([]ChatMessage, 0, 64),
	}
	return p
}
func (p *PaintRoom) RegisterHandlers() {
	p.Channel.RegisterJoinHandler(p.handleJoin)
	p.Channel.RegisterLeaveHandler(p.handleLeave)

	p.Channel.RegisterEventHandler("canvas_pixel_update", p.handlePixelUpdate)
	p.Channel.RegisterEventHandler("chat_message", p.handleChatMessage)
	p.Channel.RegisterEventHandler("webrtc_signal", p.handleSignal)

}
func (p *PaintRoom) Info() PaintRoomInfo {
	p.uMu.RLock()
	defer p.uMu.RUnlock()

	onlineUsers := 0
	for _, u := range p.users {
		if u.IsConnected {
			onlineUsers++
		}
	}

	return PaintRoomInfo{
		Name:        p.cfg.Name,
		IsTemporary: p.cfg.IsTemporary,
		OnlineUsers: onlineUsers,
		IsPassword:  p.cfg.Password != "",
	}
}

func (p *PaintRoom) Run(ctx context.Context) {
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
					p.cMu.Lock()
					if len(p.streamBuf) > 0 {
						event := wardsocket.Event{
							Type:    "canvas_pixel_update",
							Payload: p.streamBuf,
						}

						data, err := json.Marshal(event)
						if err == nil {
							p.Channel.Broadcast(data)
						}
						p.streamBuf = p.streamBuf[:0]
					}
					p.cMu.Unlock()
				}()
			case <-saveTicker.C:
				go func() {
					p.Channel.Log("paint save ticker")
					if len(p.saveBuf) == 0 {
						return
					}
					start := time.Now()
					p.cMu.Lock()
					p.saveCanvasBytes()
					p.cMu.Unlock()
					log.Println(time.Since(start))
				}()
			case <-syncTicker.C:
				go func() {
					p.Channel.Log("paint sync ticker")

					start := time.Now()

					p.cMu.Lock()
					payload := p.getCanvasBytes()
					event := wardsocket.Event{
						Type:    "canvas_pixel_update",
						Payload: payload,
					}

					data, err := json.Marshal(event)
					if err == nil {
						p.Channel.Broadcast(data)
					}
					p.cMu.Unlock()

					log.Println(time.Since(start))

				}()
			case <-ctx.Done():
				return
			}

		}
	}()
}
