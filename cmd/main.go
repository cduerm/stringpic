package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"path"
	"strings"

	"github.com/cduerm/stringpic/stringer"
)

var filename = "flower512-contrast.png"

var pinCount = 300
var paddingPixel = 0
var outputSize = 512
var nLines = 2000
var diameterMeter float64 = 0.226
var outDir = "output"

var stringDarkness = uint8(max(1, min(255, 20*(float64(outputSize)/400)*(2500/float64(nLines)))))

func init() {
	flag.StringVar(&filename, "filename", filename, "png file to convert to string art")
	flag.StringVar(&outDir, "output", outDir, "directory where to put the output files")
	flag.IntVar(&pinCount, "pinCount", pinCount, "number of pins in circular pattern")
	flag.IntVar(&outputSize, "size", outputSize, "size of output image")
	flag.IntVar(&nLines, "nLines", nLines, "number of lines")
	flag.Float64Var(&diameterMeter, "diameter [mm]", diameterMeter, "diameter of ring (for string length calculation)")
	flag.Parse()
}

func main() {
	targetImage, resultImage, err := getImages(outputSize, filename)
	if err != nil {
		panic(err)
	}

	pins := stringer.CalculatePins(pinCount, resultImage.Bounds(), paddingPixel)
	allLines := stringer.CalculateLines(pins)

	instructions, length := stringer.Generate(targetImage, resultImage, allLines, nLines, stringDarkness, diameterMeter)

	outFilenameBase, _ := strings.CutSuffix(path.Base(filename), path.Ext(filename))
	err = stringer.WriteInstructionsToDisk(path.Join(outDir, outFilenameBase+"_instructions.txt"), instructions, length)
	if err != nil {
		panic(err)
	}

	err = stringer.SaveImageToDisk(path.Join(outDir, outFilenameBase+"_stringer.png"), resultImage)
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
