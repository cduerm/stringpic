package stringer

import (
	"image"
	"math"
)

const perPinLengthMeter = 0.001

func Generate(target image.Image, options ...Option) (resultImage, targetImage *image.RGBA, instructions []int, length float64, err error) {
	o := defaultOptions
	o.target = target
	o.result = image.White

	for _, opt := range options {
		err = opt(&o)
		if err != nil {
			return nil, nil, nil, 0, err
		}
	}
	targetImage = RescaleImage(o.target, o.resolution)
	resultImage = RescaleImage(o.result, o.resolution)
	if o.pins == nil {
		o.pins = CalculatePins(o.pinCount, targetImage.Bounds(), 1)
	}
	o.allLines = CalculateLines(o.pins)

	scoreFunction := Score

	instructions = make([]int, 1, o.nLines+1)
	currentPin := 0
	for range o.nLines {
		bestScore := math.Inf(-1)
		var bestPoints []image.Point
		var bestPin = (currentPin + o.pinCount/2) % o.pinCount
		for i, linePoints := range o.allLines[currentPin] {
			if linePoints == nil {
				continue
			}
			score := scoreFunction(linePoints, targetImage, resultImage)
			if score > bestScore {
				bestScore = score
				bestPoints = linePoints
				bestPin = i
			}
		}

		PixelOver(resultImage, bestPoints, o.paintColor)
		PixelOver(targetImage, bestPoints, o.eraseColor)

		instructions = append(instructions, bestPin)
		diff := bestPoints[0].Sub(bestPoints[len(bestPoints)-1])
		length += perPinLengthMeter + math.Sqrt(float64(diff.X*diff.X+diff.Y*diff.Y))/float64(targetImage.Rect.Dx())*o.circleDiameter

		currentPin = bestPin
	}
	return resultImage, targetImage, instructions, length, nil
}
