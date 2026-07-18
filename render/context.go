package render

import (
	"image"
	"image/color"
)

// ctxState is the saveable drawing state managed by Push/Pop.
type ctxState struct {
	m        Matrix
	style    Style
	font     *Font
	fontSize float64
}

// Context is the drawing interface. It carries a transform stack and current
// style, builds paths, and hands finished paths to its Renderers.
//
// Transforms are applied to coordinates as the path is built: MoveTo, LineTo,
// QuadTo, and CubicTo map their points through the current transform before
// storing them, so Renderers only ever see device coordinates. A consequence
// (shared with fogleman/gg) is that stroke widths are in device pixels and
// are not affected by Scale.
//
// The zero style is fill black, stroke black, stroke width 1.
type Context struct {
	renderers []Renderer
	w, h      float64
	state     ctxState
	stack     []ctxState
	path      Path
}

// NewContext returns a Context that draws to every given Renderer. The
// Context's Width and Height come from the first renderer. NewContext panics
// if no renderers are given.
func NewContext(renderers ...Renderer) *Context {
	if len(renderers) == 0 {
		panic("render.NewContext: at least one Renderer is required")
	}
	w, h := renderers[0].Size()
	return &Context{
		renderers: renderers,
		w:         w,
		h:         h,
		state: ctxState{
			m: Identity(),
			style: Style{
				FillColor:   color.Black,
				StrokeColor: color.Black,
				StrokeWidth: 1,
			},
		},
	}
}

// NewContextForImage returns a Context rasterizing into the caller-owned
// image. It is shorthand for NewContext(NewRasterFromImage(img)).
func NewContextForImage(img *image.RGBA) *Context {
	return NewContext(NewRasterFromImage(img))
}

// Width returns the drawing surface width in pixels.
func (c *Context) Width() float64 { return c.w }

// Height returns the drawing surface height in pixels.
func (c *Context) Height() float64 { return c.h }

// Push saves the current transform and style onto the state stack.
func (c *Context) Push() {
	saved := c.state
	saved.style.Dashes = append([]float64(nil), c.state.style.Dashes...)
	c.stack = append(c.stack, saved)
}

// Pop restores the most recently pushed transform and style. Popping an
// empty stack is a no-op.
func (c *Context) Pop() {
	if len(c.stack) == 0 {
		return
	}
	c.state = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
}

// Translate moves the origin by (dx, dy).
func (c *Context) Translate(dx, dy float64) {
	c.Transform(Translation(dx, dy))
}

// Rotate rotates the coordinate system by angle radians about the origin.
// Positive angles rotate clockwise on screen (y-down).
func (c *Context) Rotate(angle float64) {
	c.Transform(Rotation(angle))
}

// RotateAbout rotates the coordinate system by angle radians about the point
// (x, y).
func (c *Context) RotateAbout(angle, x, y float64) {
	c.Transform(Translation(x, y).Mul(Rotation(angle)).Mul(Translation(-x, -y)))
}

// Scale scales the coordinate system by (sx, sy) about the origin.
func (c *Context) Scale(sx, sy float64) {
	c.Transform(Scaling(sx, sy))
}

// ScaleAbout scales the coordinate system by (sx, sy) about the point (x, y).
func (c *Context) ScaleAbout(sx, sy, x, y float64) {
	c.Transform(Translation(x, y).Mul(Scaling(sx, sy)).Mul(Translation(-x, -y)))
}

// Shear shears the coordinate system by sx horizontally and sy vertically.
func (c *Context) Shear(sx, sy float64) {
	c.Transform(Shearing(sx, sy))
}

// Transform post-multiplies the current transform by m, applying m in the
// current (already transformed) coordinate system.
func (c *Context) Transform(m Matrix) {
	c.state.m = c.state.m.Mul(m)
}

// SetTransform replaces the current transform with m.
func (c *Context) SetTransform(m Matrix) {
	c.state.m = m
}

// Matrix returns the current transform.
func (c *Context) Matrix() Matrix {
	return c.state.m
}

// Identity resets the current transform to the identity.
func (c *Context) Identity() {
	c.state.m = Identity()
}

// SetFillColor sets the fill color. nil disables filling.
func (c *Context) SetFillColor(col color.Color) {
	c.state.style.FillColor = col
}

// SetStrokeColor sets the stroke color. nil disables stroking.
func (c *Context) SetStrokeColor(col color.Color) {
	c.state.style.StrokeColor = col
}

// SetColor sets both the fill and stroke colors.
func (c *Context) SetColor(col color.Color) {
	c.state.style.FillColor = col
	c.state.style.StrokeColor = col
}

// SetStrokeWidth sets the stroke width in pixels.
func (c *Context) SetStrokeWidth(w float64) {
	c.state.style.StrokeWidth = w
}

// SetLineCap sets the stroke cap style.
func (c *Context) SetLineCap(cap LineCap) {
	c.state.style.Cap = cap
}

// SetLineJoin sets the stroke join style. See LineJoin for the raster
// backend's miter limitation.
func (c *Context) SetLineJoin(join LineJoin) {
	c.state.style.Join = join
}

// SetMiterLimit sets the SVG stroke-miterlimit value (SVG output only).
func (c *Context) SetMiterLimit(l float64) {
	c.state.style.MiterLimit = l
}

// SetDash sets the stroke dash pattern in pixels. Calling with no arguments
// makes strokes solid again.
func (c *Context) SetDash(dashes ...float64) {
	c.state.style.Dashes = dashes
}

// SetDashOffset sets the distance into the dash pattern at which strokes
// start.
func (c *Context) SetDashOffset(offset float64) {
	c.state.style.DashOffset = offset
}

// SetFillRule sets the fill rule.
func (c *Context) SetFillRule(r FillRule) {
	c.state.style.FillRule = r
}

// Style returns a copy of the current style.
func (c *Context) Style() Style {
	return c.state.style
}

// SetStyle replaces the current style.
func (c *Context) SetStyle(s Style) {
	c.state.style = s
}

// MoveTo starts a new subpath at (x, y).
func (c *Context) MoveTo(x, y float64) {
	c.path.MoveTo(c.state.m.Apply(x, y))
}

// LineTo adds a line segment from the current point to (x, y).
func (c *Context) LineTo(x, y float64) {
	c.path.LineTo(c.state.m.Apply(x, y))
}

// QuadTo adds a quadratic bezier to (x, y) with control point (cx, cy).
func (c *Context) QuadTo(cx, cy, x, y float64) {
	tcx, tcy := c.state.m.Apply(cx, cy)
	tx, ty := c.state.m.Apply(x, y)
	c.path.QuadTo(tcx, tcy, tx, ty)
}

// CubicTo adds a cubic bezier to (x, y) with control points (c1x, c1y) and
// (c2x, c2y).
func (c *Context) CubicTo(c1x, c1y, c2x, c2y, x, y float64) {
	t1x, t1y := c.state.m.Apply(c1x, c1y)
	t2x, t2y := c.state.m.Apply(c2x, c2y)
	tx, ty := c.state.m.Apply(x, y)
	c.path.CubicTo(t1x, t1y, t2x, t2y, tx, ty)
}

// ClosePath closes the current subpath.
func (c *Context) ClosePath() {
	c.path.Close()
}

// Close is an alias for ClosePath.
func (c *Context) Close() {
	c.ClosePath()
}

// ClearPath discards the current path without painting it.
func (c *Context) ClearPath() {
	c.path.Reset()
}

// Fill paints the current path's interior with the fill color and clears the
// path. Open subpaths are implicitly closed when filling.
func (c *Context) Fill() {
	c.FillPreserve()
	c.ClearPath()
}

// FillPreserve is Fill without clearing the path.
func (c *Context) FillPreserve() {
	s := c.state.style
	s.StrokeColor = nil
	c.render(s)
}

// Stroke paints the current path's outline with the stroke color and clears
// the path.
func (c *Context) Stroke() {
	c.StrokePreserve()
	c.ClearPath()
}

// StrokePreserve is Stroke without clearing the path.
func (c *Context) StrokePreserve() {
	s := c.state.style
	s.FillColor = nil
	c.render(s)
}

// FillStroke fills then strokes the current path and clears it.
func (c *Context) FillStroke() {
	c.render(c.state.style)
	c.ClearPath()
}

func (c *Context) render(s Style) {
	if c.path.Empty() {
		return
	}
	if !s.hasFill() && !s.hasStroke() {
		return
	}
	for _, r := range c.renderers {
		r.RenderPath(&c.path, s)
	}
}

// DrawCircle adds a circle to the current path. Call Fill, Stroke, or
// FillStroke to paint it.
func (c *Context) DrawCircle(cx, cy, r float64) {
	c.DrawPath(0, 0, CirclePath(cx, cy, r))
}

// DrawEllipse adds an axis-aligned ellipse to the current path.
func (c *Context) DrawEllipse(cx, cy, rx, ry float64) {
	c.DrawPath(0, 0, EllipsePath(cx, cy, rx, ry))
}

// DrawRectangle adds a rectangle with top-left corner (x, y) to the current
// path.
func (c *Context) DrawRectangle(x, y, w, h float64) {
	c.DrawPath(0, 0, RectPath(x, y, w, h))
}

// DrawLine adds a line segment from (x1, y1) to (x2, y2) to the current
// path.
func (c *Context) DrawLine(x1, y1, x2, y2 float64) {
	c.MoveTo(x1, y1)
	c.LineTo(x2, y2)
}

// DrawPath appends p, translated by (x, y), to the current path. Call Fill,
// Stroke, or FillStroke to paint it.
func (c *Context) DrawPath(x, y float64, p *Path) {
	c.path.appendPath(p, c.state.m.Mul(Translation(x, y)))
}

// Clear fills the entire surface of every renderer with the given color,
// ignoring the current transform and style.
func (c *Context) Clear(col color.Color) {
	if !visible(col) {
		return
	}
	s := Style{FillColor: col}
	for _, r := range c.renderers {
		w, h := r.Size()
		r.RenderPath(RectPath(0, 0, w, h), s)
	}
}

// DrawImage draws img with its top-left corner at (x, y), transformed by the
// current transform.
func (c *Context) DrawImage(img image.Image, x, y float64) {
	m := c.state.m.Mul(Translation(x, y))
	for _, r := range c.renderers {
		r.RenderImage(img, m)
	}
}

// SetFont sets the font and size in pixels used by DrawString.
func (c *Context) SetFont(f *Font, sizePx float64) {
	c.state.font = f
	c.state.fontSize = sizePx
}

// DrawString draws a single line of text with its baseline-left anchor at
// (x, y). The current transform applies to the anchor point only — glyphs
// render upright and unscaled. A font must be set with SetFont first.
func (c *Context) DrawString(s string, x, y float64) {
	if c.state.font == nil {
		return
	}
	tx, ty := c.state.m.Apply(x, y)
	for _, r := range c.renderers {
		r.RenderText(s, tx, ty, c.state.font, c.state.fontSize, c.state.style.FillColor)
	}
}

// MeasureString returns the advance width of s in pixels with the current
// font and size, or 0 if no font is set.
func (c *Context) MeasureString(s string) float64 {
	if c.state.font == nil {
		return 0
	}
	return c.state.font.Measure(s, c.state.fontSize)
}
