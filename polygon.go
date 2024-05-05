package gaul

import (
	"errors"
	"github.com/tdewolff/canvas"
	"math"
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
	points := make([]Point, 0)
	s := float64(p.Sides)
	for i := 0; i < p.Sides; i++ {
		angle := Tau*float64(i)/s + p.Rotation
		x := p.Center.X + p.Radius*math.Cos(angle)
		y := p.Center.Y + p.Radius*math.Sin(angle)
		points = append(points, Point{X: x, Y: y})
	}
	return points
}

func (p RegularPolygon) Edges() []Line {
	points := p.Points()
	edges := make([]Line, 0)
	for i := 0; i < p.Sides; i++ {
		edges = append(edges, Line{P: points[i], Q: points[(i+1)%p.Sides]})
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
	points := p.Points()
	curve := Curve{}
	curve.Closed = true
	for _, point := range points {
		curve.AddPoint(point.X, point.Y)
	}
	return curve
}

func (p RegularPolygon) Draw(ctx *canvas.Context) {
	curve := p.ToCurve()
	curve.Draw(ctx)
}

type Polygon []Point

func (p Polygon) Area() float64 {
	// shoelace formula
	n := len(p)
	area := 0.0
	for i := 0; i < n; i++ {
		area += p[i].X*p[(i+1)%n].Y - p[(i+1)%n].X*p[i].Y
	}
	return 0.5 * math.Abs(area)
}

func (p Polygon) Centroid() Point {
	n := len(p)
	var cx, cy float64
	A := p.Area()
	for i := 0; i < n; i++ {
		cx += (p[i].X + p[(i+1)%n].X) * (p[i].X*p[(i+1)%n].Y - p[(i+1)%n].X*p[i].Y)
		cy += (p[i].Y + p[(i+1)%n].Y) * (p[i].X*p[(i+1)%n].Y - p[(i+1)%n].X*p[i].Y)
	}
	cx /= 6 * A
	cy /= 6 * A
	return Point{X: cx, Y: cy}
}

func (p Polygon) Perimeter() float64 {
	n := len(p)
	var perimeter float64
	for i := 0; i < n-1; i++ {
		perimeter += Distance(p[i], p[i+1])
	}
	perimeter += Distance(p[n-1], p[0])
	return perimeter
}
