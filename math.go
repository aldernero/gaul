package gaul

import "math"

type Vec2 struct {
	X float64
	Y float64
}

// Vec2FromPoint returns a vector from the origin to p
func Vec2FromPoint(p Point) Vec2 {
	return Vec2{X: p.X, Y: p.Y}
}

// Vec2FromPoints returns a vector from p to q
func Vec2FromPoints(p, q Point) Vec2 {
	return Vec2{X: q.X - p.X, Y: q.Y - p.Y}
}

func Vec2FromLine(l Line) Vec2 {
	return Vec2FromPoints(l.P, l.Q)
}

func (v Vec2) Add(u Vec2) Vec2 {
	return Vec2{X: v.X + u.X, Y: v.Y + u.Y}
}

func (v Vec2) Sub(u Vec2) Vec2 {
	return Vec2{X: v.X - u.X, Y: v.Y - u.Y}
}

func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{X: s * v.X, Y: s * v.Y}
}

func (v Vec2) Dot(u Vec2) float64 {
	return v.X*u.X + v.Y*u.Y
}

func (v Vec2) Mag() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2))
}

func (v Vec2) Normalize() Vec2 {
	m := v.Mag()
	if m == 0 {
		panic("cannot normalize vector with zero magnitude")
	}
	return v.Scale(1 / m)
}

func (v Vec2) UnitNormal() Vec2 {
	m := v.Mag()
	return Vec2{X: v.Y / m, Y: -v.X / m}
}

func (v Vec2) ToPoint() Point {
	return Point{X: v.X, Y: v.Y}
}

func (v Vec2) ToString(prec int) string {
	return "(" + FloatString(v.X, prec) + ", " + FloatString(v.Y, prec) + ")"
}

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{X: v.X + u.X, Y: v.Y + u.Y, Z: v.Z + u.Z}
}

func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{X: v.X - u.X, Y: v.Y - u.Y, Z: v.Z - u.Z}
}

func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{X: s * v.X, Y: s * v.Y, Z: s * v.Z}
}

func (v Vec3) Dot(u Vec3) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

func (v Vec3) Mag() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2) + math.Pow(v.Z, 2))
}

func (v Vec3) Normalize() Vec3 {
	m := v.Mag()
	if m == 0 {
		panic("cannot normalize vector with zero magnitude")
	}
	return v.Scale(1 / m)
}

func (v Vec3) ToString(prec int) string {
	return "(" + FloatString(v.X, prec) + ", " + FloatString(v.Y, prec) + ", " + FloatString(v.Z, prec) + ")"
}
