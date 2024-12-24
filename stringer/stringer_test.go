package stringer

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"testing"
)

func BenchmarkLinePoins(b *testing.B) {
	p1 := Pin{0, 0}
	p2 := Pin{100, 500}
	for range b.N {
		LinePoints(p1, p2)
	}
}

func BenchmarkCalculateLines(b *testing.B) {
	pins := CalculatePins(400, image.Rect(0, 0, 512, 512), 10)
	b.ResetTimer()
	for range b.N {
		CalculateLines(pins)
	}
}

func BenchmarkColorOver(b *testing.B) {
	var c color.Color
	c1 := color.RGBA{0, 120, 200, 255}
	c2 := color.RGBA{255, 120, 100, 50}
	b.ResetTimer()
	for range b.N {
		c = ColorOver(c1, c2)
	}
	_ = c
}

func BenchmarkColorOverUint8(b *testing.B) {
	var c color.Color
	c1 := color.RGBA{0, 120, 200, 255}
	c2 := color.RGBA{255, 120, 100, 50}
	b.ResetTimer()
	for range b.N {
		c = ColorOverUint8(c1, c2)
	}
	_ = c
}

func BenchmarkPaintLineNaive(b *testing.B) {
	bounds := image.Rect(0, 0, 512, 512)
	resultImage := image.NewRGBA(bounds)
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)
	c1 := color.RGBA{255, 120, 100, 50}
	line := LinePoints(Pin{0, 0}, Pin{100, 500})
	b.ResetTimer()

	for range b.N {
		for _, p := range line {
			resultImage.Set(p.X, p.Y, c1)
		}
	}
}

func BenchmarkPaintLineWithPix(b *testing.B) {
	bounds := image.Rect(0, 0, 512, 512)
	resultImage := image.NewRGBA(bounds)
	draw.Draw(resultImage, resultImage.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)
	c1 := color.RGBA{255, 120, 100, 50}
	line := LinePoints(Pin{0, 0}, Pin{100, 500})
	b.ResetTimer()

	for range b.N {
		pix := resultImage.Pix
		for _, p := range line {
			idx := resultImage.PixOffset(p.X, p.Y)
			pix[idx+0] = c1.R
			pix[idx+1] = c1.G
			pix[idx+2] = c1.B
			pix[idx+3] = c1.A
		}
	}
}

func TestColorOver(t *testing.T) {
	testData := make([]color.RGBA, 100)
	for i := range testData {
		c := color.RGBA{}
		c.R = uint8(rand.Intn(255))
		c.G = uint8(rand.Intn(255))
		c.B = uint8(rand.Intn(255))
		c.A = uint8(rand.Intn(255))
		testData[i] = c
	}
	for _, c2 := range testData[:] {
		c1 := color.RGBA{255, 255, 255, 255}
		over1 := ColorOver(c1, c2)
		over2 := ColorOverUint8(c1, c2)
		if over1 != over2 {
			t.Fail()
			t.Log(c1, c2, over1, over2)
		}
	}
}

func TestPaintLine(t *testing.T) {
	for range 100 {
		c1 := color.RGBA{}
		c1.R = uint8(rand.Intn(255))
		c1.G = uint8(rand.Intn(255))
		c1.B = uint8(rand.Intn(255))
		c1.A = uint8(rand.Intn(255))

		c2 := color.RGBA{}
		c2.R = uint8(rand.Intn(255))
		c2.G = uint8(rand.Intn(255))
		c2.B = uint8(rand.Intn(255))
		c2.A = uint8(rand.Intn(255))

		bounds := image.Rect(0, 0, 512, 1)
		img1 := image.NewRGBA(bounds)
		draw.Draw(img1, img1.Bounds(), image.NewUniform(c2), image.Point{}, draw.Over)
		img2 := image.NewRGBA(bounds)
		draw.Draw(img2, img2.Bounds(), image.NewUniform(c2), image.Point{}, draw.Over)

		line := make([]image.Point, 512)
		for i := range line {
			line[i].X = i
		}

		PixelOver(img1, line, c1)
		for _, p := range line {
			img2.Set(p.X, p.Y, ColorOver(img2.At(p.X, p.Y), c1))
		}

		for i := range img1.Pix {
			if img1.Pix[i] != img2.Pix[i] {
				t.Fail()
			}
		}
	}
}

func TestScores(t *testing.T) {
	for range 1000 {
		pts := LinePoints(
			Pin{rand.Float64() * 512, rand.Float64() * 512},
			Pin{rand.Float64() * 512, rand.Float64() * 512})

		bounds := image.Rect(0, 0, 512, 512)

		target := image.NewRGBA(bounds)
		for i := range 512 {
			for j := range 512 {
				target.Pix[target.PixOffset(i, j)+0] = uint8(rand.Int())
				target.Pix[target.PixOffset(i, j)+1] = uint8(rand.Int())
				target.Pix[target.PixOffset(i, j)+2] = uint8(rand.Int())
				target.Pix[target.PixOffset(i, j)+3] = uint8(255)
			}
		}

		result := image.NewRGBA(bounds)
		draw.Draw(result, result.Bounds(), image.White, image.Point{}, draw.Over)
		// PixelOver(target, pts, color.RGBA{0, 0, 0, 255})
		// SaveImageToDisk("testScoresResult.png", result)
		// SaveImageToDisk("testScoresTarget.png", target)

		s1 := Score(pts, target, result)
		s2 := Score_old(pts, target, result)
		if s1 != s2 {
			t.Fail()
		}
	}
}
