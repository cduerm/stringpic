package stringer

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
)

// Pin describes the location of a pin where string can be wrapped around.
type Pin struct {
	X, Y float64
}

const randomness = 0.2

// CalculatePins takes the number of pins, the image boundary and a padding around the edges
// and calculates the pin locations in a circular pattern. Some randomness is added to 
// prevent or reduce ringing patterns.
func CalculatePins(n int, bounds image.Rectangle, padding int) []Pin {
	centerX, centerY := float64(bounds.Max.X-bounds.Min.X)/2, float64(bounds.Max.Y-bounds.Min.Y)/2
	radius := centerX - float64(padding)

	pins := make([]Pin, n)
	step := 2 * math.Pi / float64(n)
	for i := range n {
		r := 2 * (rand.Float64() - 0.5) * step * randomness
		x := centerX + radius*math.Sin(float64(i)*step+r)
		y := centerY - radius*math.Cos(float64(i)*step+r)
		pins[i] = Pin{x, y}
	}
	return pins
}

// Draw will indicate the pin location in the image with a red disk.
func (p Pin) Draw(img draw.Image) {
	col := color.RGBA{255, 0, 0, 255}
	Disk(img, p, 2, col)
}

// CalculateLines takes a list of pins and then pre-calculates all pixels that lie
// on a line between all possible pin pairs. The list of pixels will be nil, if the
// pins are too close together to reduce the occurence of very short strings.
func CalculateLines(pins []Pin) [][][]image.Point {
	n := len(pins)
	lines := make([][][]image.Point, n)
	for i, p := range pins {
		lines[i] = make([][]image.Point, n)
		for j, q := range pins {
			s, b := min(i, j), max(i, j)
			diff := min(b-s, n+s-b)
			if diff < min(n/10, 10) {
				continue
			}
			lines[i][j] = LinePoints(p, q)
		}
	}
	return lines
}
