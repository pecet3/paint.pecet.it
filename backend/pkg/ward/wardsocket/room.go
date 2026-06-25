package wardsocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

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

func (r *Room) WithContext() (*Room, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	return r, ctx
}

func (r *Room) Broadcast(msg json.RawMessage) {
	r.broadcastCh <- msg
}

func (r *Room) LeaveClient(client *Client) {
	r.leaveCh <- client
}

func (r *Room) HandleNewClient(conn *websocket.Conn, wreq *ward.Request) {
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
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Room) Run(ctx context.Context) {
	go func() {
		log.Printf("room: %s is listening ", r.Ident)
		for {
			select {
			case client := <-r.joinCh:
				r.clients[client] = true
				if len(r.joinHandlers) > 0 {
					for _, handle := range r.joinHandlers {
						handle(ctx, client)
					}
				}
			case client := <-r.leaveCh:
				if len(r.leaveHandlers) > 0 {
					for _, handle := range r.leaveHandlers {
						handle(ctx, client)
					}
				}
				if _, ok := r.clients[client]; ok {
					delete(r.clients, client)
					close(client.sendCh)
				}

			case msg := <-r.broadcastCh:
				for client := range r.clients {
					select {
					case client.sendCh <- msg:
					default:
						close(client.sendCh)
						delete(r.clients, client)
					}
				}
			case msg := <-r.eventCh:
				if handler, ok := r.eventHandlers[msg.Type]; ok {
					go handler(ctx, msg)
				} else {
					log.Printf("Unhandled event type: %s", msg.Type)
				}
			case <-ctx.Done():
				log.Println("closing room done")
			case <-r.closeCh:
				log.Println("closing room")
				if r.cancel != nil {
					r.cancel()
				}

			}
		}
	}()
}
