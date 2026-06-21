package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader konfiguruje protokół WebSocket (zezwala na połączenia z Reacta)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // W produkcji ogranicz do konkretnej domeny
	},
}

// Struktura wiadomości
type Message struct {
	Room    string `json:"room"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

// Klient reprezentuje pojedyncze połączenie użytkownika
type Client struct {
	conn *websocket.Conn
	room string
	send chan Message
}

// Pokój grupuje połączonych klientów
type Room struct {
	clients   map[*Client]bool
	broadcast chan Message
	join      chan *Client
	leave     chan *Client
}

// Menadżer pokojów
type Hub struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

var hub = Hub{
	rooms: make(map[string]*Room),
}

// Tworzy nowy pokój i uruchamia jego pętlę zdarzeń
func newRoom() *Room {
	return &Room{
		clients:   make(map[*Client]bool),
		broadcast: make(chan Message),
		join:      make(chan *Client),
		leave:     make(chan *Client),
	}
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case msg := <-r.broadcast:
			for client := range r.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}

// Obsługa czytania wiadomości od klienta
func (c *Client) readPump(r *Room) {
	defer func() {
		r.leave <- c
		c.conn.Close()
	}()

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Błąd odczytu: %v", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Błąd parsowania JSON: %v", err)
			continue
		}

		// Przesyłamy wiadomość do pokoju
		r.broadcast <- msg
	}
}

// Obsługa wysyłania wiadomości DO klienta
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for msg := range c.send {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}

// Endpoint WebSocket `/ws?room=nazwa_pokoju`
func handleConnections(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		roomName = "general" // domyślny pokój
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	hub.mu.Lock()
	room, exists := hub.rooms[roomName]
	if !exists {
		room = newRoom()
		hub.rooms[roomName] = room
		go room.run()
		log.Printf("Stworzono nowy pokój: %s", roomName)
	}
	hub.mu.Unlock()

	client := &Client{conn: conn, room: roomName, send: make(chan Message, 256)}
	room.join <- client

	go client.writePump()
	go client.readPump(room)
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	log.Println("Serwer WebSocket działa na :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
