package render

import "image/color"

// LineCap is the shape drawn at the ends of open stroked subpaths (and at
// the ends of each dash).
type LineCap uint8

const (
	// ButtCap ends the stroke exactly at the endpoint.
	ButtCap LineCap = iota
	// RoundCap ends the stroke with a semicircle.
	RoundCap
	// SquareCap extends the stroke by half the stroke width.
	SquareCap
)

// LineJoin is the shape drawn where stroked segments meet.
//
// The raster backend has no miter joiner (a limitation of freetype/raster,
// shared with fogleman/gg): MiterJoin renders as BevelJoin in PNG output,
// while the SVG backend emits true miter joins.
type LineJoin uint8

const (
	// MiterJoin extends the outer edges until they meet (SVG only; the
	// raster backend falls back to BevelJoin).
	MiterJoin LineJoin = iota
	// RoundJoin rounds the corner with a circular arc.
	RoundJoin
	// BevelJoin cuts the corner with a straight edge.
	BevelJoin
)

// FillRule determines which regions of a self-intersecting path are inside.
type FillRule uint8

const (
	// NonZero is the non-zero winding rule (the default).
	NonZero FillRule = iota
	// EvenOdd is the even-odd rule.
	EvenOdd
)

// Style describes how a path is painted. A nil or fully transparent
// FillColor disables filling; a nil or fully transparent StrokeColor (or a
// non-positive StrokeWidth) disables stroking. Stroke widths are in device
// pixels and are not scaled by the Context's transform.
type Style struct {
	FillColor   color.Color
	StrokeColor color.Color
	// Dashes is the on/off dash pattern in pixels; empty means solid.
	Dashes      []float64
	StrokeWidth float64
	// MiterLimit is the SVG stroke-miterlimit value. It only affects SVG
	// output; zero means the SVG default (4).
	MiterLimit float64
	DashOffset float64
	Cap        LineCap
	Join       LineJoin
	FillRule   FillRule
}

// visible reports whether c contributes any paint.
func visible(c color.Color) bool {
	if c == nil {
		return false
	}
	_, _, _, a := c.RGBA()
	return a > 0
}

// hasFill reports whether the style produces a fill.
func (s Style) hasFill() bool {
	return visible(s.FillColor)
}

// hasStroke reports whether the style produces a stroke.
func (s Style) hasStroke() bool {
	return s.StrokeWidth > 0 && visible(s.StrokeColor)
}
