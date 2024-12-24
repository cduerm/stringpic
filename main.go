package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"

	"github.com/cduerm/stringpic/stringer"
)

var filename = "flower512-contrast.png"

const pinCount = 300
const paddingPixel = 2
const outputSize = 512
const nLines = 4000

var stringDarkness = max(1, min(255, 20*(float64(outputSize)/400)*(2500/float64(nLines))))

func main() {
	targetImage, resultImage, err := getImages(outputSize, filename)
	if err != nil {
		panic(err)
	}

	pins := stringer.CalculatePins(pinCount, resultImage.Bounds(), paddingPixel)
	// fmt.Println(pins)
	allLines := stringer.CalculateLines(pins)

	currentPin := 0
	for range nLines {
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

		stringer.PixelOver(resultImage, bestPoints, color.RGBA{0, 0, 0, uint8(stringDarkness)})
		stringer.PixelOver(targetImage, bestPoints, color.RGBA{uint8(stringDarkness), uint8(stringDarkness), uint8(stringDarkness), uint8(stringDarkness)})

		// fmt.Printf("going from %d to %d\n", currentPin, bestPin)
		currentPin = bestPin
	}

	// for _, p := range pins {
	// 	p.Draw(resultImage)
	// }

	err = stringer.SaveImageToDisk("out.png", resultImage)
	if err != nil {
		panic(err)
	}
}

func getImages(size int, filename string) (targetImage, resultImage *image.RGBA, err error) {
	diskImage, err := stringer.OpenImageFromDisk(filename)
	if err != nil {
		return nil, nil, err
	}
	// bounds := diskImage.Bounds()

	targetImage = stringer.RescaleImage(diskImage, size)

	// targetImage = image.NewRGBA(bounds)
	// draw.Draw(targetImage, bounds, diskImage, image.Point{}, draw.Src)

	resultImage = image.NewRGBA(targetImage.Bounds())
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	return targetImage, resultImage, nil
}
