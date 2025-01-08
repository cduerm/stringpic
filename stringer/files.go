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

// OpenImageFromDisk can be used to create an image.Image by opening a file. It returns
// the image or the error of the os.Open or image.Decode function calls.
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

// SaveImageToDisk will save an image.Image to the given file. If there are errors creating
// the file or encoding the image the error is returned.
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

// WriteInstructionsToDisk writes the instructions (a list of pin IDs) to a text file, including the length
// of string required. If there is an error creating the file, it is returend.
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

// GetImages creates from a filename and a resolution a square target image and a white (empty) result image
// to process using the Generate function.
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
