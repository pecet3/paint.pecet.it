package wsmanager

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	Ident string

	clients map[*Client]bool
	cMu     sync.RWMutex

	broadcastCh chan json.RawMessage
	joinCh      chan *Client
	leaveCh     chan *Client
	eventCh     chan *Event

	eventHandlers map[string]EventHandler
	joinHandlers  []EntranceEventHandler
	leaveHandlers []EntranceEventHandler
}

func NewRoom(ident string) *Room {
	return &Room{
		Ident:       ident,
		clients:     make(map[*Client]bool),
		broadcastCh: make(chan json.RawMessage),
		joinCh:      make(chan *Client),
		leaveCh:     make(chan *Client),
		eventCh:     make(chan *Event),

		eventHandlers: make(map[string]EventHandler),
		joinHandlers:  []EntranceEventHandler{},
		leaveHandlers: []EntranceEventHandler{},
	}
}

func (r *Room) Broadcast(msg json.RawMessage) {
	r.broadcastCh <- msg
}

func (r *Room) HandleNewConn(conn *websocket.Conn) {
	client := NewClient(conn)
	r.joinCh <- client

	go client.writePump()
	go client.readPump(r)
}

func (r *Room) RegisterEventHandler(eventType string, handler EventHandler) {
	r.eventHandlers[eventType] = handler
}

func (r *Room) RegisterJoinHandler(handler func(client *Client)) {
	r.joinHandlers = append(r.joinHandlers, handler)
}

func (r *Room) RegisterLeaveHandler(handler func(client *Client)) {
	r.leaveHandlers = append(r.leaveHandlers, handler)
}

func (r *Room) Run() {
	log.Printf(" %s room is listening ", r.Ident)
	for {
		select {
		case client := <-r.joinCh:
			if len(r.joinHandlers) > 0 {
				for _, handle := range r.joinHandlers {
					handle(client)
				}
			}
		case client := <-r.leaveCh:
			if len(r.joinHandlers) > 0 {
				for _, handle := range r.leaveHandlers {
					handle(client)
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
			log.Printf(" %s", msg.Type)

			if handler, ok := r.eventHandlers[msg.Type]; ok {
				handler(msg)
			} else {
				log.Printf("Unhandled event type: %s", msg.Type)
			}
		}
	}
}
