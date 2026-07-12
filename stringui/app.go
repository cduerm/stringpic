package stringui

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"

	"github.com/cduerm/stringpic/htmlViewer"
	"github.com/cduerm/stringpic/stringer"
)

type StringerApp struct {
	Window   *app.Window
	Explorer *explorer.Explorer
	Theme    *material.Theme
	Mu       sync.Mutex

	// Sliders
	DarknessSlider   widget.Float
	EraseSlider      widget.Float
	ResolutionSlider widget.Float
	PinsSlider       widget.Float
	StepsSlider      widget.Float

	// Buttons
	OpenButton        widget.Clickable
	RecalculateButton widget.Clickable
	SaveButton        widget.Clickable

	// Editor for diameter
	DiameterEditor widget.Editor

	// Scrollable list for controls
	ControlsList layout.List

	// State variables
	State struct {
		Calculating         bool
		CancelRecalculation chan struct{}
		CompletedLines      int
		TargetImages        []*image.RGBA
		ResultImages        []*image.RGBA
		PreviewImages       []*image.RGBA
		Instructions        []int
		Lengths             []float64
		SelectedId          int
		ImageDiameterMM     float64

		// Sliders tracked state
		DarknessVal   float32
		EraseVal      float32
		PinsVal       float32
		ResolutionVal float32
		StepsVal      float32

		LastDarknessSlider   float32
		LastEraseSlider      float32
		LastPinsSlider       float32
		LastResolutionSlider float32
		LastStepsSlider      float32

		HasImage      bool
		OriginalImage image.Image

		// GPU cache for textures
		LeftImage    image.Image
		LeftImageOp  *paint.ImageOp
		RightImage   image.Image
		RightImageOp *paint.ImageOp
	}
}

func NewStringerApp() *StringerApp {
	s := &StringerApp{}

	s.Window = new(app.Window)
	s.Window.Option(
		app.Title("Stringer by cduerm"),
	)
	s.Explorer = explorer.NewExplorer(s.Window)

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{
		Bg:         color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff}, // slate-900
		Fg:         color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}, // slate-50
		ContrastBg: color.NRGBA{R: 0x0e, G: 0xa5, B: 0xe9, A: 0xff}, // sky-500
		ContrastFg: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, // white
	}
	s.Theme = th

	s.State.DarknessVal = 50
	s.State.EraseVal = 0.2
	s.State.PinsVal = 160
	s.State.ResolutionVal = 300
	s.State.StepsVal = 2500
	s.State.ImageDiameterMM = 220
	s.State.HasImage = false

	s.DarknessSlider.Value = (50.0 - 1.0) / (255.0 - 1.0)
	s.EraseSlider.Value = 0.2 / 2.0
	s.PinsSlider.Value = (160.0 - 10.0) / (600.0 - 10.0)
	s.ResolutionSlider.Value = (300.0 - 100.0) / (1000.0 - 100.0)
	s.StepsSlider.Value = 2500.0 / 6000.0

	s.State.LastDarknessSlider = s.DarknessSlider.Value
	s.State.LastEraseSlider = s.EraseSlider.Value
	s.State.LastPinsSlider = s.PinsSlider.Value
	s.State.LastResolutionSlider = s.ResolutionSlider.Value
	s.State.LastStepsSlider = s.StepsSlider.Value

	s.DiameterEditor.SetText("220")
	s.DiameterEditor.SingleLine = true

	s.ControlsList.Axis = layout.Vertical

	return s
}

func (s *StringerApp) ShowAndRun() {
	go func() {
		if err := s.Run(); err != nil {
			log.Fatal("Gio app run error:", err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (s *StringerApp) Run() error {
	var ops op.Ops
	for {
		e := s.Window.Event()
		s.Explorer.ListenEvents(e)
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			s.updateState(gtx)
			s.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (s *StringerApp) updateState(gtx layout.Context) {
	s.Mu.Lock()

	// Update sliders state
	if val := s.DarknessSlider.Value; val != s.State.LastDarknessSlider {
		s.State.LastDarknessSlider = val
		s.State.DarknessVal = float32(math.Round(float64(1 + val*(255-1))))
	}
	if val := s.EraseSlider.Value; val != s.State.LastEraseSlider {
		s.State.LastEraseSlider = val
		s.State.EraseVal = val * 2
	}
	if val := s.PinsSlider.Value; val != s.State.LastPinsSlider {
		s.State.LastPinsSlider = val
		s.State.PinsVal = float32(math.Round(float64(10+val*(600-10))/10) * 10)
	}
	if val := s.ResolutionSlider.Value; val != s.State.LastResolutionSlider {
		s.State.LastResolutionSlider = val
		s.State.ResolutionVal = float32(math.Round(float64(100+val*(1000-100))/10) * 10)
	}
	if val := s.StepsSlider.Value; val != s.State.LastStepsSlider {
		s.State.LastStepsSlider = val
		s.State.StepsVal = float32(math.Round(float64(val*6000)/100) * 100)
	}

	// Clamp selected steps based on calculation completion
	selectedId := int(math.Round(float64(s.State.StepsVal / 100)))
	maxId := len(s.State.PreviewImages) - 1
	if maxId < 0 {
		maxId = 0
	}
	if selectedId > maxId {
		selectedId = maxId
	}
	s.State.SelectedId = selectedId

	// Cache left/right images as GPU ops to avoid uploading every frame
	if len(s.State.TargetImages) > selectedId {
		target := s.State.TargetImages[selectedId]
		if target != s.State.LeftImage {
			s.State.LeftImage = target
			op := paint.NewImageOp(target)
			s.State.LeftImageOp = &op
		}
	} else {
		s.State.LeftImage = nil
		s.State.LeftImageOp = nil
	}

	if len(s.State.PreviewImages) > selectedId {
		preview := s.State.PreviewImages[selectedId]
		if preview != s.State.RightImage {
			s.State.RightImage = preview
			op := paint.NewImageOp(preview)
			s.State.RightImageOp = &op
		}
	} else {
		s.State.RightImage = nil
		s.State.RightImageOp = nil
	}

	// Actions execution (from Clickable events)
	calculating := s.State.Calculating
	hasImage := s.State.HasImage
	s.Mu.Unlock()

	if !calculating {
		for s.OpenButton.Clicked(gtx) {
			s.OpenFileDialog()
		}
		if hasImage {
			for s.SaveButton.Clicked(gtx) {
				s.SaveInstructionsDialog()
			}
		}
	}

	if hasImage {
		for s.RecalculateButton.Clicked(gtx) {
			s.Recalculate()
		}
	}
}

func (s *StringerApp) OpenFileDialog() {
	go func() {
		rc, err := s.Explorer.ChooseFile("png", "jpg", "jpeg")
		if err != nil {
			log.Println("Error choosing file:", err)
			return
		}
		if rc == nil {
			return
		}
		defer rc.Close()

		img, _, err := image.Decode(rc)
		if err != nil {
			log.Println("Error decoding image:", err)
			return
		}

		s.Mu.Lock()
		s.State.HasImage = true
		s.State.OriginalImage = img
		s.Mu.Unlock()

		s.InitRecalculation(img)
		s.Recalculate()
	}()
}

func (s *StringerApp) InitRecalculation(img image.Image) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	res := int(s.State.ResolutionVal)
	target := stringer.RescaleImage(img, res)
	empty := image.NewRGBA(target.Bounds())
	draw.Draw(empty, empty.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	emptyPreview := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	draw.Draw(emptyPreview, emptyPreview.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	s.State.TargetImages = make([]*image.RGBA, 1, s.LinesVariants()+1)
	s.State.TargetImages[0] = target
	s.State.ResultImages = make([]*image.RGBA, 1, s.LinesVariants()+1)
	s.State.ResultImages[0] = empty
	s.State.PreviewImages = make([]*image.RGBA, 1, s.LinesVariants()+1)
	s.State.PreviewImages[0] = emptyPreview
}

func (s *StringerApp) Recalculate() {
	s.Mu.Lock()
	if s.State.Calculating {
		s.Cancel()
		s.Mu.Unlock()
		for {
			s.Mu.Lock()
			calc := s.State.Calculating
			s.Mu.Unlock()
			if !calc {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		s.Mu.Lock()
	}

	if s.State.OriginalImage == nil {
		s.Mu.Unlock()
		return
	}

	res := int(s.State.ResolutionVal)
	target := stringer.RescaleImage(s.State.OriginalImage, res)
	empty := image.NewRGBA(target.Bounds())
	draw.Draw(empty, empty.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	emptyPreview := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	draw.Draw(emptyPreview, emptyPreview.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	s.State.TargetImages = make([]*image.RGBA, 1, s.LinesVariants()+1)
	s.State.TargetImages[0] = target
	s.State.ResultImages = make([]*image.RGBA, 1, s.LinesVariants()+1)
	s.State.ResultImages[0] = empty
	s.State.PreviewImages = make([]*image.RGBA, 1, s.LinesVariants()+1)
	s.State.PreviewImages[0] = emptyPreview
	s.State.Instructions = make([]int, 0)
	s.State.Lengths = make([]float64, 0)

	s.State.Calculating = true
	s.State.CompletedLines = 0
	s.State.CancelRecalculation = make(chan struct{})

	pins := int(s.State.PinsVal)
	darkness := uint8(s.State.DarknessVal)
	erase := float64(s.State.EraseVal)
	resolution := int(s.State.ResolutionVal)
	diameter := s.State.ImageDiameterMM

	cancelChan := s.State.CancelRecalculation
	s.Mu.Unlock()

	go func() {
		currentLines := 0
		bigStep := 100
		miniStep := 20

		for i := 0; i < 60; i++ {
			select {
			case <-cancelChan:
				s.Mu.Lock()
				s.State.Calculating = false
				s.Mu.Unlock()
				s.Window.Invalidate()
				return
			default:
				s.Mu.Lock()
				startImg := s.State.TargetImages[i]
				resultImg := s.State.ResultImages[i]
				s.Mu.Unlock()

				tempStart := startImg
				tempResult := resultImg
				var accumulatedInstructions []int
				var accumulatedLength float64
				var lastPins []stringer.Pin

				for j := 0; j < bigStep/miniStep; j++ {
					select {
					case <-cancelChan:
						s.Mu.Lock()
						s.State.Calculating = false
						s.Mu.Unlock()
						s.Window.Invalidate()
						return
					default:
					}

					result, err := stringer.Generate(
						tempStart,
						stringer.WithResultImage(tempResult),
						stringer.WithLinesCount(miniStep),
						stringer.WithPinCount(pins),
						stringer.WithStringDarkness(darkness),
						stringer.WithEraseFactor(erase),
						stringer.WithResolution(resolution),
						stringer.WithDiameter(diameter/1000),
						stringer.WithoutPreview(),
					)
					if err != nil {
						log.Println("Error in stringer.Generate:", err)
						s.Mu.Lock()
						s.State.Calculating = false
						s.Mu.Unlock()
						s.Window.Invalidate()
						return
					}

					accumulatedInstructions = append(accumulatedInstructions, result.Instructions...)
					accumulatedLength += result.StringLength
					lastPins = result.Pins
					tempStart = result.StartImage
					tempResult = result.EndImage

					s.Mu.Lock()
					s.State.CompletedLines = currentLines + (j+1)*miniStep
					s.Mu.Unlock()
					s.Window.Invalidate()

					time.Sleep(10 * time.Millisecond)
				}

				s.Mu.Lock()
				select {
				case <-cancelChan:
					s.State.Calculating = false
					s.Mu.Unlock()
					s.Window.Invalidate()
					return
				default:
				}

				s.State.TargetImages = append(s.State.TargetImages, tempStart)
				s.State.ResultImages = append(s.State.ResultImages, tempResult)
				s.State.Instructions = append(s.State.Instructions, accumulatedInstructions...)

				prevPreview := s.State.PreviewImages[len(s.State.PreviewImages)-1]
				preview := stringer.PreviewOver(lastPins, accumulatedInstructions, prevPreview)
				s.State.PreviewImages = append(s.State.PreviewImages, preview)

				s.State.Lengths = append(s.State.Lengths, accumulatedLength)

				currentLines += bigStep
				s.State.CompletedLines = currentLines
				s.Mu.Unlock()

				s.Window.Invalidate()
			}
		}

		s.Mu.Lock()
		s.State.Calculating = false
		s.Mu.Unlock()
		s.Window.Invalidate()
	}()
}

func (s *StringerApp) Cancel() {
	if s.State.CancelRecalculation != nil {
		close(s.State.CancelRecalculation)
		s.State.CancelRecalculation = nil
	}
}

func (s *StringerApp) SaveInstructionsDialog() {
	go func() {
		wc, err := s.Explorer.CreateFile("stringer_output.zip")
		if err != nil {
			log.Println("Error creating zip file:", err)
			return
		}
		if wc == nil {
			return
		}
		defer wc.Close()

		s.Mu.Lock()
		selectedId := s.State.SelectedId

		if selectedId >= len(s.State.PreviewImages) {
			selectedId = len(s.State.PreviewImages) - 1
		}

		var instructions []int
		if len(s.State.Instructions) > 0 {
			limit := int(s.State.StepsVal) + 1
			if limit > len(s.State.Instructions) {
				limit = len(s.State.Instructions)
			}
			instructions = s.State.Instructions[:limit]
		} else {
			instructions = []int{0}
		}

		length := 0.0
		limitId := selectedId
		if limitId > len(s.State.Lengths) {
			limitId = len(s.State.Lengths)
		}
		for _, l := range s.State.Lengths[:limitId] {
			length += l
		}

		previewImage := s.State.PreviewImages[selectedId]
		s.Mu.Unlock()

		zipWriter := zip.NewWriter(wc)

		// 1. Text instructions
		txtWriter, err := zipWriter.Create("instructions.txt")
		if err == nil {
			text := stringer.InstructionsText(instructions, length)
			txtWriter.Write([]byte(text))
		}

		// 2. Image instructions
		imgWriter, err := zipWriter.Create("stringer.png")
		if err == nil {
			png.Encode(imgWriter, previewImage)
		}

		// 3. HTML viewer instructions
		htmlWriter, err := zipWriter.Create("instructions.html")
		if err == nil {
			htmlViewer.FillTemplate(instructions, htmlWriter)
		}

		zipWriter.Close()
	}()
}

func (s *StringerApp) LinesVariants() int {
	return 60 // 6000 / 100
}

func (s *StringerApp) Layout(gtx layout.Context) layout.Dimensions {
	// Fill window background with dark slate color
	paint.Fill(gtx.Ops, s.Theme.Palette.Bg)

	// Layout left image, center card panel, and right image horizontally
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			s.Mu.Lock()
			leftOp := s.State.LeftImageOp
			s.Mu.Unlock()
			if leftOp != nil {
				layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return widget.Image{
						Src: *leftOp,
						Fit: widget.Contain,
					}.Layout(gtx)
				})
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(320)
				gtx.Constraints.Max.X = gtx.Dp(320)
				return drawCardBackground(gtx, color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}, 12, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return s.ControlsList.Layout(gtx, 10, func(gtx layout.Context, index int) layout.Dimensions {
							return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								return s.layoutControlItem(gtx, index)
							})
						})
					})
				})
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			s.Mu.Lock()
			rightOp := s.State.RightImageOp
			s.Mu.Unlock()
			if rightOp != nil {
				layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return widget.Image{
						Src: *rightOp,
						Fit: widget.Contain,
					}.Layout(gtx)
				})
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
	)
}

func (s *StringerApp) layoutControlItem(gtx layout.Context, index int) layout.Dimensions {
	th := s.Theme

	s.Mu.Lock()
	calculating := s.State.Calculating
	hasImage := s.State.HasImage
	darknessVal := s.State.DarknessVal
	eraseVal := s.State.EraseVal
	pinsVal := s.State.PinsVal
	resVal := s.State.ResolutionVal
	stepsVal := s.State.StepsVal
	completed := s.State.CompletedLines
	s.Mu.Unlock()

	switch index {
	case 0: // Open Image Button
		btn := material.Button(th, &s.OpenButton, "Open Image")
		var layoutGtx = gtx
		if calculating {
			layoutGtx.Source = gtx.Source.Disabled()
			btn.Background = color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xff} // slate-600
			btn.Color = color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}      // slate-400
		}
		return btn.Layout(layoutGtx)

	case 1: // Darkness Slider
		return s.layoutSlider(gtx, th, "String Darkness", darknessVal, "%.0f", &s.DarknessSlider)

	case 2: // Erase Ratio Slider
		return s.layoutSlider(gtx, th, "Erase Ratio", eraseVal, "%3.2f", &s.EraseSlider)

	case 3: // Pins Slider
		return s.layoutSlider(gtx, th, "Number of Pins", pinsVal, "%.0f", &s.PinsSlider)

	case 4: // Resolution Slider
		return s.layoutSlider(gtx, th, "Image Resolution", resVal, "%.0f", &s.ResolutionSlider)

	case 5: // Diameter Input Editor
		return s.layoutDiameterInput(gtx, th)

	case 6: // Recalculate Button
		btn := material.Button(th, &s.RecalculateButton, "Recalculate Images")
		var layoutGtx = gtx
		disabled := !hasImage
		if disabled {
			layoutGtx.Source = gtx.Source.Disabled()
			btn.Background = color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xff} // slate-600
			btn.Color = color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}      // slate-400
		}
		return btn.Layout(layoutGtx)

	case 7: // Progress Bar
		var progress float32
		max := 6000
		if max > 0 {
			progress = float32(completed) / float32(max)
		}
		percent := float64(completed) / float64(max) * 100
		if math.IsNaN(percent) {
			percent = 0
		}
		progressText := fmt.Sprintf("%d/%d (%.0f%%)", completed, max, percent)

		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				bar := material.ProgressBar(th, progress)
				bar.Color = th.Palette.ContrastBg
				bar.TrackColor = color.NRGBA{R: 0x33, G: 0x41, B: 0x55, A: 0xff}
				return bar.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(material.Body2(th, progressText).Layout),
		)

	case 8: // Steps Slider
		return s.layoutSlider(gtx, th, "Number of Steps", stepsVal, "%.0f", &s.StepsSlider)

	case 9: // Save Button
		btn := material.Button(th, &s.SaveButton, "Save Instructions")
		var layoutGtx = gtx
		disabled := !hasImage || calculating
		if disabled {
			layoutGtx.Source = gtx.Source.Disabled()
			btn.Background = color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xff} // slate-600
			btn.Color = color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}      // slate-400
		}
		return btn.Layout(layoutGtx)

	default:
		return layout.Dimensions{}
	}
}

func (s *StringerApp) layoutSlider(gtx layout.Context, th *material.Theme, label string, val float32, format string, state *widget.Float) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(material.Body2(th, label).Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
				layout.Rigid(material.Body2(th, fmt.Sprintf(format, val)).Layout),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			slider := material.Slider(th, state)
			return slider.Layout(gtx)
		}),
	)
}

func (s *StringerApp) layoutDiameterInput(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(material.Body2(th, "Diameter [mm]").Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(80)
			gtx.Constraints.Max.X = gtx.Dp(80)

			for {
				ev, ok := s.DiameterEditor.Update(gtx)
				if !ok {
					break
				}
				if _, ok := ev.(widget.SubmitEvent); ok {
					// handle submit if needed
				}
			}

			txt := s.DiameterEditor.Text()
			s.Mu.Lock()
			if val, err := strconv.ParseFloat(txt, 64); err == nil && val > 0 {
				s.State.ImageDiameterMM = val
			}
			s.Mu.Unlock()

			ed := material.Editor(th, &s.DiameterEditor, "220")
			ed.TextSize = th.TextSize

			return drawCardBackground(gtx, color.NRGBA{R: 0x33, G: 0x41, B: 0x55, A: 0xff}, 4, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(6)).Layout(gtx, ed.Layout)
			})
		}),
	)
}

func drawCardBackground(gtx layout.Context, bgCol color.NRGBA, radius int, w layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
			paint.FillShape(gtx.Ops, bgCol, clip.UniformRRect(bounds, gtx.Dp(unit.Dp(radius))).Op(gtx.Ops))
			return layout.Dimensions{Size: bounds.Max}
		}),
		layout.Stacked(w),
	)
}
