package stringer

import (
	"image"
	"image/color"
	"math"
)

const perPinLengthMeter = 0.0025

func Generate(targetImage, resultImage *image.RGBA, allLines [][][]image.Point, nLines int, stringDarkness uint8, diameterMeter float64) (instructions []int, length float64) {
	instructions = make([]int, 1, nLines+1)
	currentPin := 0
	for range nLines {
		bestScore := math.Inf(-1)
		var bestPoints []image.Point
		var bestPin = -1
		for i, linePoints := range allLines[currentPin] {
			if linePoints == nil {
				continue
			}
			score := Score(linePoints, targetImage, resultImage)
			if score > bestScore {
				bestScore = score
				bestPoints = linePoints
				bestPin = i
			}
		}

		PixelOver(resultImage, bestPoints, color.RGBA{0, 0, 0, stringDarkness})
		PixelOver(targetImage, bestPoints, color.RGBA{stringDarkness, stringDarkness, stringDarkness, stringDarkness})

		instructions = append(instructions, bestPin)
		diff := bestPoints[0].Sub(bestPoints[len(bestPoints)-1])
		length += perPinLengthMeter + math.Sqrt(float64(diff.X*diff.X+diff.Y*diff.Y))/float64(targetImage.Rect.Dx())*diameterMeter

		currentPin = bestPin
	}
	return instructions, length
}
