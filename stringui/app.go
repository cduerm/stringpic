package stringui

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cduerm/stringpic/htmlViewer"
	"github.com/cduerm/stringpic/stringer"
)

type StringerApp struct {
	fyne.App
	Window         fyne.Window
	FileOpenDialog *dialog.FileDialog
	FileSaveDialog *dialog.FileDialog
	Widgets        struct {
		Lines             SliderWithLabel
		Darkness          SliderWithLabel
		Erase             SliderWithLabel
		Resolution        SliderWithLabel
		Pins              SliderWithLabel
		OpenButton        *widget.Button
		RecalculateButton *widget.Button
		SaveButton        *widget.Button
		ProgressBar       *widget.ProgressBar
		LeftImage         *canvas.Image
		RightImage        *canvas.Image
		DiameterInput     *fyne.Container
	}
	State struct {
		Calculating         bool
		CancelRecalculation chan struct{}
		CompletedLines      int
		TargetImages        []*image.RGBA
		ResultImages        []*image.RGBA
		Instructions        []int
		Lengths             []float64
		SelectedId          int
		ImageDiameterMM     float64
	}
	Options struct {
		ImageDisplaySize int
	}
}

func (s *StringerApp) setupContent() {
	s.FileOpenDialog = dialog.NewFileOpen(s.openFileCallback, s.Window)
	s.FileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
	if root, err := storage.ListerForURI(storage.NewFileURI("./input")); err == nil {
		s.FileOpenDialog.SetLocation(root)
	}

	s.FileSaveDialog = dialog.NewFileSave(s.saveFileCallback, s.Window)

	s.Widgets.Lines = NewSliderWithLabel("Number of Steps", "%.0f", 0, 6000, 100, 2500)
	s.Widgets.Lines.OnChanged = func(f float64) {
		s.State.SelectedId = int((f - s.Widgets.Lines.Min + 1) / s.Widgets.Lines.Step)
		if int(f) <= s.State.CompletedLines {
			s.setImages(int((f - s.Widgets.Lines.Min + 1) / s.Widgets.Lines.Step))
		}
		s.Widgets.Lines.value.Set(f)
	}
	s.Widgets.Darkness = NewSliderWithLabel("String Darkness", "%.0f", 1, 255, 1, 50)
	s.Widgets.Erase = NewSliderWithLabel("Erase Ratio", "%3.2f", 0, 2, 0.05, 0)
	s.Widgets.Resolution = NewSliderWithLabel("Image Resolution", "%.0f", 100, 1000, 10, 500)
	s.Widgets.Pins = NewSliderWithLabel("Number of Pins", "%.0f", 10, 600, 10, 160)
	entry := widget.NewEntry()
	entry.Validator = func(s string) error {
		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New("must be a valid number")
		}
		return nil
	}
	entry.OnChanged = func(text string) {
		if entry.Validate() != nil {
			return
		}
		diameter, _ := strconv.ParseFloat(text, 64)
		s.State.ImageDiameterMM = diameter
	}
	entry.SetPlaceHolder("220")
	s.Widgets.DiameterInput = container.NewGridWithColumns(2,
		widget.NewLabel("Diameter [mm]"),
		entry)

	s.Widgets.OpenButton = widget.NewButtonWithIcon(
		"Open Image",
		theme.FolderOpenIcon(),
		func() {
			s.FileOpenDialog.Show()
		},
	)
	s.Widgets.SaveButton = widget.NewButtonWithIcon(
		"Save Instructions",
		theme.DocumentSaveIcon(),
		func() {
			s.FileSaveDialog.Show()
		},
	)
	s.Widgets.RecalculateButton = widget.NewButton(
		"Recalculate Images",
		func() {
			s.Recalculate()
		},
	)
	s.Widgets.RecalculateButton.Disable()

	s.Widgets.ProgressBar = widget.NewProgressBar()
	s.Widgets.ProgressBar.Min = s.Widgets.Lines.Min
	s.Widgets.ProgressBar.Max = s.Widgets.Lines.Max
	s.Widgets.ProgressBar.TextFormatter = func() string {
		max := int(s.Widgets.ProgressBar.Max)
		val := int(s.Widgets.ProgressBar.Value)
		percent := float64(val) / float64(max) * 100
		if math.IsNaN(percent) {
			percent = 0
		}

		return fmt.Sprintf("%4d/%d (%.0f %%)", val, max, percent)
	}
	s.Widgets.ProgressBar.SetValue(0)

	s.Options.ImageDisplaySize = 600
	s.Widgets.LeftImage = &canvas.Image{}
	s.Widgets.LeftImage.FillMode = canvas.ImageFillContain
	s.Widgets.LeftImage.SetMinSize(fyne.NewSquareSize(float32(s.Options.ImageDisplaySize)))

	s.Widgets.RightImage = &canvas.Image{}
	s.Widgets.RightImage.FillMode = canvas.ImageFillContain
	s.Widgets.RightImage.SetMinSize(fyne.NewSquareSize(float32(s.Options.ImageDisplaySize)))

	s.Window.SetContent(
		container.New(
			&LeftRightCenter{},
			s.Widgets.LeftImage,
			container.New(&MinSizeLayout{300, 0},
				container.NewVBox(
					s.Widgets.OpenButton,
					s.Widgets.Darkness.Container(),
					s.Widgets.Erase.Container(),
					s.Widgets.Pins.Container(),
					s.Widgets.Resolution.Container(),
					s.Widgets.DiameterInput,
					s.Widgets.RecalculateButton,
					s.Widgets.ProgressBar,
					layout.NewSpacer(),
					s.Widgets.Lines.Container(),
					layout.NewSpacer(),
					s.Widgets.SaveButton,
				)),
			s.Widgets.RightImage,
		),
	)
}

func (s *StringerApp) openFileCallback(reader fyne.URIReadCloser, err error) {
	if err != nil {
		panic(err)
	}
	if reader == nil {
		return
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}
	s.Widgets.RecalculateButton.Enable()

	target := stringer.RescaleImage(img, int(s.Widgets.Resolution.Value))
	empty := image.NewRGBA(target.Bounds())
	draw.Draw(empty, empty.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	s.State.TargetImages = make([]*image.RGBA, 1, s.LinesVariants())
	s.State.TargetImages[0] = target
	s.State.ResultImages = make([]*image.RGBA, 1, s.LinesVariants())
	s.State.ResultImages[0] = empty

	s.setImages(0)

	s.Recalculate()
}

func (s *StringerApp) saveFileCallback(writer fyne.URIWriteCloser, err error) {
	if s.State.Calculating {
		d := dialog.NewInformation("Calculation", "The Calculation mus be completed before saving", s.Window)
		d.Show()
		return
	}
	if err != nil {
		panic(err)
	}
	if writer == nil {
		return
	}
	uri := writer.URI()
	filepath := uri.Path()
	err = os.Remove(filepath)
	if err != nil {
		panic(err)
	}

	instructions := s.State.Instructions[:int(s.Widgets.Lines.Value)+1]
	length := 0.0
	for _, l := range s.State.Lengths[:s.State.SelectedId] {
		length += l
	}
	resultImage := s.Widgets.RightImage.Image

	err = stringer.WriteInstructionsToDisk(filepath+"_instructions.txt", instructions, length)
	if err != nil {
		panic(err)
	}

	err = stringer.SaveImageToDisk(filepath+"_stringer.png", resultImage)
	if err != nil {
		panic(err)
	}

	err = htmlViewer.WriteInstructions(filepath+"_instructions.html", instructions)
	if err != nil {
		panic(err)
	}
}

func (s *StringerApp) setImages(i int) {
	if len(s.State.TargetImages) == 0 {
		return
	}
	if i > len(s.State.TargetImages) {
		panic("target image not yet calculated")
	}
	s.Widgets.LeftImage.Image = s.State.TargetImages[i]
	fyne.Do(func() { s.Widgets.LeftImage.Refresh() })

	if i > len(s.State.ResultImages) {
		panic("result image not yet calculated")
	}
	s.Widgets.RightImage.Image = s.State.ResultImages[i]
	fyne.Do(func() { s.Widgets.RightImage.Refresh() })
}

func (s *StringerApp) Recalculate() {
	if s.State.Calculating {
		s.Cancel()
		for s.State.Calculating {
			time.Sleep(50 * time.Millisecond)
		}
	}

	s.State.TargetImages = s.State.TargetImages[:1]
	s.State.ResultImages = s.State.ResultImages[:1]
	s.State.Instructions = s.State.Instructions[:0]
	s.State.Lengths = s.State.Lengths[:0]

	s.State.Calculating = true
	s.State.CompletedLines = 0
	s.State.CancelRecalculation = make(chan struct{})

	pins := int(s.Widgets.Pins.Value)
	darkness := uint8(s.Widgets.Darkness.Value)
	erase := s.Widgets.Erase.Value
	resolution := int(s.Widgets.Resolution.Value)
	go func() {
		currentLines := s.Widgets.Lines.Min
		for i := range s.LinesVariants() {
			select {
			case <-s.State.CancelRecalculation:
				s.State.Calculating = false
				return
			default:
				nextLines := min(currentLines+s.Widgets.Lines.Step, s.Widgets.Lines.Max)
				result, target, instructions, length, err := stringer.Generate(
					s.State.TargetImages[i],
					stringer.WithResultImage(s.State.ResultImages[i]),
					stringer.WithLinesCount(int(nextLines-currentLines)),
					stringer.WithPinCount(pins),
					stringer.WithStringDarkness(darkness),
					stringer.WithEraseFactor(erase),
					stringer.WithResolution(resolution),
					stringer.WithDiameter(s.State.ImageDiameterMM/1000),
				)
				if err != nil {
					panic(err)
				}

				s.State.TargetImages = append(s.State.TargetImages, target)
				s.State.ResultImages = append(s.State.ResultImages, result)
				s.State.Instructions = append(s.State.Instructions, instructions...)
				s.State.Lengths = append(s.State.Lengths, length)

				if set := s.Widgets.Lines.Value; set > currentLines && set <= nextLines {
					s.setImages(i + 1)
				}
				currentLines = nextLines
				s.State.CompletedLines = int(currentLines)
				fyne.Do(func() { s.Widgets.ProgressBar.SetValue(currentLines) })
			}
		}
		s.State.Calculating = false
	}()
}

func (s *StringerApp) Cancel() {
	close(s.State.CancelRecalculation)
}

func (s *StringerApp) LinesVariants() int {
	count := 0
	for i := s.Widgets.Lines.Min; i <= s.Widgets.Lines.Max; i += s.Widgets.Lines.Step {
		count += 1
	}
	return count
}

func NewStringerApp() (s *StringerApp) {
	s = new(StringerApp)
	s.App = app.NewWithID("com.duermann.stringpic")
	s.Window = s.App.NewWindow("Stringer by cduerm")

	s.Window.SetIcon(resourceAppiconPng)

	s.setupContent()

	return s
}

func (s *StringerApp) ShowAndRun() {
	s.Window.ShowAndRun()
}
