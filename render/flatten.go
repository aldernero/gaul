package render

import "math"

// flattenSegLen is the target step length in pixels along a curve's control
// polygon when flattening beziers for rasterization. Smaller values increase
// fidelity and cost; at 2 px the deviation from the true curve is far below
// a pixel for any curvature that survives anti-aliasing.
const flattenSegLen = 2.0

const (
	flattenMinSteps = 4
	flattenMaxSteps = 256
)

// Flatten converts the path into polylines, one per subpath, approximating
// beziers with line segments. Each polyline is a packed x,y slice. closed[i]
// reports whether polyline i was terminated by Close; closed polylines do
// not repeat their first point. Only the raster backend flattens — the SVG
// backend serializes curves natively.
func (p *Path) Flatten() (polys [][]float64, closed []bool) {
	var poly []float64
	var curX, curY float64
	var startX, startY float64
	flush := func(cl bool) {
		if len(poly) >= 4 {
			polys = append(polys, poly)
			closed = append(closed, cl)
		}
		poly = nil
	}
	p.Iterate(func(v PathVerb, pts []float64) {
		switch v {
		case VerbMoveTo:
			flush(false)
			curX, curY = pts[0], pts[1]
			startX, startY = curX, curY
			poly = append(poly, curX, curY)
		case VerbLineTo:
			if poly == nil {
				poly = append(poly, curX, curY)
			}
			curX, curY = pts[0], pts[1]
			poly = append(poly, curX, curY)
		case VerbQuadTo:
			if poly == nil {
				poly = append(poly, curX, curY)
			}
			cx, cy, x, y := pts[0], pts[1], pts[2], pts[3]
			n := flattenSteps(dist(curX, curY, cx, cy) + dist(cx, cy, x, y))
			for i := 1; i <= n; i++ {
				t := float64(i) / float64(n)
				px, py := quadPoint(curX, curY, cx, cy, x, y, t)
				poly = append(poly, px, py)
			}
			curX, curY = x, y
		case VerbCubicTo:
			if poly == nil {
				poly = append(poly, curX, curY)
			}
			c1x, c1y, c2x, c2y, x, y := pts[0], pts[1], pts[2], pts[3], pts[4], pts[5]
			n := flattenSteps(dist(curX, curY, c1x, c1y) + dist(c1x, c1y, c2x, c2y) + dist(c2x, c2y, x, y))
			for i := 1; i <= n; i++ {
				t := float64(i) / float64(n)
				px, py := cubicPoint(curX, curY, c1x, c1y, c2x, c2y, x, y, t)
				poly = append(poly, px, py)
			}
			curX, curY = x, y
		case VerbClose:
			flush(true)
			// Per SVG semantics, drawing after a close continues from the
			// subpath start point.
			curX, curY = startX, startY
		}
	})
	flush(false)
	return polys, closed
}

func flattenSteps(controlLen float64) int {
	n := int(math.Ceil(controlLen / flattenSegLen))
	return minmax(n, flattenMinSteps, flattenMaxSteps)
}

func minmax(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}

func quadPoint(x0, y0, cx, cy, x1, y1, t float64) (float64, float64) {
	u := 1 - t
	a := u * u
	b := 2 * u * t
	c := t * t
	return a*x0 + b*cx + c*x1, a*y0 + b*cy + c*y1
}

func cubicPoint(x0, y0, c1x, c1y, c2x, c2y, x1, y1, t float64) (float64, float64) {
	u := 1 - t
	a := u * u * u
	b := 3 * u * u * t
	c := 3 * u * t * t
	d := t * t * t
	return a*x0 + b*c1x + c*c2x + d*x1, a*y0 + b*c1y + c*c2y + d*y1
}
