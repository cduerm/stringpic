package stringer

import (
	"image"
)

func LinePoints(from, to Pin) []image.Point {
	start := image.Point{int(from.X), int(from.Y)}
	end := image.Point{int(to.X), int(to.Y)}
	dx := end.Sub(start).X
	dy := end.Sub(start).Y

	points := make([]image.Point, 0)
	if abs(dx) >= abs(dy) {
		d := dx / abs(dx)
		for i := range abs(dx) + 1 {
			x := start.X + d*i
			y := start.Y + dy*(x-start.X)/dx
			points = append(points, image.Pt(x, y))
		}
	} else {
		d := dy / abs(dy)
		for i := range abs(dy) + 1 {
			y := start.Y + i*d
			x := start.X + dx*(y-start.Y)/dy
			points = append(points, image.Pt(x, y))
		}
	}
	return points
}

func Score(points []image.Point, target, result image.Image) float64 {
	var score float64 = 0
	for _, p := range points {
		rTarget, _, _, _ := target.At(p.X, p.Y).RGBA()
		rResult, _, _, _ := result.At(p.X, p.Y).RGBA()
		val := float64(rResult-rTarget) / (257 * 255)
		score += -val * val
	}
	return score / float64(len(points))
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}
