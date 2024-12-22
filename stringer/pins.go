package stringer

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

type Pin struct {
	X, Y float64
}

func CalculatePins(n int, bounds image.Rectangle, padding int) []Pin {
	centerX, centerY := float64(bounds.Max.X-bounds.Min.X)/2, float64(bounds.Max.Y-bounds.Min.Y)/2
	radius := centerX - float64(padding)

	pins := make([]Pin, n)
	step := 2 * math.Pi / float64(n)
	for i := range n {
		x := centerX + radius*math.Sin(float64(i)*step)
		y := centerY + radius*math.Cos(float64(i)*step)
		pins[i] = Pin{x, y}
	}
	return pins
}

func (p Pin) Draw(img draw.Image) {
	col := color.RGBA{255, 0, 0, 255}
	Disk(img, p, 3, col)
}
