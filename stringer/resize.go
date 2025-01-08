package stringer

import "image"

// RescaleImage takes an image and a size and will create an RGBA image with
// square dimensions. The input image will always be cropped to square aspect ratio
// and scaled using nearest neighbour algorithm.
func RescaleImage(oldImg image.Image, newSize int) *image.RGBA {
	bounds := oldImg.Bounds()
	oldSize := min(bounds.Dx(), bounds.Dy())
	newImg := image.NewRGBA(image.Rect(0, 0, newSize, newSize))

	oldCenterX, oldCenterY := float64(bounds.Dx()+1)/2, float64(bounds.Dy()+1)/2
	newCenterX, newCenterY := float64(newSize+1)/2, float64(newSize+1)/2
	scale := float64(oldSize) / float64(newSize)
	atOld := func(x, y int) (r, g, b, a uint8) {
		xf, yf := float64(x)+0.5-newCenterX, float64(y)+0.5-newCenterY
		Xf, Yf := scale*(xf), scale*(yf)
		X, Y := int(oldCenterX+Xf), int(oldCenterY+Yf)
		R, G, B, A := oldImg.At(X, Y).RGBA()
		return uint8(R / 257), uint8(G / 257), uint8(B / 257), uint8(A / 257)
	}

	for y := range newSize {
		for x := range newSize {
			idx := newImg.PixOffset(x, y)
			r, g, b, a := atOld(x, y)
			newImg.Pix[idx+0] = r
			newImg.Pix[idx+1] = g
			newImg.Pix[idx+2] = b
			newImg.Pix[idx+3] = a
		}
	}
	return newImg
}
