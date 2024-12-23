package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"os"

	"github.com/cduerm/stringpic/stringer"
)

var filename = "flower512-contrast.png"
var pinCount = 300
var paddingPixel = 10

func main() {
	diskImage, err := stringer.OpenImageFromDisk(filename)
	if err != nil {
		panic(err)
	}
	bounds := diskImage.Bounds()
	targetImage := image.NewRGBA(bounds)
	draw.Draw(targetImage, bounds, diskImage, image.Point{}, draw.Src)
	fmt.Println(bounds)
	resultImage := image.NewRGBA(targetImage.Bounds())
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	pins := stringer.CalculatePins(pinCount, bounds, paddingPixel)
	// fmt.Println(pins)
	stringer.CalculateLines(pins)

	currentPin := 0
	gone := make(map[string]struct{})
	for range 5000 {
		p := pins[currentPin]
		bestScore := math.Inf(-1)
		bestPin := 0
		for i, q := range pins {
			path := fmt.Sprintf("%d,%d", currentPin, i)
			_, inGone := gone[path]
			// inGone = false
			if i == currentPin || inGone {
				continue
			}
			linePoints := stringer.LinePoints(p, q)
			score := stringer.Score(linePoints, targetImage, resultImage)
			if score > bestScore {
				bestScore = score
				bestPin = i
			}
		}
		if rand.Float64() > 0.990 {
			bestPin = rand.Intn(len(pins) - 1)
			if bestPin == currentPin {
				bestPin = len(pins) - 1
			}
		}

		pixels := stringer.LinePoints(p, pins[bestPin])
		stringer.PixelOver(resultImage, pixels, color.RGBA{0, 0, 0, 20})
		stringer.PixelOver(targetImage, pixels, color.RGBA{20, 20, 20, 20})

		// fmt.Printf("going from %d to %d\n", currentPin, bestPin)
		gone[fmt.Sprintf("%d,%d", currentPin, bestPin)] = struct{}{}
		currentPin = bestPin
	}

	for _, p := range pins {
		p.Draw(resultImage)
	}

	err = stringer.SaveImageToDisk("out.png", resultImage)
	if err != nil {
		panic(err)
	}
}

func test() {
	for i := range 256 {
		c := color.RGBA{uint8(i), uint8(i), uint8(i), uint8(i)}
		r, _, _, _ := c.RGBA()
		fmt.Printf("%d: %d (factor: %5.3f)\n", i, r, float64(r)/float64(i))
	}
	os.Exit(1)
}
