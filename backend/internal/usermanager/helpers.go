package usermanager

import (
	"encoding/json"
	"time"

	"paint.pecet.it/pkg/ward/wardsocket"
)

func (m *Manager) broadcastEvent(eventType string, payload any) {
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
	m.room.Log(eventType, payload)
	m.room.Broadcast(evtData)
}

func (m *Manager) BroadcastServerMessage(msg string) {
	serverMsg := ServerMessage{
		Message: msg,
		Date:    time.Now(),
	}
	m.broadcastEvent("server_message", serverMsg)
}
func (m *Manager) SendChatHistory(client *wardsocket.Client) {
	for _, msg := range m.chatHistory {
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
func (m *Manager) BroadcastUserList() {
	var list []map[string]any
	for _, u := range m.users {
		list = append(list, map[string]any{
			"uuid":                 u.WardUser.Uuid(),
			"name":                 u.WardUser.Name(),
			"is_operator":          u.IsOperator,
			"is_connected":         u.IsConnected,
			"ban_duration_seconds": int64(u.BanDuration.Seconds()),
		})
	}
	m.broadcastEvent("update_users_list", list)
}
