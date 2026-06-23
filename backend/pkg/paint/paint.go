package paint

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"paint.pecet.it/pkg/wsmanager"
)

type Pixel struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

type UpdateEvent struct {
	Type    string  `json:"type"`
	Payload []Pixel `json:"payload"`
}

type Paint struct {
	Room        *wsmanager.Room
	Canvas      *image.RGBA
	mu          sync.Mutex
	pixelBuffer []Pixel
}

func New(room *wsmanager.Room, width, height int) *Paint {
	p := &Paint{
		Room:        room,
		Canvas:      image.NewRGBA(image.Rect(0, 0, width, height)),
		pixelBuffer: make([]Pixel, 0),
	}

	p.Room.RegisterEventHandler("canvas_pixel_update", p.handlePixelUpdate)
	p.Room.RegisterJoinHandler(p.handleJoinEvent)
	return p
}

func (p *Paint) handleJoinEvent(c *wsmanager.Client) {
	p.mu.Lock()
	bounds := p.Canvas.Bounds()
	pixels := make([]Pixel, 0)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := p.Canvas.RGBAAt(x, y)
			if rgba.A > 0 {
				pixels = append(pixels, Pixel{
					X:     x,
					Y:     y,
					Color: fmt.Sprintf("#%02x%02x%02x", rgba.R, rgba.G, rgba.B),
				})
			}
		}
	}
	p.mu.Unlock()

	event := UpdateEvent{
		Type:    "canvas_pixel_update",
		Payload: pixels,
	}

	data, err := json.Marshal(event)
	if err == nil {
		c.Send(data)
	}
}

func (p *Paint) handlePixelUpdate(evt *wsmanager.Event) {
	var pixels []Pixel
	err := json.Unmarshal(evt.Payload, &pixels)
	if err != nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, px := range pixels {
		p.Canvas.Set(px.X, px.Y, parseHexColor(px.Color))
		p.pixelBuffer = append(p.pixelBuffer, px)
	}
}

func (p *Paint) Run() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()

		if len(p.pixelBuffer) > 0 {
			event := UpdateEvent{
				Type:    "canvas_pixel_update",
				Payload: p.pixelBuffer,
			}

			data, err := json.Marshal(event)
			if err == nil {
				p.Room.Broadcast(data)
			}

			p.pixelBuffer = make([]Pixel, 0)
		}

		p.mu.Unlock()
	}
}

func parseHexColor(s string) color.RGBA {
	c := color.RGBA{A: 255}
	if len(s) == 7 && s[0] == '#' {
		var r, g, b uint8
		fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
		c.R = r
		c.G = g
		c.B = b
	}
	return c
}
