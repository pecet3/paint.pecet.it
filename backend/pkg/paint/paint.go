package paint

import (
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"sync"
	"time"

	"paint.pecet.it/pkg/wsmanager"
)

// offset  8
type PixelFrame struct {
	X uint16
	Y uint16
	R uint8
	G uint8
	B uint8
	A uint8
}

type UpdateEvent struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type Paint struct {
	Room   *wsmanager.Room
	Canvas *image.RGBA
	mu     sync.Mutex

	pixelFrameBuf []byte
}

func New(room *wsmanager.Room, width, height int) *Paint {
	p := &Paint{
		Room:          room,
		Canvas:        image.NewRGBA(image.Rect(0, 0, width, height)),
		pixelFrameBuf: make([]byte, 0),
	}

	p.Room.RegisterEventHandler("canvas_pixel_update", p.handlePixelUpdate)
	p.Room.RegisterJoinHandler(p.handleJoinEvent)
	return p
}

func (p *Paint) handleJoinEvent(c *wsmanager.Client) {
	c.Log("joined to room")

	p.mu.Lock()
	bounds := p.Canvas.Bounds()

	var frameBuf []byte

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := p.Canvas.RGBAAt(x, y)
			if rgba.A > 0 {
				var buf [8]byte

				binary.LittleEndian.PutUint16(buf[0:2], uint16(x))
				binary.LittleEndian.PutUint16(buf[2:4], uint16(y))

				buf[4] = rgba.R
				buf[5] = rgba.G
				buf[6] = rgba.B
				buf[7] = rgba.A
				frameBuf = append(frameBuf, buf[:]...)
			}
		}
	}
	p.mu.Unlock()

	event := UpdateEvent{
		Type:    "canvas_pixel_update",
		Payload: frameBuf,
	}

	data, err := json.Marshal(event)
	if err == nil {
		c.Send(data)
	}
}

func (p *Paint) handlePixelUpdate(evt *wsmanager.Event) {
	var data []byte
	err := json.Unmarshal(evt.Payload, &data)
	if err != nil {
		evt.Client.Log("unmarshal err: ", err)
		return
	}
	if len(data) == 0 || len(data)%8 != 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < len(data); i += 8 {
		x := int(binary.LittleEndian.Uint16(data[i : i+2]))
		y := int(binary.LittleEndian.Uint16(data[i+2 : i+4]))

		p.Canvas.SetRGBA(x, y, color.RGBA{
			R: data[i+4],
			G: data[i+5],
			B: data[i+6],
			A: data[i+7],
		})
	}
	p.pixelFrameBuf = append(p.pixelFrameBuf, data...)
}

func (p *Paint) Run() {
	go func() {
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			p.mu.Lock()

			if len(p.pixelFrameBuf) > 0 {
				event := UpdateEvent{
					Type:    "canvas_pixel_update",
					Payload: p.pixelFrameBuf,
				}

				data, err := json.Marshal(event)
				if err == nil {
					p.Room.Broadcast(data)
				}

				p.pixelFrameBuf = p.pixelFrameBuf[:0]
			}

			p.mu.Unlock()
		}
	}()
}
