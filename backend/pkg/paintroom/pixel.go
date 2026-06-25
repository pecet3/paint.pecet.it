package paintroom

import (
	"encoding/binary"
	"image/color"
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

func (p *Paint) setPixelFramesToBuffers(data []byte) {
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

func (p *Paint) getAllPixelFrames() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
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
	return frameBuf
}
