package stringer

import (
	"fmt"
	"image"
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
