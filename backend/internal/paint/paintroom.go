package paint

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"sync"
	"time"

	"paint.pecet.it/pkg/wardsocket"
)

type PaintRoom struct {
	Channel *wardsocket.Channel
	cfg     *RoomConfig
	Canvas  *image.RGBA
	cMu     sync.RWMutex
	saveBuf []byte

	uMu         sync.RWMutex
	users       map[string]*User
	chatHistory []ChatMessage
	lastLeftAt  time.Time
}

func newPaintRoom(cfg *RoomConfig, channel *wardsocket.Channel) *PaintRoom {
	p := &PaintRoom{
		cfg:         cfg,
		Channel:     channel,
		Canvas:      image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height)),
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
	p.Channel.RegisterEventHandler("canvas_get_all", p.handleGetAllCanvas)
	p.Channel.RegisterEventHandler("chat_message", p.handleChatMessage)
	p.Channel.RegisterEventHandler("webrtc_signal", p.handleSignal)

	p.Channel.RegisterEventHandler("user_kick", p.handleUserKick)
	p.Channel.RegisterEventHandler("user_operator", p.handleUserOperator)
	p.Channel.RegisterEventHandler("user_draw", p.handleUserDraw)
}
func (p *PaintRoom) Info() RoomInfo {
	p.uMu.RLock()
	defer p.uMu.RUnlock()

	onlineUsers := 0
	for _, u := range p.users {
		if u.IsConnected {
			onlineUsers++
		}
	}

	return RoomInfo{
		Name:        p.cfg.Name,
		IsTemporary: p.cfg.IsTemporary,
		OnlineUsers: onlineUsers,
		IsPassword:  p.cfg.Password != "",
		Width:       p.cfg.Width,
		Height:      p.cfg.Height,
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
		p.Log("running")
		saveTicker := time.NewTicker(1000 * 3 * time.Millisecond)
		syncTicker := time.NewTicker(1000 * 20 * time.Millisecond)
		defer func() {
			saveTicker.Stop()
			syncTicker.Stop()

			pm.DeleteRoom(p.cfg.Name)
		}()
		for {
			select {
			case <-saveTicker.C:
				go func() {
					p.cMu.Lock()
					defer p.cMu.Unlock()
					if len(p.saveBuf) == 0 {
						return
					}
					p.saveCanvasBytes()
				}()
			case <-syncTicker.C:
				if p.cfg.IsTemporary {
					go p.closeIfEmpty()
				}

			case <-ctx.Done():
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
		p.Channel.Broadcast(data)
	}
	p.cMu.Unlock()

	p.Log("paint sync ticker: ", time.Since(start))
}
