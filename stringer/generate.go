package stringer

import (
	"image"
	"math"
	"runtime"
	"sync"
)

const perPinLengthMeter = 0.001

// Generate will use a target image to calculate a string image and return the result image, the possibly altered
// target image (if options like WithEraseFactor are used), the order of pins to use and the length of string.
// An error will be returned, if one of the given options is used incorrectly.
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
		bestPoints, bestPin := getBestLineParallel(scoreFunction, currentPin, o, targetImage, resultImage)

		PixelOver(resultImage, bestPoints, o.paintColor)
		PixelOver(targetImage, bestPoints, o.eraseColor)

		instructions = append(instructions, bestPin)
		diff := bestPoints[0].Sub(bestPoints[len(bestPoints)-1])
		length += perPinLengthMeter + math.Sqrt(float64(diff.X*diff.X+diff.Y*diff.Y))/float64(targetImage.Rect.Dx())*o.circleDiameter

		currentPin = bestPin
	}
	return resultImage, targetImage, instructions, length, nil
}

func getBestLine(scoreFunction ScoreFunction, currentPin int, o options, targetImage, resultImage *image.RGBA) (bestPoints []image.Point, bestPin int) {
	bestScore := math.Inf(-1)
	bestPin = (currentPin + o.pinCount/2) % o.pinCount
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
	return
}

func getBestLineParallel(scoreFunction ScoreFunction, currentPin int, o options, targetImage, resultImage *image.RGBA) (bestPoints []image.Point, bestPin int) {
	n := runtime.NumCPU()
	bestScores := make([]float64, n)
	bestPins := make([]int, n)
	lines := o.allLines[currentPin]
	lenLines := len(lines)
	wg := new(sync.WaitGroup)
	for id := range n {
		wg.Add(1)
		go func() {
			bestScore := math.Inf(-1)
			bestPin := (currentPin + o.pinCount/2) % o.pinCount
			for i, linePoints := range lines[id*lenLines/n : (id+1)*lenLines/n] {
				if linePoints == nil {
					continue
				}
				score := scoreFunction(linePoints, targetImage, resultImage)
				if score > bestScore {
					bestScore = score
					bestPin = id*lenLines/n + i
				}
			}
			bestScores[id] = bestScore
			bestPins[id] = bestPin
			wg.Done()
		}()
	}

	wg.Wait()
	bestScore := math.Inf(-1)
	for i := range bestPins {
		if bestScores[i] > bestScore {
			bestScore = bestScores[i]
			bestPin = bestPins[i]
		}
	}
	bestPoints = lines[bestPin]
	return
}
