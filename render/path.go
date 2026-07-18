package render

// PathVerb identifies one segment type in a Path.
type PathVerb uint8

const (
	// VerbMoveTo starts a new subpath; 1 point.
	VerbMoveTo PathVerb = iota
	// VerbLineTo adds a line segment; 1 point.
	VerbLineTo
	// VerbQuadTo adds a quadratic bezier; 2 points (control, end).
	VerbQuadTo
	// VerbCubicTo adds a cubic bezier; 3 points (control1, control2, end).
	VerbCubicTo
	// VerbClose closes the current subpath; 0 points.
	VerbClose
)

var verbPoints = [...]int{VerbMoveTo: 1, VerbLineTo: 1, VerbQuadTo: 2, VerbCubicTo: 3, VerbClose: 0}

// Path is a retained sequence of path segments in absolute coordinates.
// Paths are built with MoveTo/LineTo/QuadTo/CubicTo/Close and consumed by
// Renderers; the SVG backend serializes segments natively while the raster
// backend flattens curves to polylines.
type Path struct {
	verbs  []PathVerb
	coords []float64 // packed x,y pairs, verbPoints[verb] pairs per verb

	startX, startY float64 // start of current subpath
	curX, curY     float64 // current point
	hasCurrent     bool
}

// MoveTo starts a new subpath at (x, y).
func (p *Path) MoveTo(x, y float64) {
	p.verbs = append(p.verbs, VerbMoveTo)
	p.coords = append(p.coords, x, y)
	p.startX, p.startY = x, y
	p.curX, p.curY = x, y
	p.hasCurrent = true
}

// LineTo adds a line segment from the current point to (x, y). If there is
// no current point it behaves like MoveTo.
func (p *Path) LineTo(x, y float64) {
	if !p.hasCurrent {
		p.MoveTo(x, y)
		return
	}
	p.verbs = append(p.verbs, VerbLineTo)
	p.coords = append(p.coords, x, y)
	p.curX, p.curY = x, y
}

// QuadTo adds a quadratic bezier from the current point to (x, y) with
// control point (cx, cy).
func (p *Path) QuadTo(cx, cy, x, y float64) {
	if !p.hasCurrent {
		p.MoveTo(cx, cy)
	}
	p.verbs = append(p.verbs, VerbQuadTo)
	p.coords = append(p.coords, cx, cy, x, y)
	p.curX, p.curY = x, y
}

// CubicTo adds a cubic bezier from the current point to (x, y) with control
// points (c1x, c1y) and (c2x, c2y).
func (p *Path) CubicTo(c1x, c1y, c2x, c2y, x, y float64) {
	if !p.hasCurrent {
		p.MoveTo(c1x, c1y)
	}
	p.verbs = append(p.verbs, VerbCubicTo)
	p.coords = append(p.coords, c1x, c1y, c2x, c2y, x, y)
	p.curX, p.curY = x, y
}

// Close closes the current subpath with a line back to its starting point.
func (p *Path) Close() {
	if !p.hasCurrent {
		return
	}
	p.verbs = append(p.verbs, VerbClose)
	p.curX, p.curY = p.startX, p.startY
}

// Empty reports whether the path contains no segments.
func (p *Path) Empty() bool {
	return len(p.verbs) == 0
}

// Reset clears the path while keeping the backing arrays for reuse.
func (p *Path) Reset() {
	p.verbs = p.verbs[:0]
	p.coords = p.coords[:0]
	p.hasCurrent = false
}

// Copy returns a deep copy of the path.
func (p *Path) Copy() *Path {
	q := &Path{
		verbs:  append([]PathVerb(nil), p.verbs...),
		coords: append([]float64(nil), p.coords...),
		startX: p.startX, startY: p.startY,
		curX: p.curX, curY: p.curY,
		hasCurrent: p.hasCurrent,
	}
	return q
}

// Transform returns a copy of the path with every coordinate mapped through
// m. Affine maps of bezier control points are exact, so no accuracy is lost.
func (p *Path) Transform(m Matrix) *Path {
	q := p.Copy()
	for i := 0; i < len(q.coords); i += 2 {
		q.coords[i], q.coords[i+1] = m.Apply(q.coords[i], q.coords[i+1])
	}
	q.startX, q.startY = m.Apply(p.startX, p.startY)
	q.curX, q.curY = m.Apply(p.curX, p.curY)
	return q
}

// Bounds returns the bounding box of the path's points, including bezier
// control points (a fast, conservative bound). It returns zeros for an empty
// path.
func (p *Path) Bounds() (minX, minY, maxX, maxY float64) {
	if len(p.coords) == 0 {
		return 0, 0, 0, 0
	}
	minX, minY = p.coords[0], p.coords[1]
	maxX, maxY = minX, minY
	for i := 2; i < len(p.coords); i += 2 {
		x, y := p.coords[i], p.coords[i+1]
		minX = min(minX, x)
		maxX = max(maxX, x)
		minY = min(minY, y)
		maxY = max(maxY, y)
	}
	return minX, minY, maxX, maxY
}

// Iterate calls fn for each verb with its points. The pts slice is a view
// into the path's storage and must not be retained or modified.
func (p *Path) Iterate(fn func(v PathVerb, pts []float64)) {
	i := 0
	for _, v := range p.verbs {
		n := 2 * verbPoints[v]
		fn(v, p.coords[i:i+n])
		i += n
	}
}

// appendPath appends q's segments to p, transforming them through m.
func (p *Path) appendPath(q *Path, m Matrix) {
	q.Iterate(func(v PathVerb, pts []float64) {
		p.verbs = append(p.verbs, v)
		for i := 0; i < len(pts); i += 2 {
			x, y := m.Apply(pts[i], pts[i+1])
			p.coords = append(p.coords, x, y)
			switch v {
			case VerbMoveTo:
				p.startX, p.startY = x, y
			}
			p.curX, p.curY = x, y
		}
		if v == VerbClose {
			p.curX, p.curY = p.startX, p.startY
		}
		p.hasCurrent = true
	})
}

// kappa is the control-point offset ratio that makes four cubic beziers
// approximate a circle to within about 0.02% of the radius.
const kappa = 0.5522847498307933

// CirclePath returns a circle of radius r centered at (cx, cy), built from
// four cubic beziers so that both raster and SVG backends receive smooth
// curves.
func CirclePath(cx, cy, r float64) *Path {
	return EllipsePath(cx, cy, r, r)
}

// EllipsePath returns an axis-aligned ellipse with radii (rx, ry) centered
// at (cx, cy), built from four cubic beziers.
func EllipsePath(cx, cy, rx, ry float64) *Path {
	ox, oy := rx*kappa, ry*kappa
	p := &Path{}
	p.MoveTo(cx+rx, cy)
	p.CubicTo(cx+rx, cy+oy, cx+ox, cy+ry, cx, cy+ry)
	p.CubicTo(cx-ox, cy+ry, cx-rx, cy+oy, cx-rx, cy)
	p.CubicTo(cx-rx, cy-oy, cx-ox, cy-ry, cx, cy-ry)
	p.CubicTo(cx+ox, cy-ry, cx+rx, cy-oy, cx+rx, cy)
	p.Close()
	return p
}

// RectPath returns a rectangle with top-left corner (x, y), width w, and
// height h.
func RectPath(x, y, w, h float64) *Path {
	p := &Path{}
	p.MoveTo(x, y)
	p.LineTo(x+w, y)
	p.LineTo(x+w, y+h)
	p.LineTo(x, y+h)
	p.Close()
	return p
}
