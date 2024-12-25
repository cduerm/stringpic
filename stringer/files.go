package stringer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	_ "image/jpeg"
)

func OpenImageFromDisk(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil
}

func WriteInstructionsToDisk(filename string, instructions []int, length float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "Start at pin #0 at the top and count in clockwise direction\nYou will need around %.1f m of string\n", length)
	for i, pin := range instructions[1:] {
		fmt.Fprintf(file, "step %d: Go to pin #%d\n", i+1, pin)
	}
	fmt.Fprintf(file, "Your're done. Congratulations!\n")

	return nil
}

func GetImages(size int, filename string) (targetImage, resultImage *image.RGBA, err error) {
	diskImage, err := OpenImageFromDisk(filename)
	if err != nil {
		return nil, nil, err
	}

	targetImage = RescaleImage(diskImage, size)

	resultImage = image.NewRGBA(targetImage.Bounds())
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	return targetImage, resultImage, nil
}
