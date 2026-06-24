package wsmanager

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/guardian"
)

type Client struct {
	conn      *websocket.Conn
	sendCh    chan json.RawMessage
	roomIdent string
	Request   *guardian.Request
}

func (c *Client) Log(v ...any) {
	log.Println("room:", c.roomIdent, c.Request.LogInfo(), v)
}

func (c *Client) Send(msg json.RawMessage) {
	c.sendCh <- msg
}

func NewClient(r *Room, conn *websocket.Conn, greq *guardian.Request) *Client {
	return &Client{conn: conn, sendCh: make(chan json.RawMessage), Request: greq, roomIdent: r.Ident}
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
