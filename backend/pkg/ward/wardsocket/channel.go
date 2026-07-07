package wardsocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"
	"sync/atomic"
)

var channelCounter uint64

type Channel struct {
	ID      uint64
	clients map[*Client]bool
	cMu     sync.RWMutex

	joinCh  chan *Client
	leaveCh chan *Client
	eventCh chan *Event

	eventHandlers map[string]func(context.Context, *Event)

	joinHandlers  []func(context.Context, *Client)
	leaveHandlers []func(context.Context, *Client)

	cancel  context.CancelFunc
	closeCh chan struct{}
}

func NewChannel() *Channel {
	r := &Channel{
		ID:      atomic.AddUint64(&channelCounter, 1),
		clients: make(map[*Client]bool),
		joinCh:  make(chan *Client, 10),
		leaveCh: make(chan *Client, 10),
		eventCh: make(chan *Event, 10),

		eventHandlers: make(map[string]func(context.Context, *Event)),
		joinHandlers:  []func(context.Context, *Client){},
		leaveHandlers: []func(context.Context, *Client){},

		closeCh: make(chan struct{}),
	}
	return r
}

func (r *Channel) WithCancelContext(ctx context.Context) (*Channel, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	return r, ctx
}

func (r *Channel) Broadcast(msg json.RawMessage, omitClients ...*Client) {
	r.cMu.RLock()
	defer r.cMu.RUnlock()
	if len(omitClients) > 0 {
		for client := range r.clients {
			if !slices.Contains(omitClients, client) {
				client.sendCh <- msg
			}
		}
	} else {
		for client := range r.clients {
			client.sendCh <- msg
		}
	}
}

func (r *Channel) LeaveClient(client *Client) {
	r.leaveCh <- client
}

func (r *Channel) JoinClient(client *Client) {

	r.joinCh <- client

	go client.writePump(r)
	go client.readPump(r)
}

func (r *Channel) RegisterEventHandler(eventType string, handler func(context.Context, *Event)) {
	r.eventHandlers[eventType] = handler
}

func (r *Channel) RegisterJoinHandler(handler func(context.Context, *Client)) {
	r.joinHandlers = append(r.joinHandlers, handler)
}

func (r *Channel) RegisterLeaveHandler(handler func(context.Context, *Client)) {
	r.leaveHandlers = append(r.leaveHandlers, handler)
}

func (r *Channel) Close() {
	r.closeCh <- struct{}{}
}

func (r *Channel) LogInfo() string {
	return fmt.Sprintf("Channel %d |", r.ID)
}

func (r *Channel) Log(v ...any) {
	args := append([]any{r.LogInfo()}, v...)
	log.Println(args...)
}

func (r *Channel) Run(ctx context.Context) {
	go func() {
		r.Log("is listening")
		for {
			select {
			case client := <-r.joinCh:
				r.clients[client] = true
				if len(r.joinHandlers) > 0 {
					for _, handle := range r.joinHandlers {
						go handle(ctx, client)
					}
				}
			case client := <-r.leaveCh:

				if len(r.leaveHandlers) > 0 {
					for _, handle := range r.leaveHandlers {
						go handle(ctx, client)
					}
				}

				r.cMu.Lock()
				if _, ok := r.clients[client]; ok {
					delete(r.clients, client)
					close(client.sendCh)
				}
				r.cMu.Unlock()
			case msg := <-r.eventCh:
				if handler, ok := r.eventHandlers[msg.Type]; ok {
					go handler(ctx, msg)
				} else {
					r.Log("unhandled event type:", msg.Type)
				}
			case <-r.closeCh:
				r.Log("closing")
				r.cMu.Lock()
				for client := range r.clients {
					close(client.sendCh)
					delete(r.clients, client)
				}
				r.cMu.Unlock()
				r.Log("removed all clients")
				if r.cancel != nil {
					r.Log("executing context cancel func")
					r.cancel()
				}
				return
			}
		}
	}()
}
