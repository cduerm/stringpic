package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type slider struct {
	*widget.Slider
	value  binding.Float
	label  string
	format string
}

func NewSlider(label, format string, min, max, step, start float64) slider {
	s := slider{}
	s.value = binding.NewFloat()
	s.label = label
	s.format = format
	s.Slider = widget.NewSliderWithData(min, max, s.value)
	s.Slider.Step = step
	s.value.Set(start)
	return s
}

func (s slider) Container() *fyne.Container {
	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel(s.label),
			layout.NewSpacer(),
			widget.NewLabelWithData(binding.FloatToStringWithFormat(s.value, s.format)),
		),
		s.Slider,
	)
}

func (s slider) Get() (float64, error) {
	return s.value.Get()
}
