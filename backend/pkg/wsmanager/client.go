package wsmanager

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	sendCh chan json.RawMessage
}

func (c *Client) Send(msg json.RawMessage) {
	c.sendCh <- msg
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn, sendCh: make(chan json.RawMessage)}
}
func (c *Client) readPump(r *Room) {
	defer func() {
		r.leaveCh <- c
		c.conn.Close()
	}()
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Ws read message err: %v", err)
			break
		}

		var e Event
		if err := json.Unmarshal(bytes, &e); err != nil {
			log.Printf("parsing err JSON: %v", err)
			continue
		}
		e.Client = c

		r.eventCh <- &e
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for msg := range c.sendCh {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}
