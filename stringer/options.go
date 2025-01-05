package stringer

import (
	"errors"
	"image"
	"image/color"
)

type options struct {
	target, result         image.Image
	allLines               [][][]image.Point
	pinCount               int
	nLines                 int
	eraseColor, paintColor color.RGBA
	circleDiameter         float64
	resolution             int
	pins                   []Pin
}

var defaultOptions = options{
	nLines:         3000,
	paintColor:     color.RGBA{0, 0, 0, 30},
	eraseColor:     color.RGBA{30, 30, 30, 30},
	circleDiameter: 1.0,
	pinCount:       240,
	resolution:     500,
}

type Option func(o *options) error

func errorOption(s string) Option {
	return func(o *options) error {
		return errors.New(s)
	}
}

func WithResultImage(img image.Image) Option {
	if img == nil {
		return errorOption("InputImage error: image must not be nil")
	}
	return func(o *options) error {
		o.result = img
		return nil
	}
}

func WithLinesCount(n int) Option {
	if n < 0 {
		return errorOption("n must be positive integer")
	}
	return func(o *options) error {
		o.nLines = n
		return nil
	}
}

func WithPinCount(n int) Option {
	if n < 2 {
		return errorOption("number of pins must be at least 2")
	}
	return func(o *options) error {
		o.pinCount = n
		return nil
	}
}

func WithPins(pins []Pin) Option {
	if len(pins) < 2 {
		return errorOption("number of pins must be at least 2")
	}
	return func(o *options) error {
		o.pins = pins
		return nil
	}
}

func WithStringDarkness(d uint8) Option {
	if d == 0 {
		return errorOption("string darkness must be at least 1")
	}
	return func(o *options) error {
		o.paintColor = color.RGBA{0, 0, 0, d}
		o.eraseColor = color.RGBA{0, 0, 0, 0}
		return nil
	}
}

func WithPaintColor(c color.Color) Option {
	if c == nil {
		return errorOption("paint color cannot be nil")
	}
	return func(o *options) error {
		r, g, b, a := c.RGBA()
		o.paintColor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		return nil
	}
}

func WithEraseColor(c color.Color) Option {
	if c == nil {
		return errorOption("erase color cannot be nil")
	}
	return func(o *options) error {
		r, g, b, a := c.RGBA()
		o.eraseColor = color.RGBA{uint8(r / 257), uint8(g / 257), uint8(b / 257), uint8(a / 257)}
		return nil
	}
}

func WithEraseFactor(f float64) Option {
	if f < 0 {
		return errorOption("factor must be larger than 0")
	}
	return func(o *options) error {
		r, g, b, a := o.paintColor.RGBA()
		factor := f / 257

		o.eraseColor = color.RGBA{uint8(min(255, float64(r)*factor)), uint8(min(255, float64(g)*factor)), uint8(min(255, float64(b)*factor)), uint8(min(255, float64(a)*factor))}
		return nil
	}
}

func WithDiameter(d float64) Option {
	return func(o *options) error {
		o.circleDiameter = d
		return nil
	}
}

func WithResolution(n int) Option {
	if n < 1 {
		return errorOption("resolution must be positive integer")
	}
	return func(o *options) error {
		o.resolution = n
		return nil
	}
}
