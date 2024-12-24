package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strings"

	"github.com/cduerm/stringpic/stringer"
)

var filename = "flower512-contrast.png"

var pinCount = 300
var paddingPixel = 0
var outputSize = 512
var nLines = 4000

var stringDarkness = max(1, min(255, 20*(float64(outputSize)/400)*(2500/float64(nLines))))

func init() {
	flag.StringVar(&filename, "filename", filename, "png file to convert to string art")
	flag.IntVar(&pinCount, "pinCount", pinCount, "number of pins in circular pattern")
	flag.IntVar(&outputSize, "size", outputSize, "size of output image")
	flag.IntVar(&nLines, "nLines", nLines, "number of lines")
	flag.Parse()
}

func main() {
	targetImage, resultImage, err := getImages(outputSize, filename)
	if err != nil {
		panic(err)
	}

	pins := stringer.CalculatePins(pinCount, resultImage.Bounds(), paddingPixel)
	allLines := stringer.CalculateLines(pins)

	var instructions = new(strings.Builder)
	var length = 0.0
	fmt.Fprintln(instructions, "start at pin 1, top center, counting clockwise")
	currentPin := 0
	for i := range nLines {
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

		stringer.PixelOver(resultImage, bestPoints, color.RGBA{0, 0, 0, uint8(stringDarkness)})
		stringer.PixelOver(targetImage, bestPoints, color.RGBA{uint8(stringDarkness), uint8(stringDarkness), uint8(stringDarkness), uint8(stringDarkness)})

		fmt.Fprintf(instructions, "line % 4d: next Pin is % 3d\n", i+1, bestPin+1)
		length += 1

		currentPin = bestPin
	}

	err = os.WriteFile("instructions.txt", []byte(instructions.String()), os.ModePerm)
	if err != nil {
		panic(err)
	}

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

	targetImage = stringer.RescaleImage(diskImage, size)

	resultImage = image.NewRGBA(targetImage.Bounds())
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	return targetImage, resultImage, nil
}
