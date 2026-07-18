package render

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func countPixels(img *image.RGBA, pred func(c color.RGBA) bool) int {
	n := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if pred(img.RGBAAt(x, y)) {
				n++
			}
		}
	}
	return n
}

func TestRasterFilledCircleCoverage(t *testing.T) {
	const r = 50.0
	ras := NewRaster(200, 200)
	ctx := NewContext(ras)
	ctx.SetFillColor(color.White)
	ctx.DrawCircle(100, 100, r)
	ctx.Fill()

	// Center pixel is pure white.
	assert.Equal(t, color.RGBA{255, 255, 255, 255}, ras.Image().RGBAAt(100, 100))
	// A pixel outside the circle is untouched.
	assert.Equal(t, color.RGBA{}, ras.Image().RGBAAt(10, 10))
	// Coverage approximates the area.
	painted := countPixels(ras.Image(), func(c color.RGBA) bool { return c.A > 127 })
	assert.InDelta(t, math.Pi*r*r, float64(painted), 0.02*math.Pi*r*r)
}

func TestRasterDeterminism(t *testing.T) {
	draw := func() *Raster {
		ras := NewRaster(120, 120)
		ctx := NewContext(ras)
		ctx.SetFillColor(color.RGBA{200, 50, 50, 255})
		ctx.SetStrokeColor(color.RGBA{50, 50, 200, 255})
		ctx.SetStrokeWidth(3)
		ctx.DrawCircle(60, 60, 40)
		ctx.FillStroke()
		return ras
	}
	assert.True(t, bytes.Equal(draw().Image().Pix, draw().Image().Pix))
}

func TestRasterStrokeLine(t *testing.T) {
	ras := NewRaster(100, 100)
	ctx := NewContext(ras)
	ctx.SetStrokeColor(color.White)
	ctx.SetStrokeWidth(4)
	ctx.MoveTo(10, 50)
	ctx.LineTo(90, 50)
	ctx.Stroke()

	assert.Equal(t, uint8(255), ras.Image().RGBAAt(50, 50).A)
	assert.Equal(t, uint8(255), ras.Image().RGBAAt(50, 51).A)
	// Well off the 4px-wide line: untouched.
	assert.Equal(t, uint8(0), ras.Image().RGBAAt(50, 60).A)
	// Butt cap by default: nothing before the start point.
	assert.Equal(t, uint8(0), ras.Image().RGBAAt(7, 50).A)
}

func TestRasterDashedStroke(t *testing.T) {
	ras := NewRaster(100, 100)
	ctx := NewContext(ras)
	ctx.SetStrokeColor(color.White)
	ctx.SetStrokeWidth(2)
	ctx.SetDash(10, 10)
	ctx.MoveTo(0, 50)
	ctx.LineTo(100, 50)
	ctx.Stroke()

	// Middle of the first dash is painted, middle of the first gap is not.
	assert.Equal(t, uint8(255), ras.Image().RGBAAt(5, 50).A)
	assert.Equal(t, uint8(0), ras.Image().RGBAAt(15, 50).A)
	assert.Equal(t, uint8(255), ras.Image().RGBAAt(25, 50).A)
}

func TestRasterEvenOdd(t *testing.T) {
	scene := func(rule FillRule) *Raster {
		ras := NewRaster(100, 100)
		ctx := NewContext(ras)
		ctx.SetFillRule(rule)
		ctx.SetFillColor(color.White)
		// Two concentric same-winding rects.
		ctx.DrawRectangle(10, 10, 80, 80)
		ctx.DrawRectangle(30, 30, 40, 40)
		ctx.Fill()
		return ras
	}
	// Even-odd: the inner rect is a hole. Non-zero: solid.
	assert.Equal(t, uint8(0), scene(EvenOdd).Image().RGBAAt(50, 50).A)
	assert.Equal(t, uint8(255), scene(NonZero).Image().RGBAAt(50, 50).A)
	// Both fill the ring between the rects.
	assert.Equal(t, uint8(255), scene(EvenOdd).Image().RGBAAt(20, 50).A)
}

func TestRasterTranslucentOverlap(t *testing.T) {
	ras := NewRaster(100, 100)
	ctx := NewContext(ras)
	ctx.SetFillColor(color.RGBA{255, 0, 0, 128})
	ctx.DrawRectangle(10, 10, 50, 50)
	ctx.Fill()
	ctx.SetFillColor(color.RGBA{0, 0, 255, 128})
	ctx.DrawRectangle(40, 40, 50, 50)
	ctx.Fill()

	// Overlap region composites blue over red.
	c := ras.Image().RGBAAt(50, 50)
	assert.Greater(t, c.B, c.R)
	assert.Greater(t, c.R, uint8(0))
	assert.Greater(t, int(c.A), 128)
}

func TestRasterTransformStack(t *testing.T) {
	ras := NewRaster(100, 100)
	ctx := NewContext(ras)
	ctx.SetFillColor(color.White)
	ctx.Push()
	ctx.Translate(50, 50)
	ctx.Rotate(math.Pi / 4)
	ctx.DrawRectangle(-20, -20, 40, 40)
	ctx.Fill()
	ctx.Pop()

	// The rotated square's corner reaches (50, 50-20*sqrt2): inside at
	// (50, 25), outside the unrotated square's corner at (32, 32).
	assert.Equal(t, uint8(255), ras.Image().RGBAAt(50, 25).A)
	assert.Equal(t, uint8(0), ras.Image().RGBAAt(32, 32).A)
	// Pop restored the transform: drawing at origin lands at origin.
	ctx.DrawRectangle(0, 0, 5, 5)
	ctx.Fill()
	assert.Equal(t, uint8(255), ras.Image().RGBAAt(2, 2).A)
}

func TestRasterReuseNoGhosting(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	ras := NewRasterFromImage(img)
	ctx := NewContext(ras)

	ctx.SetFillColor(color.White)
	ctx.DrawCircle(15, 15, 10)
	ctx.Fill()
	assert.Equal(t, uint8(255), img.RGBAAt(15, 15).A)

	// Frame 2: clear and draw elsewhere; frame 1 must not ghost through.
	ctx.Clear(color.Black)
	ctx.DrawCircle(35, 35, 10)
	ctx.Fill()
	assert.Equal(t, color.RGBA{0, 0, 0, 255}, img.RGBAAt(15, 15))
	assert.Equal(t, color.RGBA{255, 255, 255, 255}, img.RGBAAt(35, 35))
}

func TestRasterDrawImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := range 10 {
		for x := range 10 {
			src.SetRGBA(x, y, color.RGBA{0, 255, 0, 255})
		}
	}
	ras := NewRaster(50, 50)
	ctx := NewContext(ras)
	ctx.DrawImage(src, 20, 20)
	assert.Equal(t, color.RGBA{0, 255, 0, 255}, ras.Image().RGBAAt(25, 25))
	assert.Equal(t, color.RGBA{}, ras.Image().RGBAAt(15, 15))

	// Non-integer transform goes through the resampling path.
	ctx.Push()
	ctx.Translate(0.5, 0)
	ctx.DrawImage(src, 35, 35)
	ctx.Pop()
	assert.Equal(t, uint8(255), ras.Image().RGBAAt(40, 40).A)
}

func TestContextStrokePreserve(t *testing.T) {
	ras := NewRaster(60, 60)
	ctx := NewContext(ras)
	ctx.SetStrokeColor(color.White)
	ctx.SetFillColor(color.RGBA{255, 0, 0, 255})
	ctx.DrawRectangle(10, 10, 40, 40)
	ctx.StrokePreserve()
	ctx.Fill()
	// Fill happened (path preserved through StrokePreserve).
	assert.Equal(t, color.RGBA{255, 0, 0, 255}, ras.Image().RGBAAt(30, 30))
	// Path cleared after Fill: another Stroke paints nothing new.
	ctx.SetStrokeColor(color.RGBA{0, 0, 255, 255})
	ctx.Stroke()
	assert.Equal(t, color.RGBA{255, 0, 0, 255}, ras.Image().RGBAAt(30, 30))
}
