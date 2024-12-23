package stringer

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

const AaSmoothing float64 = 1

func Disk(dst draw.Image, center Pin, radius float64, col color.Color) {
	// r2 := radius * radius
	for x := int(center.X - radius - 1); x < int(center.X+radius+2); x++ {
		for y := int(center.Y - radius - 1); y < int(center.Y+radius+2); y++ {
			if !image.Pt(x, y).In(dst.Bounds()) {
				continue
			}
			dx := float64(x) - center.X
			dy := float64(y) - center.Y
			// delta := dx*dx + dy*dy - r2
			// inside := delta < 0
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

func Circle(dst draw.Image, center Pin, radius float64, thickness float64, col color.Color) {
	thickness = thickness / 2
	// r2 := radius * radius
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
		c := ColorOverUint8(cUnder, cOver)
		// fmt.Println(cUnder, cOver, c)
		pix[idx+0] = c.R
		pix[idx+1] = c.G
		pix[idx+2] = c.B
		pix[idx+3] = c.A
	}
}
