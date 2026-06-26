package wardsocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

type Client struct {
	conn    *websocket.Conn
	Request *ward.Request

	sendCh chan json.RawMessage
}

func (c *Client) Send(msg json.RawMessage) {
	c.sendCh <- msg
}

func NewClient(r *Room, conn *websocket.Conn, wreq *ward.Request) *Client {
	return &Client{conn: conn, sendCh: make(chan json.RawMessage), Request: wreq}
}
func (c *Client) readPump(r *Room) {
	defer func() {
		r.leaveCh <- c
		c.conn.Close()
	}()
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			c.Request.Log("Ws read message err: %v", err)
			break
		}

		var e Event
		if err := json.Unmarshal(bytes, &e); err != nil {
			c.Request.Log("parsing err JSON: %v", err)
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
			c.Request.Log(err)
			break
		}
	}
}
