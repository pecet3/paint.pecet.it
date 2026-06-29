package paintroom

import (
	"context"
	"encoding/json"
	"time"

	"paint.pecet.it/pkg/ward/wardsocket"
)

type Event struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type ChatMessage struct {
	Name    string    `json:"name"`
	Uuid    string    `json:"uuid"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

type ServerMessage struct {
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

func (p *Paint) handleJoinEvent(ctx context.Context, c *wardsocket.Client) {
	canvasEvent := Event{
		Type:    "canvas_pixel_update",
		Payload: p.getAllPixelFrames(),
	}
	if data, err := json.Marshal(canvasEvent); err == nil {
		c.Send(data)
	}

}

func (p *Paint) handlePixelUpdate(ctx context.Context, evt *wardsocket.Event) {
	var data []byte
	err := json.Unmarshal(evt.Payload, &data)
	if err != nil {
		evt.Client.Request.Log("unmarshal err: ", err)
		return
	}
	if len(data) == 0 || len(data)%8 != 0 {
		return
	}
	p.setPixelFramesToBuffers(data)
}
