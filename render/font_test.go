package render

import (
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/gofont/goregular"
)

func loadGoRegular(t *testing.T) *Font {
	t.Helper()
	f, err := LoadFont(goregular.TTF, "Go Regular")
	require.NoError(t, err)
	return f
}

func TestFontMeasure(t *testing.T) {
	f := loadGoRegular(t)
	w := f.Measure("Hello", 14)
	assert.Greater(t, w, 10.0)
	assert.Less(t, w, 100.0)
	// Longer strings are wider, larger sizes are wider.
	assert.Greater(t, f.Measure("Hello, world", 14), w)
	assert.Greater(t, f.Measure("Hello", 28), w)

	ascent, descent := f.Metrics(14)
	assert.Greater(t, ascent, 0.0)
	assert.Greater(t, descent, 0.0)
}

func TestRasterDrawString(t *testing.T) {
	ras := NewRaster(200, 50)
	ctx := NewContext(ras)
	ctx.SetFillColor(color.White)
	ctx.SetFont(loadGoRegular(t), 24)
	ctx.DrawString("Hi", 10, 35)

	painted := countPixels(ras.Image(), func(c color.RGBA) bool { return c.A > 0 })
	assert.Greater(t, painted, 20)
	// Text lands near the anchor, not elsewhere.
	found := false
	for y := 10; y < 40 && !found; y++ {
		for x := 10; x < 50 && !found; x++ {
			found = ras.Image().RGBAAt(x, y).A > 0
		}
	}
	assert.True(t, found)
}

func TestSVGDrawString(t *testing.T) {
	svg := NewSVG(200, 50)
	ctx := NewContext(svg)
	ctx.SetFillColor(color.Black)
	ctx.SetFont(loadGoRegular(t), 14)
	ctx.DrawString("a < b & c", 10, 30)

	doc := svgString(svg)
	assert.Contains(t, doc, `font-family="Go Regular"`)
	assert.Contains(t, doc, `font-size="14px"`)
	assert.Contains(t, doc, "a &lt; b &amp; c")
	assertWellFormed(t, doc)
}

func TestDrawStringTransformedAnchor(t *testing.T) {
	svg := NewSVG(200, 200)
	ctx := NewContext(svg)
	ctx.SetFillColor(color.Black)
	ctx.SetFont(loadGoRegular(t), 14)
	ctx.Push()
	ctx.Translate(100, 100)
	ctx.DrawString("x", 0, 0)
	ctx.Pop()

	doc := svgString(svg)
	assert.Contains(t, doc, `x="100"`)
	assert.Contains(t, doc, `y="100"`)
}

func TestDrawStringWithoutFontIsNoop(t *testing.T) {
	svg := NewSVG(100, 100)
	ctx := NewContext(svg)
	ctx.DrawString("nothing", 10, 10)
	assert.Equal(t, 0.0, ctx.MeasureString("nothing"))
	doc := svgString(svg)
	assert.False(t, strings.Contains(doc, "<text"))
}
