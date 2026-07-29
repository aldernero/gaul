package gaul

import (
	"errors"
	"math"

	"github.com/aldernero/gaul/render"
)

var (
	ErrInvalidSides = errors.New("invalid number of sides")
)

type RegularPolygon struct {
	Sides    int
	Radius   float64
	Rotation float64 // in radians
	Center   Point
}

func NewRegularPolygon(sides int, radius float64, center Point) (RegularPolygon, error) {
	if sides < 3 {
		return RegularPolygon{}, ErrInvalidSides
	}
	return RegularPolygon{
		Sides:  sides,
		Radius: radius,
		Center: center,
	}, nil
}

func NewRegularPolygonWithSideLength(sides int, sideLength float64, center Point) (RegularPolygon, error) {
	if sides < 3 {
		return RegularPolygon{}, ErrInvalidSides
	}
	radius := sideLength / (2 * math.Sin(math.Pi/float64(sides)))
	return RegularPolygon{
		Sides:  sides,
		Radius: radius,
		Center: center,
	}, nil
}

func (p RegularPolygon) Points() []Point {
	points := make([]Point, p.Sides)
	s := float64(p.Sides)
	for i := 0; i < p.Sides; i++ {
		sin, cos := math.Sincos(Tau*float64(i)/s + p.Rotation)
		points[i] = Point{X: p.Center.X + p.Radius*cos, Y: p.Center.Y + p.Radius*sin}
	}
	return points
}

func (p RegularPolygon) Edges() []Line {
	points := p.Points()
	edges := make([]Line, p.Sides)
	for i := 0; i < p.Sides; i++ {
		edges[i] = Line{P: points[i], Q: points[(i+1)%p.Sides]}
	}
	return edges
}

func (p RegularPolygon) Area() float64 {
	s := float64(p.Sides)
	return 0.5 * s * p.Radius * p.Radius * math.Sin(Tau/s)
}

func (p RegularPolygon) SideLength() float64 {
	return 2 * p.Radius * math.Sin(math.Pi/float64(p.Sides))
}

func (p RegularPolygon) Apothem() float64 {
	return p.Radius * math.Cos(math.Pi/float64(p.Sides))
}

func (p RegularPolygon) Perimeter() float64 {
	return p.SideLength() * float64(p.Sides)
}

func (p RegularPolygon) Lerp(t float64) Point {
	curve := p.ToCurve()
	return curve.Lerp(t)
}

func (p RegularPolygon) ToCurve() Curve {
	return Curve{Points: p.Points(), Closed: true}
}

func (p RegularPolygon) Draw(ctx *render.Context) {
	curve := p.ToCurve()
	curve.Draw(ctx)
}

type Polygon []Point

// signedArea is the shoelace area of the polygon, positive for one winding
// direction and negative for the other. Centroid needs the sign; Area does
// not.
func (p Polygon) signedArea() float64 {
	n := len(p)
	area := 0.0
	for i := 0; i < n; i++ {
		q := p[(i+1)%n]
		area += p[i].X*q.Y - q.X*p[i].Y
	}
	return 0.5 * area
}

// Area returns the area enclosed by the polygon, always non-negative
// regardless of winding direction.
func (p Polygon) Area() float64 {
	if len(p) < 3 {
		return 0
	}
	return math.Abs(p.signedArea())
}

// Centroid returns the area centroid of the polygon. A polygon with zero area
// (empty, degenerate, or self-cancelling) falls back to the mean of its
// vertices.
func (p Polygon) Centroid() Point {
	n := len(p)
	if n == 0 {
		return Point{}
	}
	// The shoelace centroid needs the signed area: using the absolute value
	// negates the result for one of the two winding directions.
	a := p.signedArea()
	if a == 0 {
		var totalX, totalY float64
		for _, v := range p {
			totalX += v.X
			totalY += v.Y
		}
		return Point{X: totalX / float64(n), Y: totalY / float64(n)}
	}
	var cx, cy float64
	for i := 0; i < n; i++ {
		q := p[(i+1)%n]
		cross := p[i].X*q.Y - q.X*p[i].Y
		cx += (p[i].X + q.X) * cross
		cy += (p[i].Y + q.Y) * cross
	}
	return Point{X: cx / (6 * a), Y: cy / (6 * a)}
}

// Perimeter returns the total edge length of the closed polygon.
func (p Polygon) Perimeter() float64 {
	n := len(p)
	if n < 2 {
		return 0
	}
	var perimeter float64
	for i := 0; i < n-1; i++ {
		perimeter += Distance(p[i], p[i+1])
	}
	perimeter += Distance(p[n-1], p[0])
	return perimeter
}
