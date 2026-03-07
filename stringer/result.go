package stringer

import "image"

type Result struct {
	StartImage   *image.RGBA
	EndImage     *image.RGBA
	Image        *image.RGBA
	Instructions []int
	StringLength float64
	Pins         []Pin
}
