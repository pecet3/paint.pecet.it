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

// func (p *Paint) setPixelFramesToBuffers(data []byte) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	for i := 0; i < len(data); i += 8 {
// 		x := int(binary.LittleEndian.Uint16(data[i : i+2]))
// 		y := int(binary.LittleEndian.Uint16(data[i+2 : i+4]))

// 		p.Canvas.SetRGBA(x, y, color.RGBA{
// 			R: data[i+4],
// 			G: data[i+5],
// 			B: data[i+6],
// 			A: data[i+7],
// 		})
// 	}
// 	p.pixelFrameBuf = append(p.pixelFrameBuf, data...)
// }

func (p *Paint) saveCanvasBytes() {
	data := p.saveBuf
	for i := 0; i < len(data); i += 8 {
		x := int(binary.LittleEndian.Uint16(data[i : i+2]))
		y := int(binary.LittleEndian.Uint16(data[i+2 : i+4]))

		sr := uint32(data[i+4])
		sg := uint32(data[i+5])
		sb := uint32(data[i+6])
		sa := uint32(data[i+7])

		if sa == 0 {
			continue
		}

		dst := p.Canvas.RGBAAt(x, y)
		dr := uint32(dst.R)
		dg := uint32(dst.G)
		db := uint32(dst.B)
		da := uint32(dst.A)

		var outR, outG, outB, outA uint32

		if sa == 255 {

			outR, outG, outB, outA = sr, sg, sb, sa
		} else {
			sr = (sr * sa) / 255
			sg = (sg * sa) / 255
			sb = (sb * sa) / 255

			invA := 255 - sa

			outR = sr + (dr*invA)/255
			outG = sg + (dg*invA)/255
			outB = sb + (db*invA)/255
			outA = sa + (da*invA)/255
		}

		p.Canvas.SetRGBA(x, y, color.RGBA{
			R: uint8(outR),
			G: uint8(outG),
			B: uint8(outB),
			A: uint8(outA),
		})
	}
	p.saveBuf = p.saveBuf[:0]
}

func (p *Paint) getCanvasBytes() []byte {

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
