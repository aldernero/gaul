package gaul

import "math"

// Affine2D is a type for dealing with affine transformations in 2D using homogeneous coordinates
type Affine2D struct {
	a, b, c float64
	d, e, f float64
	g, h, i float64
}

// NewAffine2D returns an identity affine transformation (no rotation, scaling, shearing, or translation)
func NewAffine2D() *Affine2D {
	var affine Affine2D
	affine.a = 1
	affine.e = 1
	affine.i = 1
	return &affine
}

// NewAffine2DWithScale returns an affine transformation with scaling configured
func NewAffine2DWithScale(sx, sy float64) *Affine2D {
	var affine Affine2D
	affine.a = sx
	affine.e = sy
	affine.i = 1
	return &affine
}

// NewAffine2DWithRotation returns an affine transformation with rotation configured
func NewAffine2DWithRotation(angle float64) *Affine2D {
	var affine Affine2D
	affine.a = math.Cos(angle)
	affine.b = -math.Sin(angle)
	affine.d = math.Sin(angle)
	affine.e = math.Cos(angle)
	affine.i = 1
	return &affine
}

// NewAffine2DWithTranslation returns an affine transformation with translation configured
func NewAffine2DWithTranslation(tx, ty float64) *Affine2D {
	var affine Affine2D
	affine.a = 1
	affine.c = tx
	affine.e = 1
	affine.f = ty
	affine.i = 1
	return &affine
}

// NewAffine2DWithShear returns an affine transformation with shearing configured
func NewAffine2DWithShear(sx, sy float64) *Affine2D {
	var affine Affine2D
	affine.a = 1
	affine.b = sx
	affine.d = sy
	affine.e = 1
	affine.i = 1
	return &affine
}

// Mult calculates the new Affine2D corresponding to multiplying p and q
func Mult(p, q *Affine2D) *Affine2D {
	return &Affine2D{
		a: p.a*q.a + p.b*q.d + p.c*q.g,
		b: p.a*q.b + p.b*q.e + p.c*q.h,
		c: p.a*q.c + p.b*q.f + p.c*q.i,
		d: p.d*q.a + p.e*q.d + p.f*q.g,
		e: p.d*q.b + p.e*q.e + p.f*q.h,
		f: p.d*q.c + p.e*q.f + p.f*q.i,
		g: p.g*q.a + p.h*q.d + p.i*q.g,
		h: p.g*q.b + p.h*q.e + p.i*q.h,
		i: p.g*q.c + p.h*q.f + p.i*q.i,
	}
}

// Add calculates the new Affine2D corresponding to adding p and q
func Add(p, q *Affine2D) *Affine2D {
	return &Affine2D{
		a: p.a + q.a,
		b: p.b + q.b,
		c: p.c + q.c,
		d: p.d + q.d,
		e: p.e + q.e,
		f: p.f + q.f,
		g: p.g + q.g,
		h: p.h + q.h,
		i: p.i + q.i,
	}
}

// MultVec returns a new vector which is the result of applying the affine transformation
func (a *Affine2D) MultVec(v Vec2) Vec2 {
	return Vec2{
		X: a.a*v.X + a.b*v.Y + a.c,
		Y: a.d*v.X + a.e*v.Y + a.f,
	}
}

// Transform applies the affine transformation to a vector
func (a *Affine2D) Transform(v Vec2) Vec2 {
	return a.MultVec(v)
}

// TransformPoint applies the affine transformation to a point
func (a *Affine2D) TransformPoint(p Point) Point {
	vec := a.MultVec(p.ToVec2())
	return vec.ToPoint()
}

// TransformLine applies the affine transformation to a line
func (a *Affine2D) TransformLine(l Line) Line {
	return Line{
		P: a.TransformPoint(l.P),
		Q: a.TransformPoint(l.Q),
	}
}

// TransformCurve applies the affine transformation to a curve
func (a *Affine2D) TransformCurve(c Curve) Curve {
	var curve Curve
	curve.Closed = c.Closed
	for _, p := range c.Points {
		curve.Points = append(curve.Points, a.TransformPoint(p))
	}
	return curve
}

// RotateAboutPoint rotates a vector around a point instead of the origin
func (a *Affine2D) RotateAboutPoint(vec Vec2, angle float64, point Point) Vec2 {
	moveOrigin := NewAffine2DWithTranslation(-point.X, -point.Y)
	rotate := NewAffine2DWithRotation(angle)
	returnOrigin := NewAffine2DWithTranslation(point.X, point.Y)
	temp1 := moveOrigin.Transform(vec)
	temp2 := rotate.Transform(temp1)
	result := returnOrigin.Transform(temp2)
	return result
}

// RotatePointAboutPoint rotates a point around a point instead of the origin
func (a *Affine2D) RotatePointAboutPoint(point Point, angle float64, rotationPoint Point) Point {
	vec := a.RotateAboutPoint(point.ToVec2(), angle, rotationPoint)
	return vec.ToPoint()
}

// RotateLineAboutPoint rotates a line around a point instead of the origin
func (a *Affine2D) RotateLineAboutPoint(line Line, angle float64, rotationPoint Point) Line {
	return Line{
		P: a.RotatePointAboutPoint(line.P, angle, rotationPoint),
		Q: a.RotatePointAboutPoint(line.Q, angle, rotationPoint),
	}
}

// RotateCurveAboutPoint rotates a curve around a point instead of the origin
func (a *Affine2D) RotateCurveAboutPoint(curve Curve, angle float64, rotationPoint Point) Curve {
	var result Curve
	result.Closed = curve.Closed
	for _, p := range curve.Points {
		result.Points = append(result.Points, a.RotatePointAboutPoint(p, angle, rotationPoint))
	}
	return result
}
