package usermanager

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

func (m *Manager) handleJoin(ctx context.Context, client *wardsocket.Client) {
	m.mu.Lock()
	uuid := client.Request.User.Uuid()
	user, exists := m.users[uuid]

	operatorCount := 0
	for _, u := range m.users {
		if u.IsConnected && u.IsOperator {
			operatorCount++
		}
	}

	shouldBeOperator := operatorCount == 0

	if !exists {
		user = &RoomUser{
			WardUser:    client.Request.User,
			IsOperator:  shouldBeOperator,
			IsConnected: true,
			Values:      make(map[string]any),
		}
		m.users[uuid] = user
	} else {
		user.IsConnected = true
		if shouldBeOperator {
			user.IsOperator = true
		}
	}
	m.mu.Unlock()

	m.SendChatHistory(client)

	serverMsg := ServerMessage{
		Message: client.Request.User.Name() + " joined",
		Date:    time.Now(),
	}
	m.broadcastEvent("server_message", serverMsg)
}

func (m *Manager) handleLeave(ctx context.Context, client *wardsocket.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	uuid := client.Request.User.Uuid()
	if user, exists := m.users[uuid]; exists {
		user.IsConnected = false

		if user.IsOperator {
			assigned := false
			for _, u := range m.users {
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

	m.BroadcastServerMessage(client.Request.User.Name() + " left")

	m.BroadcastUserList()
}

func (m *Manager) handleChatMessage(ctx context.Context, event *wardsocket.Event) {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	uuid := event.Client.Request.User.Uuid()
	if user, exists := m.users[uuid]; exists && user.BanDuration > 0 {
		return
	}

	msg := ChatMessage{
		Name:    event.Client.Request.User.Name(),
		Uuid:    uuid,
		Message: payload.Message,
		Date:    time.Now(),
	}

	m.chatHistory = append(m.chatHistory, msg)
	if len(m.chatHistory) > 64 {
		m.chatHistory = m.chatHistory[1:]
	}

	m.broadcastEvent("chat_message", msg)
}

func (m *Manager) handleUpdateOperator(ctx context.Context, event *wardsocket.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	senderUuid := event.Client.Request.User.Uuid()
	sender, exists := m.users[senderUuid]
	if !exists || !sender.IsOperator {
		return
	}

	var payload struct {
		TargetUuid string `json:"target_uuid"`
		IsOperator bool   `json:"is_operator"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	if target, ok := m.users[payload.TargetUuid]; ok {
		target.IsOperator = payload.IsOperator
		m.broadcastEvent("update_is_operator", payload)
	}
}

func (m *Manager) handleUpdateBanDuration(ctx context.Context, event *wardsocket.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	senderUuid := event.Client.Request.User.Uuid()
	sender, exists := m.users[senderUuid]
	if !exists || !sender.IsOperator {
		return
	}

	var payload struct {
		TargetUuid      string `json:"target_uuid"`
		DurationSeconds int64  `json:"duration_seconds"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	if target, ok := m.users[payload.TargetUuid]; ok {
		target.BanDuration = time.Duration(payload.DurationSeconds) * time.Second
		m.broadcastEvent("update_ban_duration", payload)
	}
}

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
