package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cduerm/stringpic/stringer"
)

var completed = binding.NewFloat()
var nLines = binding.NewFloat()
var nLinesSlider *widget.Slider
var lineDarkness = binding.NewFloat()
var lineDarknessSlider *widget.Slider
var size = binding.NewFloat()
var sizeSlider *widget.Slider

var targetImage, resultImage *image.RGBA
var targetImages, resultImages []*image.RGBA

var leftImage, rightImage *canvas.Image

var nLinesValues []int
var imgSize int = 600
var targetFilename string

var myApp = app.New()
var myWindow = myApp.NewWindow("Stringer by cduerm")
var fileOpenDialog = dialog.NewFileOpen(updateImage, myWindow)

func main() {
	if root, err := storage.ListerForURI(storage.NewFileURI("./input")); err == nil {
		fileOpenDialog.SetLocation(root)
	}

	nLinesSlider = widget.NewSliderWithData(0, 6000, nLines)
	nLinesSlider.Step = 100
	nLines.Set(2000)
	lineDarknessSlider = widget.NewSliderWithData(1, 255, lineDarkness)
	lineDarknessSlider.Step = 1
	lineDarkness.Set(30)
	sizeSlider = widget.NewSliderWithData(200, 1000, size)
	sizeSlider.Step = 10
	size.Set(500)

	if targetFilename == "" {
		// fileOpenDialog.Show()
	}
	leftImage, rightImage = images(targetFilename, imgSize)

	center := container.NewVBox(
		widget.NewButtonWithIcon("Open Image", theme.FolderOpenIcon(), func() {
			fileOpenDialog.Show()
		}),
		container.NewHBox(
			widget.NewLabel("number of strings"),
			layout.NewSpacer(),
			widget.NewLabelWithData(binding.FloatToStringWithFormat(nLines, "%.0f")),
		),
		container.New(&MinSizeLayout{300, 0}, nLinesSlider),

		container.NewHBox(
			widget.NewLabel("string darkness"),
			layout.NewSpacer(),
			widget.NewLabelWithData(binding.FloatToStringWithFormat(lineDarkness, "%.0f")),
		),
		lineDarknessSlider,

		container.NewHBox(
			widget.NewLabel("image resolution"),
			layout.NewSpacer(),
			widget.NewLabelWithData(binding.FloatToStringWithFormat(size, "%.0f")),
		),
		sizeSlider,
		// nLinesSlider,
	)

	nLines.AddListener(updateImageListener)
	completed.AddListener(updateImageListener)
	lineDarknessSlider.OnChangeEnded = func(f float64) {
		calculateImages()
	}
	sizeSlider.OnChangeEnded = func(f float64) {
		calculateImages()
	}

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

var updateImageListener = binding.NewDataListener(func() {
	f, _ := nLines.Get()
	c, err := completed.Get()
	if err != nil {
		panic(err)
	}
	if nLinesValues == nil || f > c {
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
	fmt.Printf("displaying image %d out of %d\n", idx, len(targetImages))
})

const (
	pinCount     = 160
	paddingPixel = 0
)

func updateImage(reader fyne.URIReadCloser, err error) {
	if err != nil {
		panic(err)
	}
	if reader == nil {
		return
	}
	targetFilename = reader.URI().Path()
	fmt.Println(targetFilename)

	li, ri := images(targetFilename, imgSize)
	leftImage.Image = li.Image
	rightImage.Image = ri.Image
	// updateImageListener.DataChanged()
	calculateImages()
}

func calculateImages() {
	nLinesValues = nil
	targetImages = nil
	resultImages = nil
	for val := int(nLinesSlider.Min); val < int(nLinesSlider.Max+1); val += int(nLinesSlider.Step) {
		nLinesValues = append(nLinesValues, val)
	}
	// fmt.Println(nLinesValues)

	calcSize, err := size.Get()
	if err != nil {
		panic(err)
	}
	target, result := copyImage(targetImage), copyImage(resultImage)

	target = stringer.RescaleImage(target, int(calcSize))
	result = stringer.RescaleImage(result, int(calcSize))

	pins := stringer.CalculatePins(pinCount, result.Bounds(), paddingPixel)
	allLines := stringer.CalculateLines(pins)

	lastLines := 0
	for _, nowLines := range nLinesValues {
		// fmt.Println(nowLines - lastLines)
		stringer.Generate(target, result, allLines, nowLines-lastLines, uint8(lineDarknessSlider.Value), 1)
		targetImages = append(targetImages, copyImage(target))
		resultImages = append(resultImages, copyImage(result))
		err := completed.Set(float64(nowLines))
		if err != nil {
			panic(err)
		}
		fmt.Println(nowLines)
		lastLines = nowLines
	}
}

func copyImage(img image.Image) *image.RGBA {
	new := image.NewRGBA(img.Bounds())
	draw.Draw(new, new.Bounds(), img, image.Point{}, draw.Over)
	return new
}

func images(filename string, size int) (imgLeft, imgRight *canvas.Image) {
	var maxSize = int(sizeSlider.Max)
	resultImage = image.NewRGBA(image.Rect(0, 0, maxSize, maxSize))
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	diskImage, err := stringer.OpenImageFromDisk(filename)
	if err != nil {
		// dialog.ShowError(fmt.Errorf("could not open the file %s: %w", filename, err), myWindow)
		diskImage = resultImage
	}
	targetImage = stringer.RescaleImage(diskImage, maxSize)

	leftImage := canvas.NewImageFromImage(targetImage)
	leftImage.FillMode = canvas.ImageFillContain
	leftImage.SetMinSize(fyne.NewSquareSize(float32(size)))

	rightImage := canvas.NewImageFromImage(resultImage)
	rightImage.SetMinSize(fyne.NewSquareSize(float32(size)))
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
