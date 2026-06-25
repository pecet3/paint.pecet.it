package wardsocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

type Client struct {
	conn    *websocket.Conn
	Request *ward.Request

	sendCh    chan json.RawMessage
	roomIdent string

	stateMu sync.RWMutex
	states  map[string]any
}

func (c *Client) RegisterState(key string, initialValue any) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	c.states[key] = initialValue
}

func (c *Client) SetState(key string, value any) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	c.states[key] = value
}

func (c *Client) GetState(key string) (any, bool) {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	val, exists := c.states[key]
	return val, exists
}
func GetClientStateAs[T any](c *Client, key string) (T, bool) {
	var zero T
	val, exists := c.GetState(key)
	if !exists {
		return zero, false
	}
	typedVal, ok := val.(T)
	return typedVal, ok
}

func (c *Client) Log(v ...any) {
	log.Println("room:", c.roomIdent, c.Request.LogInfo(), v)
}

func (c *Client) Send(msg json.RawMessage) {
	c.sendCh <- msg
}

func NewClient(r *Room, conn *websocket.Conn, wreq *ward.Request) *Client {
	return &Client{conn: conn, sendCh: make(chan json.RawMessage), Request: wreq, roomIdent: r.Ident}
}
func (c *Client) readPump(r *Room) {
	defer func() {
		r.leaveCh <- c
		c.conn.Close()
	}()
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			c.Log("Ws read message err: %v", err)
			break
		}

		var e Event
		if err := json.Unmarshal(bytes, &e); err != nil {
			c.Log("parsing err JSON: %v", err)
			continue
		}
		e.Client = c

		r.eventCh <- &e
	}
}

func (c *Client) writePump(r *Room) {
	defer func() {
		r.leaveCh <- c
		c.conn.Close()
	}()

	for msg := range c.sendCh {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			c.Log(err)
			break
		}
	}
}
