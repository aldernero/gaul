package gaul

import "math"

type Vec2 struct {
	X float64
	Y float64
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
