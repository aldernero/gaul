package gaul

import (
	"math"

	"github.com/aldernero/interp"
)

const (
	defaultBezierResolution = 100
	defaultSplineResolution = 100
)

// Chaikin senerates a Chaikin curve given a set of control points,
// a cutoff ratio, and the number of steps to use in the
// calculation.
func Chaikin(c Curve, q float64, n int) Curve {
	points := make([]Point, 0)
	// Start with control points
	points = append(points, c.Points...)
	left := q / 2
	right := 1 - (q / 2)
	for i := 0; i < n; i++ {
		newPoints := make([]Point, 0)
		for j := 0; j < len(points)-1; j++ {
			p1 := points[j]
			p2 := points[j+1]
			q := Point{
				X: right*p1.X + left*p2.X,
				Y: right*p1.Y + left*p2.Y,
			}
			r := Point{
				X: left*p1.X + right*p2.X,
				Y: left*p1.Y + right*p2.Y,
			}
			newPoints = append(newPoints, q, r)
		}
		if c.Closed {
			p1 := points[len(points)-1]
			p2 := points[0]
			q := Point{
				X: right*p1.X + left*p2.X,
				Y: right*p1.Y + left*p2.Y,
			}
			r := Point{
				X: left*p1.X + right*p2.X,
				Y: left*p1.Y + right*p2.Y,
			}
			newPoints = append(newPoints, q, r)
		}
		points = []Point{}
		points = append(points, newPoints...)
	}
	return Curve{Points: points, Closed: c.Closed}
}

// Lissajous represents the core parameters for a 2D Lissajous curve
type Lissajous struct {
	Nx int
	Ny int
	Px float64
	Py float64
}

// GenLissajous generates a Lissajous curve given parameters, a number of points
// to use (i.e. resolution), and an offset and scale (typically to convert
// to screen coordinates)
func GenLissajous(l Lissajous, n int, offset Point, s float64) Curve {
	curve := Curve{}
	maxPhase := Tau / float64(Gcd(l.Nx, l.Ny))
	dt := maxPhase / float64(n)
	for t := 0.0; t < maxPhase; t += dt {
		xPos := s*math.Sin(float64(l.Nx)*t+l.Px) + offset.X
		yPos := s*math.Sin(float64(l.Ny)*t+l.Py) + offset.Y
		point := Point{X: xPos, Y: yPos}
		curve.Points = append(curve.Points, point)
	}
	return curve
}

// PaduaPoints calculates Padua points for a certain class of Lissajous curves,
// where Nx = Ny +/- 1. The correspond to intersection points and
// some of the outside points on the curve
// See https://en.wikipedia.org/wiki/Padua_points for more details.
func PaduaPoints(n int) []Point {
	points := make([]Point, 0)
	for i := 0; i <= n; i++ {
		delta := 0
		if n%2 == 1 && i%2 == 1 {
			delta = 1
		}
		for j := 1; j < (n/2)+2+delta; j++ {
			x := math.Cos(float64(i) * Pi / float64(n))
			var y float64
			if i%2 == 1 {
				y = math.Cos(float64(2*j-2) * Pi / float64(n+1))
			} else {
				y = math.Cos(float64(2*j-1) * Pi / float64(n+1))
			}
			points = append(points, Point{X: x, Y: y})
		}
	}
	return points
}

// PulsarPlot transforms a slice of curves into a slice of curves representing the segments that make up a pulsar plot
func PulsarPlot(curves []Curve) []Curve {
	result := make([]Curve, 0)
	if len(curves) == 0 {
		return result
	}
	n := len(curves[0].Points)
	mins := make([]float64, n)
	for _, c := range curves {
		segment := Curve{}
		for i := 0; i < n; i++ {
			px := c.Points[i].X
			py := c.Points[i].Y
			if py >= mins[i] {
				mins[i] = py
				segment.AddPoint(px, py)
			} else {
				if len(segment.Points) > 1 {
					result = append(result, segment)
				}
				segment = Curve{}
			}
		}
		if len(segment.Points) > 1 {
			result = append(result, segment)
		}
	}
	return result
}

// Bezier curves

type quadBezier struct {
	p0, p1, p2 Point
	resolution int
}

func (b quadBezier) Curve() Curve {
	var result Curve
	ts := Linspace(0, 1, b.resolution, true)
	l1 := Line{P: b.p0, Q: b.p1}
	l2 := Line{P: b.p1, Q: b.p2}
	for _, t := range ts {
		q0 := l1.Lerp(t)
		q1 := l2.Lerp(t)
		q := Line{P: q0, Q: q1}.Lerp(t)
		result.AddPoint(q.X, q.Y)
	}
	return result
}

// QuadBezier generates a quadratic Bézier curve between two points given a control point
func QuadBezier(p, q, c Point) Curve {
	qb := quadBezier{
		p0:         p,
		p1:         c,
		p2:         q,
		resolution: defaultBezierResolution,
	}
	return qb.Curve()
}

// QuadBezierWithResolution generates a quadratic Bézier curve with a specific resolution
func QuadBezierWithResolution(p, q, c Point, resolution int) Curve {
	qb := quadBezier{
		p0:         p,
		p1:         c,
		p2:         q,
		resolution: resolution,
	}
	return qb.Curve()
}

type cubicBezier struct {
	p0, p1, p2, p3 Point
	resolution     int
}

func (b cubicBezier) Curve() Curve {
	var result Curve
	ts := Linspace(0, 1, b.resolution, true)
	l1 := Line{P: b.p0, Q: b.p1}
	l2 := Line{P: b.p1, Q: b.p2}
	l3 := Line{P: b.p2, Q: b.p3}
	for _, t := range ts {
		q0 := l1.Lerp(t)
		q1 := l2.Lerp(t)
		q2 := l3.Lerp(t)
		q01 := Line{P: q0, Q: q1}.Lerp(t)
		q12 := Line{P: q1, Q: q2}.Lerp(t)
		q := Line{P: q01, Q: q12}.Lerp(t)
		result.AddPoint(q.X, q.Y)
	}
	return result
}

// CubicBezier generates a cubic Bézier curve between two points given two control points
func CubicBezier(p, q, c1, c2 Point) Curve {
	cb := cubicBezier{
		p0:         p,
		p1:         c1,
		p2:         c2,
		p3:         q,
		resolution: defaultBezierResolution,
	}
	return cb.Curve()
}

// CubicBezierWithResolution generates a cubic Bézier curve with a specific resolution
func CubicBezierWithResolution(p, q, c1, c2 Point, resolution int) Curve {
	cb := cubicBezier{
		p0:         p,
		p1:         c1,
		p2:         c2,
		p3:         q,
		resolution: resolution,
	}
	return cb.Curve()
}

// Quartic bezier curve

type quarticBezier struct {
	p0, p1, p2, p3, p4 Point
	resolution         int
}

func (b quarticBezier) Curve() Curve {
	var result Curve
	ts := Linspace(0, 1, b.resolution, true)
	l1 := Line{P: b.p0, Q: b.p1}
	l2 := Line{P: b.p1, Q: b.p2}
	l3 := Line{P: b.p2, Q: b.p3}
	l4 := Line{P: b.p3, Q: b.p4}
	for _, t := range ts {
		q0 := l1.Lerp(t)
		q1 := l2.Lerp(t)
		q2 := l3.Lerp(t)
		q3 := l4.Lerp(t)
		q01 := Line{P: q0, Q: q1}.Lerp(t)
		q12 := Line{P: q1, Q: q2}.Lerp(t)
		q23 := Line{P: q2, Q: q3}.Lerp(t)
		q012 := Line{P: q01, Q: q12}.Lerp(t)
		q123 := Line{P: q12, Q: q23}.Lerp(t)
		q := Line{P: q012, Q: q123}.Lerp(t)
		result.AddPoint(q.X, q.Y)
	}
	return result
}

// QuarticBezier generates a quartic Bézier curve between two points given three control points
func QuarticBezier(p, q, c1, c2, c3 Point) Curve {
	qb := quarticBezier{
		p0:         p,
		p1:         c1,
		p2:         c2,
		p3:         c3,
		p4:         q,
		resolution: defaultBezierResolution,
	}
	return qb.Curve()
}

// QuarticBezierWithResolution generates a quartic Bézier curve with a specific resolution
func QuarticBezierWithResolution(p, q, c1, c2, c3 Point, resolution int) Curve {
	qb := quarticBezier{
		p0:         p,
		p1:         c1,
		p2:         c2,
		p3:         c3,
		p4:         q,
		resolution: resolution,
	}
	return qb.Curve()
}

func QuarticBezierWithTriangleControl(p, q Point, tri Triangle) Curve {
	qb := quarticBezier{
		p0:         p,
		p1:         tri.A,
		p2:         tri.B,
		p3:         tri.C,
		p4:         q,
		resolution: defaultBezierResolution,
	}
	return qb.Curve()
}

// Splines

// CubicSpline represents a cubic spline
type CubicSpline struct {
	spline interp.CubicSpline
	res    int
}

// NewCubicSpline creates a cubic spline from a set of points
func NewCubicSpline(points []Point) CubicSpline {
	n := len(points)
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range points {
		xs[i] = p.X
		ys[i] = p.Y
	}
	cs, _ := interp.NewCubicSpline(xs, ys)
	return CubicSpline{spline: cs, res: defaultSplineResolution}
}

// NewCubicSplineWithResolution creates a cubic spline from a set of points
// with a specific resolution
func NewCubicSplineWithResolution(points []Point, res int) CubicSpline {
	n := len(points)
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range points {
		xs[i] = p.X
		ys[i] = p.Y
	}
	cs, _ := interp.NewCubicSpline(xs, ys)
	return CubicSpline{spline: cs, res: res}
}

func NewRegularCubicSpline(points []Point) CubicSpline {
	n := len(points)
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range points {
		xs[i] = p.X
		ys[i] = p.Y
	}
	cs, _ := interp.NewRegularCubicSpline(xs, ys)
	return CubicSpline{spline: cs, res: defaultSplineResolution}
}

// At evaluates the spline at a given x value
func (cs CubicSpline) At(x float64) float64 {
	return cs.spline.Eval(x)
}

// ToCurve converts the spline to a curve
func (cs CubicSpline) ToCurve() Curve {
	var result Curve
	min, max := cs.spline.GetMinMaxX()
	xs := Linspace(min, max, cs.res, true)
	for _, x := range xs {
		y := cs.spline.Eval(x)
		result.AddPoint(x, y)
	}
	return result
}
