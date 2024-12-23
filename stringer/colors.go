package stringer

import (
	"image/color"
)

func ColorOver(under, over color.Color) color.Color {
	r1, g1, b1, a1 := under.RGBA()
	r2, g2, b2, a2 := over.RGBA()

	r := uint8((r2 + r1*(0xffff-a2)/0xffff) / 257)
	g := uint8((g2 + g1*(0xffff-a2)/0xffff) / 257)
	b := uint8((b2 + b1*(0xffff-a2)/0xffff) / 257)
	a := uint8((a2 + a1*(0xffff-a2)/0xffff) / 257)
	result := color.RGBA{r, g, b, a}
	return result
}

func ColorOverUint8(under, over color.RGBA) color.RGBA {
	r := (over.R + uint8(uint16(under.R)*uint16(0xff-over.A)/0xff))
	g := (over.G + uint8(uint16(under.G)*uint16(0xff-over.A)/0xff))
	b := (over.B + uint8(uint16(under.B)*uint16(0xff-over.A)/0xff))
	a := (over.A + uint8(uint16(under.A)*uint16(0xff-over.A)/0xff))
	result := color.RGBA{r, g, b, a}
	return result
}
