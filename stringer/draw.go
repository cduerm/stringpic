package stringer

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Smothing radius for anti-aliasing. This is not used currently
const AaSmoothing float64 = 1

// Disk draws a solid disk at pin location
func Disk(dst draw.Image, center Pin, radius float64, col color.Color) {
	for x := int(center.X - radius - 1); x < int(center.X+radius+2); x++ {
		for y := int(center.Y - radius - 1); y < int(center.Y+radius+2); y++ {
			if !image.Pt(x, y).In(dst.Bounds()) {
				continue
			}
			dx := float64(x) - center.X
			dy := float64(y) - center.Y
			delta := math.Sqrt(dx*dx+dy*dy) - radius
			if delta > AaSmoothing {
				continue
			} else if delta < -AaSmoothing {
				dst.Set(x, y, col)
			} else {
				// Anti-Alias TBD
			}
		}
	}
}

// Circle draws a circle with radius and thickness at the pin location
func Circle(dst draw.Image, center Pin, radius float64, thickness float64, col color.Color) {
	thickness = thickness / 2
	for x := int(center.X - radius - thickness - 1); x < int(center.X+radius+thickness+2); x++ {
		for y := int(center.Y - radius - thickness - 1); y < int(center.Y+radius+thickness+2); y++ {
			if !image.Pt(x, y).In(dst.Bounds()) {
				continue
			}
			dx := float64(x) - center.X
			dy := float64(y) - center.Y
			delta := math.Sqrt(dx*dx+dy*dy) - radius
			if delta > thickness+AaSmoothing || delta < -thickness-AaSmoothing {
				continue
			} else if delta < thickness-AaSmoothing || delta > -thickness+AaSmoothing {
				dst.Set(x, y, col)
			} else {
				// Anti-Alias TBD
			}
		}
	}
}

// PixelOver draws (using the over operation) the specified color at all pixels. It is only implemented
// for RGBA images currently.
func PixelOver(dst draw.Image, pixels []image.Point, cOver color.RGBA) {
	var pix []uint8
	var img *image.RGBA
	if i, ok := dst.(*image.RGBA); !ok {
		panic("can only use image.RGBA image")
	} else {
		img = i
		pix = img.Pix
	}
	for _, p := range pixels {
		idx := img.PixOffset(p.X, p.Y)
		cUnder := color.RGBA{pix[idx+0], pix[idx+1], pix[idx+2], pix[idx+3]}
		c := ColorOverRGBA(cUnder, cOver)
		pix[idx+0] = c.R
		pix[idx+1] = c.G
		pix[idx+2] = c.B
		pix[idx+3] = c.A
	}
}
