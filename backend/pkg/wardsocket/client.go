package wardsocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingPeriod     = 30 * time.Second
	writeWait      = 60 * time.Second
	pongWait       = 45 * time.Second
	maxMessageSize = 1024 * 1024 * 10
)

type User interface {
	Name() string
	Uuid() string
	Rank() int
}

type nullUser struct{}

var nUser = &nullUser{}

const (
	null = "null"
)

func (u *nullUser) Uuid() string {
	return null
}
func (u *nullUser) Name() string {
	return null
}
func (u *nullUser) Rank() int {
	return 0
}

type Client struct {
	conn *websocket.Conn
	User User

	logFunc func(v ...any)
	sendCh  chan json.RawMessage
}

func (c *Client) Log(v ...any) {
	c.logFunc(v...)
}

func (c *Client) Send(msg json.RawMessage) {
	c.sendCh <- msg
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn, sendCh: make(chan json.RawMessage, 10), User: nUser, logFunc: func(v ...any) {
		log.Println(v...)
	}}
}

func (c *Client) AssingUser(u User) *Client {
	c.User = u
	return c
}
func (c *Client) RegisterLogFunc(fn func(...any)) *Client {
	c.logFunc = fn
	return c
}
func (c *Client) readPump(r *Channel) {
	defer func() {
		r.leaveCh <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			c.Log("Ws read message err", err)
			break
		}

		var e Event
		if err := json.Unmarshal(bytes, &e); err != nil {
			c.Log("parsing err JSON:", err)
			continue
		}
		e.Client = c

		r.eventCh <- &e
	}
}

func (c *Client) writePump(r *Channel) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		r.leaveCh <- c
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.sendCh:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(msg)
			if err != nil {
				c.Log(err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Log("Ping send err:", err)
				return
			}
		}
	}
}
