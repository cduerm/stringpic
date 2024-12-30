package main

import "fyne.io/fyne/v2"

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
