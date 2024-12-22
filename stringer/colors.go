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
	// fmt.Println(under, r1, g1, b1, a1, over, r2, g2, b2, a2, result, r/255, g/255, b/255, a/255)
	return result
}
