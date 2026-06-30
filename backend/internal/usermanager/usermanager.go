package usermanager

import (
	"context"
	"sync"
	"time"

	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

type contextKey string

const ManagerContextKey contextKey = "usermanager"

type RoomUser struct {
	WardUser    ward.User
	IsOperator  bool
	BanDuration time.Duration
	IsConnected bool
	Values      map[string]any
	vMu         sync.RWMutex
}

func (ru *RoomUser) SetValue(key string, value any) {
	ru.vMu.Lock()
	defer ru.vMu.Unlock()
	ru.Values[key] = value
}

func (ru *RoomUser) GetValue(key string) (any, bool) {
	ru.vMu.RLock()
	defer ru.vMu.RUnlock()
	val, ok := ru.Values[key]
	return val, ok
}

type Manager struct {
	mu          sync.RWMutex
	users       map[string]*RoomUser
	room        *wardsocket.Room
	chatHistory []ChatMessage
}

func New(room *wardsocket.Room) *Manager {
	m := &Manager{
		users:       make(map[string]*RoomUser),
		room:        room,
		chatHistory: make([]ChatMessage, 0, 64),
	}

	return m
}

func (m *Manager) RegisterHandlers() {
	m.room.RegisterJoinHandler(m.handleJoin)
	m.room.RegisterLeaveHandler(m.handleLeave)
	m.room.RegisterEventHandler("chat_message", m.handleChatMessage)
	m.room.RegisterEventHandler("update_is_operator", m.handleUpdateOperator)
	m.room.RegisterEventHandler("update_ban_duration", m.handleUpdateBanDuration)
}

func (m *Manager) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ManagerContextKey, m)
}

func FromContext(ctx context.Context) (*Manager, bool) {
	m, ok := ctx.Value(ManagerContextKey).(*Manager)
	return m, ok
}

func (m *Manager) GetRoomUserFromClient(client *wardsocket.Client) (*RoomUser, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uuid := client.Request.User.Uuid()
	user, exists := m.users[uuid]
	return user, exists
}
