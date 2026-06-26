package wardsocket

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

type Client struct {
	conn    *websocket.Conn
	Request *ward.Request

	sendCh    chan json.RawMessage
	roomIdent string
}

func (c *Client) Log(v ...any) {
	args := append([]any{"wardsocket room:", c.roomIdent}, v...)
	c.Request.Log(args...)
}
func (c *Client) Logf(format string, v ...any) {
	prefix := fmt.Sprintf("wardsocket room: %s ", c.roomIdent)
	c.Request.Logf(prefix+format, v...)
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
