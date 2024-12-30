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
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cduerm/stringpic/stringer"
)

var completed = binding.NewFloat()

var nLines slider
var lineDarkness slider
var size slider

var eraseColor color.Color = color.RGBA{30, 30, 30, 30}

var targetImage, resultImage *image.RGBA
var targetImages, resultImages []*image.RGBA

var leftImage, rightImage *canvas.Image

var nLinesValues []int
var displayImgSize int = 600
var targetFilename string

var myApp = app.New()
var myWindow = myApp.NewWindow("Stringer by cduerm")
var fileOpenDialog = dialog.NewFileOpen(updateImage, myWindow)

func main() {
	if targetFilename == "" {
		fileOpenDialog.Show()
	}

	myWindow.SetContent(windowContent())
	myWindow.SetFixedSize(true)

	setupListeners()

	myWindow.ShowAndRun()
}

func init() {
	if root, err := storage.ListerForURI(storage.NewFileURI("./input")); err == nil {
		fileOpenDialog.SetLocation(root)
	}
}

func windowContent() *fyne.Container {
	nLines = NewSlider("number of steps", "%.0f", 0, 6000, 100, 2000)
	lineDarkness = NewSlider("string darkness", "%.0f", 1, 255, 1, 30)
	size = NewSlider("image resolution", "%.0f", 100, 1000, 10, 500)

	leftImage, rightImage = images(targetFilename, displayImgSize)
	cp := dialog.NewColorPicker("pick color", "", func(c color.Color) {
		eraseColor = c
		calculateImages()
	}, myWindow)
	cp.Advanced = true

	center := container.New(&MinSizeLayout{300, 0}, container.NewPadded(container.NewVBox(
		widget.NewButtonWithIcon("Open Image", theme.FolderOpenIcon(), func() {
			fileOpenDialog.Show()
		}),
		nLines.Container(),
		lineDarkness.Container(),
		size.Container(),
		widget.NewButton("erase color", func() { cp.Show() }),
	)))

	content := container.New(
		&LeftRightCenter{},
		leftImage,
		center,
		rightImage,
	)

	return content
}

func setupListeners() {
	updateImageListener := binding.NewDataListener(func() {
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
		// fmt.Printf("displaying image %d out of %d\n", idx, len(targetImages))
	})

	nLines.value.AddListener(updateImageListener)
	completed.AddListener(updateImageListener)
	lineDarkness.OnChangeEnded = func(f float64) {
		calculateImages()
	}
	size.OnChangeEnded = func(f float64) {
		calculateImages()
	}
}

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

	li, ri := images(targetFilename, displayImgSize)
	leftImage.Image = li.Image
	rightImage.Image = ri.Image
	calculateImages()
}

func calculateImages() {
	err := completed.Set(0)
	if err != nil {
		panic(err)
	}
	nLinesValues = nil
	targetImages = nil
	resultImages = nil
	for val := int(nLines.Min); val < int(nLines.Max+1); val += int(nLines.Step) {
		nLinesValues = append(nLinesValues, val)
	}

	calcSize, err := size.Get()
	if err != nil {
		panic(err)
	}
	target, result := copyImage(targetImage), copyImage(resultImage)

	target = stringer.RescaleImage(target, int(calcSize))
	result = stringer.RescaleImage(result, int(calcSize))

	lastLines := 0
	for _, nowLines := range nLinesValues {
		result, target, _, _, _ = stringer.GenerateWithOptions(target,
			stringer.WithResultImage(result),
			stringer.WithPinCount(pinCount),
			stringer.WithLinesCount(nowLines-lastLines),
			stringer.WithStringDarkness(uint8(lineDarkness.Value)),
			stringer.WithResolution(int(size.Value)),
			// stringer.WithEraseColor(eraseColor),
		)
		targetImages = append(targetImages, copyImage(target))
		resultImages = append(resultImages, copyImage(result))
		err := completed.Set(float64(nowLines))
		if err != nil {
			panic(err)
		}
		// fmt.Println(nowLines)
		lastLines = nowLines
	}
}

func copyImage(img image.Image) *image.RGBA {
	new := image.NewRGBA(img.Bounds())
	draw.Draw(new, new.Bounds(), img, image.Point{}, draw.Over)
	return new
}

func images(filename string, imgSize int) (imgLeft, imgRight *canvas.Image) {
	var maxSize = int(size.Max)
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
	leftImage.SetMinSize(fyne.NewSquareSize(float32(imgSize)))

	rightImage := canvas.NewImageFromImage(resultImage)
	rightImage.SetMinSize(fyne.NewSquareSize(float32(imgSize)))
	rightImage.FillMode = canvas.ImageFillContain
	return leftImage, rightImage
}
