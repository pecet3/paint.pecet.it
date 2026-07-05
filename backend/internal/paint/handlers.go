package paint

import (
	"context"
	"encoding/json"
	"time"

	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

type User struct {
	WardUser      ward.User
	IsOperator    bool
	IsKicked      bool
	IsConnected   bool
	IsAbleDrawing bool
	LastChatMsgAt time.Time
	JoinedAt      time.Time
}
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
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	IsOperator    bool   `json:"is_operator"`
	IsConnected   bool   `json:"is_connected"`
	IsAbleDrawing bool   `json:"is_able_drawing"`
}

type SignalPayload struct {
	TargetUUID string          `json:"targetUuid"`
	SenderUUID string          `json:"senderUuid"`
	SignalType string          `json:"signalType"`
	Data       json.RawMessage `json:"data"`
}

func (p *PaintRoom) handleJoin(ctx context.Context, c *wardsocket.Client) {
	p.cMu.Lock()
	p.saveCanvasBytes()
	canvasEvent := wardsocket.ByteEvent{
		Type:    "canvas_pixel_update",
		Payload: p.getCanvasBytes(),
	}
	p.cMu.Unlock()
	if data, err := json.Marshal(canvasEvent); err == nil {
		c.Send(data)
	}

	p.uMu.Lock()

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
			WardUser:      c.Request.User,
			IsOperator:    shouldBeOperator,
			IsConnected:   true,
			JoinedAt:      time.Now(),
			IsAbleDrawing: true,
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
	p.uMu.Unlock()
	p.BroadcastServerMessage(c.Request.User.Name() + " joined")

}

func (p *PaintRoom) handleLeave(ctx context.Context, client *wardsocket.Client) {
	p.uMu.Lock()
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
	p.lastLeftAt = time.Now()
	p.uMu.Unlock()

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
	user, exists := p.users[uuid]
	if !exists {

		return
	}
	now := time.Now()

	if now.Before(user.LastChatMsgAt.Add(time.Millisecond * 2000)) {
		user.LastChatMsgAt = now
		return
	}
	user.LastChatMsgAt = now

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

func (p *PaintRoom) handleSignal(ctx context.Context, e *wardsocket.Event) {
	var payload SignalPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		e.Client.Request.Log("Invalid WebRTC signal payload:", err)
		return
	}

	payload.SenderUUID = e.Client.Request.User.Uuid()

	outgoingPayloadBytes, err := json.Marshal(payload)
	if err != nil {
		e.Client.Request.Log("Failed to marshal WebRTC payload:", err)
		return
	}

	outgoingEvent := []byte(`{"type":"webrtc_signal","payload":` + string(outgoingPayloadBytes) + `}`)
	p.Channel.Broadcast(outgoingEvent)
}
