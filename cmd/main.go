package main

import (
	"flag"
	"fmt"
	"image/color"
	"path"
	"strings"

	"github.com/cduerm/stringpic/htmlViewer"
	"github.com/cduerm/stringpic/stringer"
)

var filename = ""

var pinCount = 300
var outputSize = 512
var nLines = 2000
var diameterMM float64 = 0.226
var outDir = "output"
var eraseFactor = 0.2

var stringDarkness = int(max(1, min(255, 20*(float64(outputSize)/400)*(2500/float64(nLines)))))

func init() {
	flag.StringVar(&filename, "filename", filename, "png file to convert to string art")
	flag.StringVar(&outDir, "output", outDir, "directory where to put the output files")
	flag.IntVar(&pinCount, "pinCount", pinCount, "number of pins in circular pattern")
	flag.IntVar(&outputSize, "size", outputSize, "size of output image")
	flag.IntVar(&nLines, "nLines", nLines, "number of lines")
	flag.IntVar(&stringDarkness, "darkness", int(stringDarkness), "string darkness (value between 1 and 255)")
	flag.Float64Var(&eraseFactor, "erase", eraseFactor, "how much of the original image will be erased by the lines")
	flag.Float64Var(&diameterMM, "diameter", diameterMM, "diameter of ring (for string length calculation) in mm")
}

func main() {
	flag.Parse()

	target, err := stringer.OpenImageFromDisk(filename)
	if err != nil {
		panic(err)
	}

	eraseColor := uint8(float64(stringDarkness) * eraseFactor)
	fmt.Println(eraseColor)
	resultImage, _, instructions, length, err := stringer.Generate(target,
		stringer.WithPinCount(pinCount),
		stringer.WithDiameter(diameterMM/1000),
		stringer.WithLinesCount(nLines),
		stringer.WithStringDarkness(uint8(stringDarkness)),
		stringer.WithResolution(outputSize),
		stringer.WithEraseColor(color.RGBA{eraseColor, eraseColor, eraseColor, eraseColor}),
		// stringer.WithPaintColor(color.RGBA{0, 0, 0, 30}),
	)
	if err != nil {
		panic(err)
	}

	outFilenameBase, _ := strings.CutSuffix(path.Base(filename), path.Ext(filename))
	err = stringer.WriteInstructionsToDisk(path.Join(outDir, outFilenameBase+"_instructions.txt"), instructions, length)
	if err != nil {
		panic(err)
	}

	err = stringer.SaveImageToDisk(path.Join(outDir, outFilenameBase+"_stringer.png"), resultImage)
	if err != nil {
		panic(err)
	}

	err = htmlViewer.WriteInstructions(path.Join(outDir, outFilenameBase+"_instructions.html"), instructions)
	if err != nil {
		panic(err)
	}
}
