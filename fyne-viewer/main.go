package main

import (
	"fmt"
	"image"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/cduerm/stringpic/stringer"
)

var nLines = binding.NewFloat()
var nLinesSlider *widget.Slider

var targetImage, resultImage *image.RGBA
var targetImages, resultImages []*image.RGBA
var completed float64 = 0
var nLinesValues []int

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Stringer by cduerm")

	leftImage, rightImage := images()
	nLinesSlider = widget.NewSliderWithData(0, 6000, nLines)
	nLinesSlider.Step = 100
	nLines.Set(0)
	center := container.NewVBox(
		widget.NewLabel("number of strings"),
		container.New(&MinSizeLayout{300, 0}, nLinesSlider),
		// nLinesSlider,
		widget.NewLabelWithData(binding.FloatToStringWithFormat(nLines, "%.0f")),
	)

	nLines.AddListener(binding.NewDataListener(func() {
		f, _ := nLines.Get()
		if nLinesValues == nil || f > completed {
			return
		}

		idx := 0
		for i, val := range nLinesValues {
			if int(f) <= val {
				idx = i
				break
			}
		}
		leftImage.Image = targetImages[idx]
		leftImage.Refresh()
		rightImage.Image = resultImages[idx]
		rightImage.Refresh()
	}))
	content := container.New(
		&LeftRightCenter{},
		leftImage,
		center,
		rightImage,
	)
	myWindow.SetContent(content)

	go calculateImages()

	myWindow.SetFixedSize(true)
	myWindow.ShowAndRun()
}

const (
	pinCount     = 160
	paddingPixel = 0
)

func calculateImages() {
	for val := int(nLinesSlider.Min); val < int(nLinesSlider.Max+1); val += int(nLinesSlider.Step) {
		nLinesValues = append(nLinesValues, val)
	}
	fmt.Println(nLinesValues)

	var err error
	targetImage, resultImage, err = stringer.GetImages(600, "input/flower.png")
	if err != nil {
		panic(err)
	}
	pins := stringer.CalculatePins(pinCount, resultImage.Bounds(), paddingPixel)
	allLines := stringer.CalculateLines(pins)

	lastLines := 0
	for _, nowLines := range nLinesValues {
		stringer.Generate(targetImage, resultImage, allLines, nowLines-lastLines, uint8(10), 1)
		targetImages = append(targetImages, copyImage(targetImage))
		resultImages = append(resultImages, copyImage(resultImage))
		completed = float64(nowLines)
		fmt.Println(nowLines)
		lastLines = nowLines
	}
}

func copyImage(img image.Image) *image.RGBA {
	new := image.NewRGBA(img.Bounds())
	draw.Draw(new, new.Bounds(), img, image.Point{}, draw.Over)
	return new
}

func images() (imgLeft, imgRight *canvas.Image) {
	var err error
	targetImage, resultImage, err = stringer.GetImages(512, "input/flower.png")
	if err != nil {
		panic(err)
	}
	leftImage := canvas.NewImageFromImage(targetImage)
	leftImage.FillMode = canvas.ImageFillContain
	leftImage.SetMinSize(fyne.NewSquareSize(600))
	rightImage := canvas.NewImageFromImage(resultImage)
	rightImage.SetMinSize(fyne.NewSquareSize(600))
	rightImage.FillMode = canvas.ImageFillContain
	return leftImage, rightImage
}

type MinSizeLayout struct {
	w, h float32
}

func (l *MinSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	for _, o := range objects {
		childSize := o.MinSize()
		l.w = max(l.w, childSize.Width)
		l.h = max(l.h, childSize.Height)
	}
	return fyne.NewSize(l.w, l.h)
}

func (l *MinSizeLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(containerSize)
	}
}

type LeftRightCenter struct{}

func (l *LeftRightCenter) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		w += childSize.Width
		h = max(h, childSize.Height)
	}
	return fyne.NewSize(w, h)
}

func (l *LeftRightCenter) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	centerWidth := objects[1].MinSize().Width
	sideWidth := (containerSize.Width - centerWidth) / 2

	objects[0].Resize(fyne.NewSquareSize(sideWidth))
	objects[0].Move(pos)
	pos.X += sideWidth
	objects[1].Resize(objects[1].MinSize())
	objects[1].Move(pos)
	pos.X += objects[1].MinSize().Width
	objects[2].Resize(fyne.NewSquareSize(sideWidth))
	objects[2].Move(pos)
	pos.X += sideWidth
}
