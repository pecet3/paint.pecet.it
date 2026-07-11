package paint

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/wardsocket"
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

// COMMENT: asa
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

type RoomUser struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	IsOperator  bool   `json:"is_operator"`
	IsConnected bool   `json:"is_connected"`
	IsDrawing   bool   `json:"is_drawing"`
	IsKicked    bool   `json:"is_kicked"`
}

type SignalPayload struct {
	TargetUUID string          `json:"targetUuid"`
	SenderUUID string          `json:"senderUuid"`
	SignalType string          `json:"signalType"`
	Data       json.RawMessage `json:"data"`
}

type UserManagmentPayload struct {
	Uuid string `json:"uuid"`
}

func (p *PaintRoom) handleUserKick(ctx context.Context, event *wardsocket.Event) {
	p.uMu.Lock()
	defer p.uMu.Unlock()
	p.Log(event)
	eventUser, ok := p.users[event.Client.Request.User.Uuid()]
	if !ok {
		p.Log("user doesn't belong to room requested user managment event ", event.Client.Request.User.Uuid())
		return
	}

	if !eventUser.IsOperator {
		p.Log("no operator requested user managment event ", event.Client.Request.User.Uuid())
		return
	}
	var payload UserManagmentPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		p.Log(err)

		return
	}

	log.Println(payload)
	manageUser, ok := p.users[payload.Uuid]
	if !ok {
		p.Log("user to manage doesn't belong to room", payload.Uuid)
		return
	}
	manageUser.IsKicked = !manageUser.IsKicked

	p.BroadcastUserList()
}

func (p *PaintRoom) handleGetAllCanvas(ctx context.Context, event *wardsocket.Event) {
	p.cMu.RLock()
	canvasEvent := wardsocket.ByteEvent{
		Type:    "canvas_pixel_update",
		Payload: append(p.getCanvasBytes(), p.saveBuf...),
	}
	p.cMu.RUnlock()

	if data, err := json.Marshal(canvasEvent); err == nil {
		event.Client.Send(data)
	}
}

func (p *PaintRoom) handleJoin(ctx context.Context, c *wardsocket.Client) {
	p.Log(c.Request.User.Uuid(), "joined")

	p.uMu.Lock()
	defer p.uMu.Unlock()
	uuid := c.Request.User.Uuid()
	user, exists := p.users[uuid]

	if !exists {
		user = &User{
			WardUser:      c.Request.User,
			IsOperator:    len(p.users) == 0,
			IsConnected:   true,
			JoinedAt:      time.Now(),
			IsAbleDrawing: true,
		}
		p.users[uuid] = user
	} else {
		user.IsConnected = true

	}

	p.SendChatHistory(c)
	p.BroadcastUserList()
	p.BroadcastServerMessage(c.Request.User.Name() + " joined")

	outgoingEvent := []byte(`{"type":"join_confirmation","payload":""}`)
	c.Send(outgoingEvent)
}

func (p *PaintRoom) handleLeave(ctx context.Context, client *wardsocket.Client) {
	p.Log(client.Request.LogInfo(), "left")
	p.uMu.Lock()
	defer p.uMu.Unlock()
	uuid := client.Request.User.Uuid()
	if user, exists := p.users[uuid]; exists {
		user.IsConnected = false
	}

	p.lastLeftAt = time.Now()

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
	out := []byte(`{"type":"canvas_pixel_update","payload":` + string(evt.Payload) + `}`)
	p.Channel.Broadcast(out, evt.Client)

	p.cMu.Lock()
	defer p.cMu.Unlock()
	p.saveBuf = append(p.saveBuf, data...)
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
