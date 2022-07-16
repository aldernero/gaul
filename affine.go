package gaul

// Affine transformation in 2D using homogeneous coordinates
// The matrix is of the form
// | a b c |
// | d e f | where g=h=0 and i=1
// | g h i |
type Affine2D struct {
	a, b, c float64
	d, e, f float64
	g, h, i float64
}

// Returns an identity affine transformation (no rotation, scaling, shearing, or translation)
func NewAffine2D() *Affine2D {
	var affine Affine2D
	affine.a = 1
	affine.e = 1
	affine.i = 1
	return &affine
}

// Mult calculates the new Affine2D corresponding to multiplying p and q
func Mult(p, q Affine2D) *Affine2D {
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
func Add(p, q Affine2D) *Affine2D {
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
func (a Affine2D) TransformCurve(c Curve) Curve {
	var curve Curve
	curve.Closed = c.Closed
	for _, p := range c.Points {
		curve.Points = append(curve.Points, a.TransformPoint(p))
	}
	return curve
}
