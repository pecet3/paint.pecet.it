package wardsocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/ward"
)

const (
	pingPeriod     = 20 * time.Second
	writeWait      = 10 * time.Second
	pongWait       = 25 * time.Second
	maxMessageSize = 512 * 1024
)

type Client struct {
	conn    *websocket.Conn
	Request *ward.Request

	mu         sync.RWMutex
	pingSentAt time.Time
	latency    time.Duration

	sendCh   chan json.RawMessage
	streamCh chan json.RawMessage
}

func (c *Client) Send(msg json.RawMessage) {
	c.sendCh <- msg
}

func (c *Client) SendStream(msg json.RawMessage) {
	select {
	case c.streamCh <- msg:
	default:

	}
}

func (c *Client) GetLatency() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latency
}

func NewClient(conn *websocket.Conn, wreq *ward.Request) *Client {
	return &Client{
		conn:     conn,
		sendCh:   make(chan json.RawMessage, 10),
		streamCh: make(chan json.RawMessage, 10),
		Request:  wreq,
	}
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

		c.mu.Lock()
		if !c.pingSentAt.IsZero() {
			c.latency = time.Since(c.pingSentAt)
			c.pingSentAt = time.Time{}
		}
		c.mu.Unlock()

		return nil
	})

	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			c.Request.Log("Ws read message err", err)
			break
		}

		var e Event
		if err := json.Unmarshal(bytes, &e); err != nil {
			c.Request.Log("parsing err JSON:", err)
			continue
		}
		e.Client = c

		r.eventCh <- &e
	}
}

func (c *Client) writePump(r *Channel) {
	ticker := time.NewTicker(pingPeriod)

	var streamDelay <-chan time.Time
	streamReady := true

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

			if err := c.conn.WriteJSON(msg); err != nil {
				c.Request.Log(err)
				return
			}

		case msg, ok := <-c.streamCh:
			if !ok {
				return
			}

			if !streamReady {
				continue
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteJSON(msg); err != nil {
				c.Request.Log(err)
				return
			}

			c.mu.RLock()
			log.Println(c.latency)
			currentLatency := c.latency
			c.mu.RUnlock()

			if currentLatency > 0 {
				streamReady = false
				streamDelay = time.After(currentLatency)
			}

		case <-streamDelay:
			streamReady = true

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			c.mu.Lock()
			c.pingSentAt = time.Now()
			c.mu.Unlock()

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Request.Log("Ping send err:", err)
				return
			}
		}
	}
}
