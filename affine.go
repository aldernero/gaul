package gaul

import (
	"math"

	"github.com/aldernero/gaul/render"
)

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

// SetScale sets the scale factor for the x and y directions
func (a *Affine2D) SetScale(sx, sy float64) {
	a.a = sx
	a.e = sy
}

// SetTranslation sets the translation for the x and y directions
func (a *Affine2D) SetTranslation(tx, ty float64) {
	a.c = tx
	a.f = ty
}

// SetRotation sets the rotation for the transformation, replacing the linear
// (2x2) part of the matrix. Like SetScale and SetTranslation this assigns the
// component rather than composing with what is already there; use Mult with
// NewAffine2DWithRotation to compose a rotation onto an existing transform.
func (a *Affine2D) SetRotation(angle float64) {
	sin, cos := math.Sincos(angle)
	a.a = cos
	a.b = -sin
	a.d = sin
	a.e = cos
}

// SetRotationDegrees sets the rotation of the transformation
func (a *Affine2D) SetRotationDegrees(angle float64) {
	a.SetRotation(Deg2Rad(angle))
}

// SetShearFactor sets the shear factor in the x and y directions
func (a *Affine2D) SetShearFactor(sx, sy float64) {
	a.b = sx
	a.d = sy
}

// SetShearAngle sets the shear factor in the x and y directions
func (a *Affine2D) SetShearAngle(ax, ay float64) {
	a.SetShearFactor(math.Tan(ax), math.Tan(ay))
}

// SetShearAngleDegrees sets the shear factor in the x and y directions
func (a *Affine2D) SetShearAngleDegrees(ax, ay float64) {
	a.SetShearAngle(Deg2Rad(ax), Deg2Rad(ay))
}

// ToMatrix converts the transformation to a render.Matrix, dropping the
// projective row (g, h, i), which is unused by affine transforms. This lets
// an Affine2D be applied to a render.Context via Transform/SetTransform.
func (a *Affine2D) ToMatrix() render.Matrix {
	return render.Matrix{A: a.a, C: a.b, E: a.c, B: a.d, D: a.e, F: a.f}
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

// NewAffine2DWithRotationAbout returns an affine transformation that rotates
// by angle radians about the given point rather than the origin.
func NewAffine2DWithRotationAbout(angle float64, point Point) *Affine2D {
	return Mult(
		NewAffine2DWithTranslation(point.X, point.Y),
		Mult(NewAffine2DWithRotation(angle), NewAffine2DWithTranslation(-point.X, -point.Y)),
	)
}

// RotateVecAboutPoint rotates a vector around a point instead of the origin.
func RotateVecAboutPoint(vec Vec2, angle float64, point Point) Vec2 {
	return NewAffine2DWithRotationAbout(angle, point).MultVec(vec)
}

// RotatePointAboutPoint rotates a point around a point instead of the origin.
func RotatePointAboutPoint(point Point, angle float64, rotationPoint Point) Point {
	return RotateVecAboutPoint(point.ToVec2(), angle, rotationPoint).ToPoint()
}

// RotateLineAboutPoint rotates a line around a point instead of the origin.
func RotateLineAboutPoint(line Line, angle float64, rotationPoint Point) Line {
	m := NewAffine2DWithRotationAbout(angle, rotationPoint)
	return m.TransformLine(line)
}

// RotateCurveAboutPoint rotates a curve around a point instead of the origin.
func RotateCurveAboutPoint(curve Curve, angle float64, rotationPoint Point) Curve {
	return NewAffine2DWithRotationAbout(angle, rotationPoint).TransformCurve(curve)
}

// RotateAboutPoint rotates a vector around a point instead of the origin.
//
// Deprecated: the receiver is unused. Use the package-level
// [RotateVecAboutPoint] instead.
func (a *Affine2D) RotateAboutPoint(vec Vec2, angle float64, point Point) Vec2 {
	return RotateVecAboutPoint(vec, angle, point)
}

// RotatePointAboutPoint rotates a point around a point instead of the origin.
//
// Deprecated: the receiver is unused. Use the package-level
// [RotatePointAboutPoint] instead.
func (a *Affine2D) RotatePointAboutPoint(point Point, angle float64, rotationPoint Point) Point {
	return RotatePointAboutPoint(point, angle, rotationPoint)
}

// RotateLineAboutPoint rotates a line around a point instead of the origin.
//
// Deprecated: the receiver is unused. Use the package-level
// [RotateLineAboutPoint] instead.
func (a *Affine2D) RotateLineAboutPoint(line Line, angle float64, rotationPoint Point) Line {
	return RotateLineAboutPoint(line, angle, rotationPoint)
}

// RotateCurveAboutPoint rotates a curve around a point instead of the origin.
//
// Deprecated: the receiver is unused. Use the package-level
// [RotateCurveAboutPoint] instead.
func (a *Affine2D) RotateCurveAboutPoint(curve Curve, angle float64, rotationPoint Point) Curve {
	return RotateCurveAboutPoint(curve, angle, rotationPoint)
}
