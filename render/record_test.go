package render

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecorderReplayMatchesDirect(t *testing.T) {
	scene := func(ctx *Context) {
		ctx.Clear(color.Black)
		ctx.SetFillColor(color.RGBA{200, 60, 60, 255})
		ctx.SetStrokeColor(color.White)
		ctx.SetStrokeWidth(3)
		ctx.DrawCircle(60, 60, 40)
		ctx.FillStroke()
		ctx.SetDash(6, 4)
		ctx.MoveTo(10, 110)
		ctx.LineTo(110, 110)
		ctx.Stroke()
	}

	// Draw directly into a raster.
	direct := NewRaster(120, 120)
	scene(NewContext(direct))

	// Draw into a recorder only, then replay into a fresh raster.
	rec := NewRecorder(120, 120)
	scene(NewContext(rec))
	replayed := NewRaster(120, 120)
	rec.Replay(replayed)

	assert.True(t, bytes.Equal(direct.Image().Pix, replayed.Image().Pix))
	// Replay again — recording is reusable.
	replayed2 := NewRaster(120, 120)
	rec.Replay(replayed2)
	assert.True(t, bytes.Equal(direct.Image().Pix, replayed2.Image().Pix))
}

func TestRecorderReplayToSVG(t *testing.T) {
	rec := NewRecorder(100, 100)
	ctx := NewContext(rec)
	ctx.SetStrokeColor(color.Black)
	ctx.MoveTo(0, 0)
	ctx.QuadTo(50, 100, 100, 0)
	ctx.Stroke()

	svg := NewSVG(100, 100)
	rec.Replay(svg)
	var buf strings.Builder
	_, err := svg.WriteTo(&buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Q50 100 100 0")
}

func TestRecorderIsolatedFromContextReuse(t *testing.T) {
	// The Context reuses its path buffer between paints; the recorder must
	// keep deep copies.
	rec := NewRecorder(100, 100)
	ctx := NewContext(rec)
	ctx.SetFillColor(color.White)
	ctx.DrawRectangle(0, 0, 10, 10)
	ctx.Fill()
	ctx.DrawCircle(50, 50, 20)
	ctx.Fill()

	svg := NewSVG(100, 100)
	rec.Replay(svg)
	var buf strings.Builder
	if _, err := svg.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	doc := buf.String()
	assert.Contains(t, doc, "M0 0") // rect survived
	assert.Contains(t, doc, "C")    // circle beziers
	assert.Equal(t, 2, strings.Count(doc, "<path"))
}

func TestRecorderReset(t *testing.T) {
	rec := NewRecorder(50, 50)
	ctx := NewContext(rec)
	ctx.SetFillColor(color.White)
	ctx.DrawRectangle(0, 0, 10, 10)
	ctx.Fill()
	assert.False(t, rec.Empty())
	rec.Reset()
	assert.True(t, rec.Empty())
}

func TestRasterScale(t *testing.T) {
	// The same logical scene at scale 2 lands in a 2x image with everything
	// doubled: geometry, stroke width, dashes.
	logical := func(r *Raster) {
		ctx := NewContext(r)
		// Context sees logical 100x100 regardless of scale.
		w, h := r.Size()
		assert.Equal(t, 100.0, w)
		assert.Equal(t, 100.0, h)
		ctx.SetFillColor(color.White)
		ctx.DrawRectangle(25, 25, 50, 50)
		ctx.Fill()
	}

	r1 := NewRaster(100, 100)
	logical(r1)
	r2 := NewRaster(200, 200)
	r2.SetScale(2)
	logical(r2)

	// Filled areas scale by 4x.
	count := func(r *Raster) int {
		return countPixels(r.Image(), func(c color.RGBA) bool { return c.A > 127 })
	}
	assert.InDelta(t, 4*count(r1), count(r2), 0.02*float64(4*count(r1)))
	// The rect's center maps to the scaled center.
	assert.Equal(t, uint8(255), r2.Image().RGBAAt(100, 100).A)
	// Outside the scaled rect is empty.
	assert.Equal(t, uint8(0), r2.Image().RGBAAt(40, 100).A)
}

func TestRasterScaleStrokeWidth(t *testing.T) {
	r := NewRaster(200, 200)
	r.SetScale(2)
	ctx := NewContext(r)
	ctx.SetStrokeColor(color.White)
	ctx.SetStrokeWidth(4) // logical px -> 8 physical px wide
	ctx.MoveTo(10, 50)
	ctx.LineTo(90, 50)
	ctx.Stroke()

	// Line center at physical y=100; 8 px wide spans ~[96, 104).
	assert.Equal(t, uint8(255), r.Image().RGBAAt(100, 100).A)
	assert.Equal(t, uint8(255), r.Image().RGBAAt(100, 97).A)
	assert.Equal(t, uint8(0), r.Image().RGBAAt(100, 110).A)
}

func TestRecorderReplayToScaledRaster(t *testing.T) {
	// A recorded frame replayed at scale 2 doubles cleanly — the sketchy
	// save-at-higher-DPI path.
	rec := NewRecorder(100, 100)
	ctx := NewContext(rec)
	ctx.SetFillColor(color.White)
	ctx.DrawCircle(50, 50, 30)
	ctx.Fill()

	hi := NewRaster(200, 200)
	hi.SetScale(2)
	rec.Replay(hi)
	assert.Equal(t, uint8(255), hi.Image().RGBAAt(100, 100).A)
	assert.Equal(t, uint8(255), hi.Image().RGBAAt(100, 45).A) // inside r=60 disk
	assert.Equal(t, uint8(0), hi.Image().RGBAAt(100, 30).A)   // outside
}
