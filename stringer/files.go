package stringer

import (
	"image"
	"image/png"
	"os"
)

func OpenImageFromDisk(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func SaveImageToDisk(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil
}
