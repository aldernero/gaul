package render

import (
	"bytes"
	"encoding/xml"
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func svgString(s *SVG) string {
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func assertWellFormed(t *testing.T, doc string) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(doc))
	for {
		_, err := dec.Token()
		if err != nil {
			require.Contains(t, err.Error(), "EOF")
			return
		}
	}
}

func TestSVGDocumentShape(t *testing.T) {
	svg := NewSVG(800, 600)
	ctx := NewContext(svg)
	ctx.SetFillColor(color.RGBA{255, 0, 0, 255})
	ctx.DrawRectangle(10, 10, 100, 50)
	ctx.Fill()

	doc := svgString(svg)
	assert.Contains(t, doc, `viewBox="0 0 800 600"`)
	assert.Contains(t, doc, `fill="#ff0000"`)
	assert.Contains(t, doc, `M10 10`)
	assert.Contains(t, doc, "Z")
	assertWellFormed(t, doc)
}

func TestSVGCurvesStayNative(t *testing.T) {
	svg := NewSVG(200, 200)
	ctx := NewContext(svg)
	ctx.SetStrokeColor(color.Black)
	ctx.MoveTo(0, 0)
	ctx.QuadTo(50, 100, 100, 0)
	ctx.CubicTo(120, 50, 180, 50, 200, 0)
	ctx.Stroke()

	doc := svgString(svg)
	assert.Contains(t, doc, "Q50 100 100 0")
	assert.Contains(t, doc, "C120 50 180 50 200 0")
	// Curves must never be flattened into line runs.
	assert.Equal(t, 0, strings.Count(doc, "L"))
}

func TestSVGStrokesAreRealStrokes(t *testing.T) {
	svg := NewSVG(100, 100)
	ctx := NewContext(svg)
	ctx.SetStrokeColor(color.RGBA{0, 0, 255, 255})
	ctx.SetStrokeWidth(2.5)
	ctx.SetLineCap(RoundCap)
	ctx.SetLineJoin(RoundJoin)
	ctx.SetDash(4, 2)
	ctx.SetDashOffset(1)
	ctx.MoveTo(10, 10)
	ctx.LineTo(90, 90)
	ctx.Stroke()

	doc := svgString(svg)
	assert.Contains(t, doc, `fill="none"`)
	assert.Contains(t, doc, `stroke="#0000ff"`)
	assert.Contains(t, doc, `stroke-width="2.5"`)
	assert.Contains(t, doc, `stroke-linecap="round"`)
	assert.Contains(t, doc, `stroke-linejoin="round"`)
	assert.Contains(t, doc, `stroke-dasharray="4 2"`)
	assert.Contains(t, doc, `stroke-dashoffset="1"`)
	assertWellFormed(t, doc)
}

func TestSVGFillRuleAndOpacity(t *testing.T) {
	svg := NewSVG(100, 100)
	ctx := NewContext(svg)
	ctx.SetFillRule(EvenOdd)
	ctx.SetFillColor(color.NRGBA{0, 255, 0, 128})
	ctx.DrawRectangle(0, 0, 50, 50)
	ctx.Fill()

	doc := svgString(svg)
	assert.Contains(t, doc, `fill-rule="evenodd"`)
	assert.Contains(t, doc, `fill="#00ff00"`)
	assert.Contains(t, doc, `fill-opacity="0.502"`)
}

func TestSVGTransformsAreBakedIn(t *testing.T) {
	svg := NewSVG(100, 100)
	ctx := NewContext(svg)
	ctx.SetFillColor(color.Black)
	ctx.Push()
	ctx.Translate(50, 50)
	ctx.DrawRectangle(0, 0, 10, 10)
	ctx.Fill()
	ctx.Pop()

	doc := svgString(svg)
	assert.Contains(t, doc, "M50 50")
	assert.NotContains(t, doc, "transform")
}

func TestSVGPrecision(t *testing.T) {
	svg := NewSVG(100, 100)
	svg.Precision = 1
	ctx := NewContext(svg)
	ctx.SetFillColor(color.Black)
	ctx.MoveTo(1.23456, 2.34567)
	ctx.LineTo(3, 4)
	ctx.ClosePath()
	ctx.Fill()

	doc := svgString(svg)
	assert.Contains(t, doc, "M1.2 2.3")
}

func TestDualBackend(t *testing.T) {
	// One drawing session produces both raster and SVG output.
	ras := NewRaster(100, 100)
	svg := NewSVG(100, 100)
	ctx := NewContext(ras, svg)
	ctx.SetFillColor(color.White)
	ctx.SetStrokeColor(color.RGBA{255, 0, 0, 255})
	ctx.SetStrokeWidth(2)
	ctx.DrawCircle(50, 50, 30)
	ctx.FillStroke()

	// Raster got pixels.
	assert.Equal(t, color.RGBA{255, 255, 255, 255}, ras.Image().RGBAAt(50, 50))
	// SVG got a native-curve path with both fill and stroke.
	doc := svgString(svg)
	assert.Contains(t, doc, `fill="#ffffff"`)
	assert.Contains(t, doc, `stroke="#ff0000"`)
	assert.Contains(t, doc, "C")
	assertWellFormed(t, doc)
}
