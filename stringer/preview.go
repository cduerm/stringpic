package stringer

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/fogleman/gg"
)

type previewSettings struct {
	Size      int
	LineWidth float64
	LineColor color.Color
	Padding   float64
}

var PreviewSettings = previewSettings{
	Size:      2000,
	LineWidth: 2.8,
	LineColor: color.NRGBA{0, 0, 0, 80},
}

func Preview(pins []Pin, instructions []int) *image.RGBA {
	dc := gg.NewContext(PreviewSettings.Size, PreviewSettings.Size)
	dc.SetColor(color.White)
	dc.Clear()
	return PreviewOver(pins, instructions, dc.Image().(*image.RGBA))
}

func PreviewOver(pins []Pin, instructions []int, under *image.RGBA) *image.RGBA {
	var xl, xu, yl, yu float64 = math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64
	for _, p := range pins {
		xl = min(xl, p.X)
		xu = max(xu, p.X)
		yl = min(yl, p.Y)
		yu = max(yu, p.Y)
	}
	delta := max(xu-xl, yu-yl) + 2*min(xl, yl)

	dc := gg.NewContext(PreviewSettings.Size, PreviewSettings.Size)
	dc.DrawImage(under, 0, 0)
	dc.Scale(float64(PreviewSettings.Size)/delta, float64(PreviewSettings.Size)/delta)
	dc.SetColor(PreviewSettings.LineColor)
	dc.SetLineCapRound()
	dc.SetLineWidth(PreviewSettings.LineWidth)
	start := randomOffset(instructions[0], pins, 0.2)
	for i := range instructions[1:] {
		end := randomOffset(instructions[i+1], pins, 0.2)
		dc.DrawLine(start.X, start.Y, end.X, end.Y)
		start = end
	}
	dc.Stroke()
	return dc.Image().(*image.RGBA)
}

func randomOffset(pinId int, pins []Pin, fac float64) Pin {
	offset := rand.Float64()*2 - 1
	var dx, dy float64
	if offset > 0 {
		dx = fac * offset * (pins[(pinId+1)%len(pins)].X - pins[pinId].X)
		dy = fac * offset * (pins[(pinId+1)%len(pins)].Y - pins[pinId].Y)
	} else {
		dx = fac * offset * (pins[pinId].X - pins[(pinId-1+len(pins))%len(pins)].X)
		dy = fac * offset * (pins[pinId].Y - pins[(pinId-1+len(pins))%len(pins)].Y)
	}
	p := pins[pinId]
	return Pin{p.X + dx, p.Y + dy}
}
