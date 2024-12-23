package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"

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
	allLines := stringer.CalculateLines(pins)

	currentPin := 0
	for range 5000 {
		bestScore := math.Inf(-1)
		var bestPoints []image.Point
		var bestPin = -1
		for i, linePoints := range allLines[currentPin] {
			if linePoints == nil {
				continue
			}
			score := stringer.Score(linePoints, targetImage, resultImage)
			if score > bestScore {
				bestScore = score
				bestPoints = linePoints
				bestPin = i
			}
		}
		if rand.Float64() > 0.990 {
			bestPoints = allLines[currentPin][rand.Intn(len(allLines[currentPin])-1)]
		}

		stringer.PixelOver(resultImage, bestPoints, color.RGBA{0, 0, 0, 20})
		stringer.PixelOver(targetImage, bestPoints, color.RGBA{20, 20, 20, 20})

		// fmt.Printf("going from %d to %d\n", currentPin, bestPin)
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
