package paint

import (
	"encoding/json"
	"time"

	"paint.pecet.it/pkg/ward/wardsocket"
)

func (p *PaintRoom) broadcastEvent(eventType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	evt := wardsocket.Event{
		Type:    eventType,
		Payload: data,
	}
	evtData, err := json.Marshal(evt)
	if err != nil {
		return
	}
	p.Channel.Log(eventType, payload)
	p.Channel.Broadcast(evtData)
}

func (p *PaintRoom) BroadcastServerMessage(msg string) {
	serverMsg := ServerMessage{
		Message: msg,
		Date:    time.Now(),
	}
	p.broadcastEvent("server_message", serverMsg)
}
func (p *PaintRoom) SendChatHistory(client *wardsocket.Client) {
	for _, msg := range p.chatHistory {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		evt := wardsocket.Event{
			Type:    "chat_message",
			Payload: data,
		}
		evtData, err := json.Marshal(evt)
		if err != nil {
			continue
		}
		client.Send(evtData)
	}
}
func (p *PaintRoom) BroadcastUserList() {
	var list []RoomUserEvt
	for _, u := range p.users {
		list = append(list, RoomUserEvt{
			UUID:          u.WardUser.Uuid(),
			Name:          u.WardUser.Name(),
			IsOperator:    u.IsOperator,
			IsConnected:   u.IsConnected,
			IsAbleDrawing: u.IsAbleDrawing,
		})

	}
	p.broadcastEvent("update_users_list", list)
}
