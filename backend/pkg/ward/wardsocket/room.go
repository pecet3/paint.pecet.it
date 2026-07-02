package wardsocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Client  *Client
}

type Room struct {
	Ident string

	clients map[*Client]bool
	cMu     sync.RWMutex

	broadcastCh chan json.RawMessage
	joinCh      chan *Client
	leaveCh     chan *Client
	eventCh     chan *Event

	eventHandlers map[string]func(context.Context, *Event)

	joinHandlers  []func(context.Context, *Client)
	leaveHandlers []func(context.Context, *Client)

	cancel  context.CancelFunc
	closeCh chan struct{}
}

func NewRoom(ident string) *Room {
	r := &Room{
		Ident:       ident,
		clients:     make(map[*Client]bool),
		broadcastCh: make(chan json.RawMessage),
		joinCh:      make(chan *Client),
		leaveCh:     make(chan *Client),
		eventCh:     make(chan *Event),

		eventHandlers: make(map[string]func(context.Context, *Event)),
		joinHandlers:  []func(context.Context, *Client){},
		leaveHandlers: []func(context.Context, *Client){},

		closeCh: make(chan struct{}),
	}
	return r
}

func (r *Room) WithCancelContext(ctx context.Context) (*Room, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	return r, ctx
}

func (r *Room) Broadcast(msg json.RawMessage) {
	r.broadcastCh <- msg
}

func (r *Room) LeaveClient(client *Client) {
	r.leaveCh <- client
}

func (r *Room) JoinConn(conn *websocket.Conn, wreq *ward.Request) {
	client := NewClient(r, conn, wreq)

	r.joinCh <- client

	go client.writePump(r)
	go client.readPump(r)
}

func (r *Room) RegisterEventHandler(eventType string, handler func(context.Context, *Event)) {
	r.eventHandlers[eventType] = handler
}

func (r *Room) RegisterJoinHandler(handler func(context.Context, *Client)) {
	r.joinHandlers = append(r.joinHandlers, handler)
}

func (r *Room) RegisterLeaveHandler(handler func(context.Context, *Client)) {
	r.leaveHandlers = append(r.leaveHandlers, handler)
}

func (r *Room) Close() {
	<-r.closeCh
}

func (r *Room) LogInfo() string {
	return fmt.Sprintf("room: %s", r.Ident)
}

func (r *Room) Log(v ...any) {
	args := append([]any{r.LogInfo()}, v...)
	log.Println(args...)
}

func (r *Room) Run(ctx context.Context) {
	go func() {
		r.Log("is listening")
		for {
			select {
			case client := <-r.joinCh:
				r.clients[client] = true
				if len(r.joinHandlers) > 0 {
					for _, handle := range r.joinHandlers {
						go handle(ctx, client)
					}
				}
			case client := <-r.leaveCh:
				if len(r.leaveHandlers) > 0 {
					for _, handle := range r.leaveHandlers {
						go handle(ctx, client)
					}
				}
				r.cMu.Lock()
				if _, ok := r.clients[client]; ok {
					delete(r.clients, client)
					close(client.sendCh)
				}
				r.cMu.Unlock()

			case msg := <-r.broadcastCh:
				r.cMu.RLock()
				for client := range r.clients {
					client.sendCh <- msg
				}
				r.cMu.RUnlock()

			case msg := <-r.eventCh:
				r.Log(msg.Type)
				if handler, ok := r.eventHandlers[msg.Type]; ok {
					go handler(ctx, msg)
				} else {
					r.Log("Unhandled event type: %s", msg.Type)
				}
			case <-ctx.Done():
				log.Println("closing room done")
			case <-r.closeCh:
				r.Log("closing room")
				for client := range r.clients {
					close(client.sendCh)
					delete(r.clients, client)
				}
				r.Log("removed all clients")
				if r.cancel != nil {
					r.Log("executing context cancel func")
					r.cancel()
				}

			}
		}
	}()
}
