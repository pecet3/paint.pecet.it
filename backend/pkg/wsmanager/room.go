package wsmanager

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	clients map[*Client]bool
	cMu     sync.RWMutex

	broadcastCh chan json.RawMessage
	joinCh      chan *Client
	leaveCh     chan *Client
	eventCh     chan *Event

	events map[string]EventHandler
}

func NewRoom() *Room {
	return &Room{
		clients:     make(map[*Client]bool),
		broadcastCh: make(chan json.RawMessage),
		joinCh:      make(chan *Client),
		leaveCh:     make(chan *Client),
		eventCh:     make(chan *Event),
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
	if r.events == nil {
		r.events = make(map[string]EventHandler)
	}
	r.events[eventType] = handler
}

func (r *Room) Run() {
	log.Println("room listening ")
	for {
		select {
		case client := <-r.joinCh:
			log.Println(client, " join")
			r.clients[client] = true
		case client := <-r.leaveCh:
			log.Println(client, " leave")
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.sendCh)
			}

		case msg := <-r.broadcastCh:
			for client := range r.clients {
				select {
				case client.sendCh <- msg:
				default:
					log.Println(22)
					close(client.sendCh)
					delete(r.clients, client)
				}
			}
		case msg := <-r.eventCh:
			log.Printf("aaa %s", msg.Type)

			if handler, ok := r.events[msg.Type]; ok {
				handler(msg)
			} else {
				log.Printf("Unhandled event type: %s", msg.Type)
			}
		}
	}
}
