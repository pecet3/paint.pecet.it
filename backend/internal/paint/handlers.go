package paint

import (
	"context"
	"encoding/json"
	"time"

	"paint.pecet.it/pkg/ward/wardsocket"
)

type ChatMessage struct {
	Name    string    `json:"name"`
	Uuid    string    `json:"uuid"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

type ServerMessage struct {
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

type RoomUserEvt struct {
	UUID               string `json:"uuid"`
	Name               string `json:"name"`
	IsOperator         bool   `json:"is_operator"`
	IsConnected        bool   `json:"is_connected"`
	BanDurationSeconds int64  `json:"ban_duration_seconds"`
}

func (p *PaintRoom) handleJoin(ctx context.Context, c *wardsocket.Client) {
	p.cMu.Lock()
	p.saveCanvasBytes()
	canvasEvent := wardsocket.Event{
		Type:    "canvas_pixel_update",
		Payload: p.getCanvasBytes(),
	}
	p.cMu.Unlock()
	if data, err := json.Marshal(canvasEvent); err == nil {
		c.Send(data)
	}

	p.uMu.Lock()
	defer p.uMu.Unlock()
	uuid := c.Request.User.Uuid()
	user, exists := p.users[uuid]

	operatorCount := 0
	for _, u := range p.users {
		if u.IsConnected && u.IsOperator {
			operatorCount++
		}
	}

	shouldBeOperator := operatorCount == 0

	if !exists {
		user = &User{
			WardUser:    c.Request.User,
			IsOperator:  shouldBeOperator,
			IsConnected: true,
		}
		p.users[uuid] = user
	} else {
		user.IsConnected = true
		if shouldBeOperator {
			user.IsOperator = true
		}
	}

	p.SendChatHistory(c)
	p.BroadcastUserList()
	p.BroadcastServerMessage(c.Request.User.Name() + " joined")

}

func (p *PaintRoom) handleLeave(ctx context.Context, client *wardsocket.Client) {
	p.uMu.Lock()
	defer p.uMu.Unlock()

	uuid := client.Request.User.Uuid()
	if user, exists := p.users[uuid]; exists {
		user.IsConnected = false

		if user.IsOperator {
			assigned := false
			for _, u := range p.users {
				if u.IsConnected && u.WardUser.Uuid() != uuid {
					u.IsOperator = true
					assigned = true
					break
				}
			}
			if assigned {
				user.IsOperator = false
			}
		}
	}

	p.BroadcastServerMessage(client.Request.User.Name() + " left")

	p.BroadcastUserList()
}

func (p *PaintRoom) handleChatMessage(ctx context.Context, event *wardsocket.Event) {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	p.uMu.Lock()
	defer p.uMu.Unlock()

	uuid := event.Client.Request.User.Uuid()
	if user, exists := p.users[uuid]; exists && user.BanDuration > 0 {
		return
	}

	msg := ChatMessage{
		Name:    event.Client.Request.User.Name(),
		Uuid:    uuid,
		Message: payload.Message,
		Date:    time.Now(),
	}

	p.chatHistory = append(p.chatHistory, msg)
	if len(p.chatHistory) > 64 {
		p.chatHistory = p.chatHistory[1:]
	}

	p.broadcastEvent("chat_message", msg)
}

func (p *PaintRoom) handlePixelUpdate(ctx context.Context, evt *wardsocket.Event) {
	var data []byte
	err := json.Unmarshal(evt.Payload, &data)
	if err != nil {
		evt.Client.Request.Log("unmarshal err: ", err)
		return
	}
	if len(data) == 0 || len(data)%8 != 0 {
		return
	}
	p.cMu.Lock()
	defer p.cMu.Unlock()
	p.streamBuf = append(p.streamBuf, data...)
	p.saveBuf = append(p.saveBuf, p.streamBuf...)
}
