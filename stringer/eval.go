package stringer

import (
	"image"
	"image/color"
)

func LinePoints(from, to Pin) []image.Point {
	start := image.Point{int(from.X), int(from.Y)}
	end := image.Point{int(to.X), int(to.Y)}
	dx := end.Sub(start).X
	dy := end.Sub(start).Y
	n := max(abs(dx), abs(dy))

	points := make([]image.Point, 0, n+1)
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

type ScoreFunction func([]image.Point, *image.RGBA, *image.RGBA) float64

func Score(points []image.Point, target, result *image.RGBA) float64 {
	var score float64 = 0
	for _, p := range points {
		targetOffset := target.PixOffset(p.X, p.Y)
		resultOffset := result.PixOffset(p.X, p.Y)
		deltaR := (float64(result.Pix[resultOffset+0]) - float64(target.Pix[targetOffset+0])) / 255
		deltaG := (float64(result.Pix[resultOffset+1]) - float64(target.Pix[targetOffset+1])) / 255
		deltaB := (float64(result.Pix[resultOffset+2]) - float64(target.Pix[targetOffset+2])) / 255
		score += evalDiff(deltaR) + evalDiff(deltaG) + evalDiff(deltaB)
	}
	return score / float64(len(points))
}

// not yet properly implemented, is meant to allow for stings of different colors
func ScoreWithColors(paintColor, eraseColor color.Color) ScoreFunction {
	// pr, pg, pb, pa := paintColor.RGBA()
	// paintR := float64(uint8(pr / pa))
	// paintG := float64(uint8(pg / pa))
	// paintB := float64(uint8(pb / pa))
	// er, eg, eb, ea := eraseColor.RGBA()
	// ea = max(ea, 1)
	// eraseR := min(1, float64(er)/float64(ea))
	// eraseG := min(1, float64(eg)/float64(ea))
	// eraseB := min(1, float64(eb)/float64(ea))

	return func(points []image.Point, target, result *image.RGBA) float64 {
		var score float64 = 0
		for _, p := range points {
			targetOffset := target.PixOffset(p.X, p.Y)
			resultOffset := result.PixOffset(p.X, p.Y)
			deltaR := (float64(result.Pix[resultOffset+0]) - float64(target.Pix[targetOffset+0])) / 255
			deltaG := (float64(result.Pix[resultOffset+1]) - float64(target.Pix[targetOffset+1])) / 255
			deltaB := (float64(result.Pix[resultOffset+2]) - float64(target.Pix[targetOffset+2])) / 255
			score += evalDiff(deltaR) + evalDiff(deltaG) + evalDiff(deltaB)
		}
		return score / float64(len(points))
	}
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}

func evalDiff(delta float64) float64 {
	if delta < 0 { // result pixel is darker than the target --> penalize
		return -delta * delta
	}
	return delta * delta
}
