package paint

import (
	"context"
	"encoding/json"
	"fmt"
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
	lastLeftAt  time.Time
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

func (p *PaintRoom) LogInfo() string {
	return fmt.Sprintf("%s Room %s | ", p.Channel.LogInfo(), p.cfg.Name)
}

func (p *PaintRoom) Log(v ...any) {
	args := append([]any{p.LogInfo()}, v...)
	log.Println(args...)
}

func (p *PaintRoom) Run(pm *Paint, ctx context.Context) {
	go func() {
		p.Log("is listening")
		streamTicker := time.NewTicker(100 * time.Millisecond)
		saveTicker := time.NewTicker(5000 * time.Millisecond)
		syncTicker := time.NewTicker(1000 * 10 * time.Millisecond)
		defer func() {
			streamTicker.Stop()
			saveTicker.Stop()
			syncTicker.Stop()

			pm.DeleteRoom(p.cfg.Name)
		}()
		for {
			select {
			case <-streamTicker.C:
				go func() {
					p.cMu.Lock()
					if len(p.streamBuf) > 0 {
						event := wardsocket.ByteEvent{
							Type:    "canvas_pixel_update",
							Payload: p.streamBuf,
						}

						data, err := json.Marshal(event)
						if err == nil {
							p.Channel.BroadcastStream(data)
						}
						p.streamBuf = p.streamBuf[:0]
					}
					p.cMu.Unlock()
				}()
			case <-saveTicker.C:
				go func() {
					p.cMu.Lock()
					defer p.cMu.Unlock()
					if len(p.saveBuf) == 0 {
						return
					}
					event := wardsocket.ByteEvent{
						Type:    "canvas_pixel_update",
						Payload: p.saveBuf,
					}

					data, err := json.Marshal(event)
					if err == nil {
						p.Channel.BroadcastStream(data)
					}

					p.saveCanvasBytes()
				}()
			case <-syncTicker.C:
				if p.cfg.IsTemporary {
					p.closeIfEmpty()
				}

			case <-ctx.Done():
				log.Println("context done")
				return
			}

		}
	}()
}
func (p *PaintRoom) closeIfEmpty() {
	if !p.isEmpty() {
		return
	}
	if time.Now().Before(p.lastLeftAt.Add(30 * time.Second)) {
		return
	}
	p.Channel.Close()
}

func (p *PaintRoom) isEmpty() bool {
	p.uMu.RLock()
	defer p.uMu.RUnlock()
	for _, u := range p.users {
		if u.IsConnected {
			return false
		}
	}
	return true
}

func (p *PaintRoom) sync() {
	start := time.Now()
	if p.isEmpty() {
		return
	}
	p.cMu.Lock()
	payload := p.getCanvasBytes()
	event := wardsocket.ByteEvent{
		Type:    "canvas_pixel_update",
		Payload: payload,
	}

	data, err := json.Marshal(event)
	if err == nil {
		p.Channel.BroadcastStream(data)
	}
	p.cMu.Unlock()

	p.Log("paint sync ticker: ", time.Since(start))
}
