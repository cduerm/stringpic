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

// Options allow to modify the string image generation.
type Option func(o *options) error

func errorOption(s string) Option {
	return func(o *options) error {
		return errors.New(s)
	}
}

// WithResultImage can be used to start the process with a non-empty result image.
// This can be useful, of a previous generation shall be continued.
func WithResultImage(img image.Image) Option {
	if img == nil {
		return errorOption("InputImage error: image must not be nil")
	}
	return func(o *options) error {
		o.result = img
		return nil
	}
}

// WithLinesCount defines the number of line segments used.
func WithLinesCount(n int) Option {
	if n < 0 {
		return errorOption("n must be positive integer")
	}
	return func(o *options) error {
		o.nLines = n
		return nil
	}
}

// WithPinCount uses pins in a circular pattern with the given number of pins.
func WithPinCount(n int) Option {
	if n < 2 {
		return errorOption("number of pins must be at least 2")
	}
	return func(o *options) error {
		o.pinCount = n
		return nil
	}
}

// WithPins allows to directly provide a list of pins. Can be used to define a non-circular
// pattern of pins.
func WithPins(pins []Pin) Option {
	if len(pins) < 2 {
		return errorOption("number of pins must be at least 2")
	}
	return func(o *options) error {
		o.pins = pins
		return nil
	}
}

// WithStringDarkness allows to set a custom string darkness, e.g. to control how often the same
// line will be used. 
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

// WithPaintColor allows to define a certain color (including alpha) for the string
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

// WithEraseColor allows to define a color (with alpha) that is used to paint out the already
// used connections in the target image. 
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

// WithEraseFactor allows to paint out a ceratin fraction of the painted line. Can be useful to control
// contrast in the final result. It always uses white to paint out the lines. Otherwise use WithEraseColor.
func WithEraseFactor(f float64) Option {
	if f < 0 {
		return errorOption("factor must be larger than 0")
	}
	return func(o *options) error {
		_, _, _, a := o.paintColor.RGBA()
		factor := f / 257
		value := uint8(min(255, float64(a)*factor))

		o.eraseColor = color.RGBA{value, value, value, value}
		return nil
	}
}

// WithDiameter allows to set the image diameter for an accurate string length calculation. Unit is meters.
func WithDiameter(d float64) Option {
	return func(o *options) error {
		o.circleDiameter = d
		return nil
	}
}

// WithResolution controls the pixel size of the output image. More pixels allow for more pleasant preview, 
// but a lower resolution might yield a better representation with physical string as fewer "buckets" for
// darkening the image are available.
func WithResolution(n int) Option {
	if n < 1 {
		return errorOption("resolution must be positive integer")
	}
	return func(o *options) error {
		o.resolution = n
		return nil
	}
}
