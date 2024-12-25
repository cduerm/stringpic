package main

import (
	"flag"
	"path"
	"strings"

	"github.com/cduerm/stringpic/htmlViewer"
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
	targetImage, resultImage, err := stringer.GetImages(outputSize, filename)
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

	err = htmlViewer.WriteInstructions(path.Join(outDir, outFilenameBase+"_instructions.html"), instructions)
	if err != nil {
		panic(err)
	}
}
