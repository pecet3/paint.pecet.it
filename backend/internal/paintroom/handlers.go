package paintroom

import (
	"context"
	"encoding/json"

	"paint.pecet.it/pkg/ward/wardsocket"
)

type Event struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

func (p *Paint) handleJoinEvent(ctx context.Context, c *wardsocket.Client) {
	p.mu.Lock()
	p.saveCanvasBytes()
	canvasEvent := Event{
		Type:    "canvas_pixel_update",
		Payload: p.getCanvasBytes(),
	}
	p.mu.Unlock()
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
	p.mu.Lock()
	defer p.mu.Unlock()
	p.streamBuf = append(p.streamBuf, data...)
	p.saveBuf = append(p.saveBuf, p.streamBuf...)
}
