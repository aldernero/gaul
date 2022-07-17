package gaul

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

// TestAffine2D_Mult tests basic matrix multiplication
func TestAffine2D_Mult(t *testing.T) {
	assert := assert.New(t)
	matrix1 := Affine2D{
		a: 2,
		b: 3,
		c: 5,
		d: 7,
		e: 11,
		f: 13,
		g: 17,
		h: 19,
		i: 23,
	}
	matrix2 := Affine2D{
		a: 29,
		b: 31,
		c: 37,
		d: 41,
		e: 43,
		f: 47,
		g: 53,
		h: 59,
		i: 61,
	}
	result := Mult(matrix1, matrix2)
	assert.Equal(446.0, result.a)
	assert.Equal(486.0, result.b)
	assert.Equal(520.0, result.c)
	assert.Equal(1343.0, result.d)
	assert.Equal(1457.0, result.e)
	assert.Equal(1569.0, result.f)
	assert.Equal(2491.0, result.g)
	assert.Equal(2701.0, result.h)
	assert.Equal(2925.0, result.i)
}

// TestAffine2D_Add tests basic matrix addition
func TestAffine2D_Add(t *testing.T) {
	assert := assert.New(t)
	matrix1 := Affine2D{
		a: 2,
		b: 3,
		c: 5,
		d: 7,
		e: 11,
		f: 13,
		g: 17,
		h: 19,
		i: 23,
	}
	matrix2 := Affine2D{
		a: 29,
		b: 31,
		c: 37,
		d: 41,
		e: 43,
		f: 47,
		g: 53,
		h: 59,
		i: 61,
	}
	result := Add(matrix1, matrix2)
	assert.Equal(31.0, result.a)
	assert.Equal(34.0, result.b)
	assert.Equal(42.0, result.c)
	assert.Equal(48.0, result.d)
	assert.Equal(54.0, result.e)
	assert.Equal(60.0, result.f)
	assert.Equal(70.0, result.g)
	assert.Equal(78.0, result.h)
	assert.Equal(84.0, result.i)
}

// TestAffine2D_Transform tests vector transformations
func TestAffine2D_Transform(t *testing.T) {
	assert := assert.New(t)
	// Identity transform, should not change anything
	affine1 := NewAffine2D()
	vec := Vec2{X: 1.2, Y: -1.3}
	result1 := affine1.Transform(vec)
	assert.Equal(1.2, result1.X)
	assert.Equal(-1.3, result1.Y)
	affine2 := NewAffine2DWithScale(2, 3)
	result2 := affine2.Transform(vec)
	assert.True(Equalf(2.4, result2.X))
	assert.True(Equalf(-3.9, result2.Y))
	vec2 := Vec2{X: 1, Y: 0}
	affine3 := NewAffine2DWithRotation(Deg2Rad(90))
	result3 := affine3.Transform(vec2)
	assert.True(Equalf(0, result3.X))
	assert.Equal(1.0, result3.Y)
	affine4 := NewAffine2DWithTranslation(-1, 1)
	result4 := affine4.Transform(vec)
	assert.True(Equalf(0.2, result4.X))
	assert.True(Equalf(-0.3, result4.Y))
}

// TestAffine2D_Transform tests point transformations
func TestAffine2D_TransformPoint(t *testing.T) {
	// Identity transform, should not change anything
	affine1 := NewAffine2D()
	point := Point{X: 1.2, Y: -1.3}
	result1 := affine1.TransformPoint(point)
	assert.Equal(t, 1.2, result1.X)
	assert.Equal(t, -1.3, result1.Y)
	affine2 := NewAffine2DWithScale(2, 3)
	result2 := affine2.TransformPoint(point)
	assert.True(t, Equalf(2.4, result2.X))
	assert.True(t, Equalf(-3.9, result2.Y))
	point2 := Point{X: 1, Y: 0}
	affine3 := NewAffine2DWithRotation(Deg2Rad(90))
	result3 := affine3.TransformPoint(point2)
	assert.True(t, Equalf(0, result3.X))
	assert.Equal(t, 1.0, result3.Y)
	affine4 := NewAffine2DWithTranslation(-1, 1)
	result4 := affine4.TransformPoint(point)
	assert.True(t, Equalf(0.2, result4.X))
	assert.True(t, Equalf(-0.3, result4.Y))
}

func TestAffine2D_TransformLine(t *testing.T) {
	assert := assert.New(t)
	affine1 := NewAffine2D()
	line := Line{
		P: Point{X: 1, Y: 1},
		Q: Point{X: 2, Y: 2},
	}
	result1 := affine1.TransformLine(line)
	assert.Equal(1.0, result1.P.X)
	assert.Equal(1.0, result1.P.Y)
	assert.Equal(2.0, result1.Q.X)
	assert.Equal(2.0, result1.Q.Y)
	affine2 := NewAffine2DWithShear(-0.2, 0.3)
	result2 := affine2.TransformLine(line)
	assert.Equal(0.8, result2.P.X)
	assert.Equal(1.3, result2.P.Y)
	assert.Equal(1.6, result2.Q.X)
	assert.Equal(2.6, result2.Q.Y)
}

func TestAffine2D_TransformCurve(t *testing.T) {
	assert := assert.New(t)
	affine1 := NewAffine2D()
	var curve Curve
	curve.Closed = true
	curve.AddPoint(0, 0)
	curve.AddPoint(1, 0)
	curve.AddPoint(1, 1)
	curve.AddPoint(0, 1)
	result1 := affine1.TransformCurve(curve)
	assert.Equal(0.0, result1.Points[0].X)
	assert.Equal(0.0, result1.Points[0].Y)
	assert.Equal(1.0, result1.Points[1].X)
	assert.Equal(0.0, result1.Points[1].Y)
	assert.Equal(1.0, result1.Points[2].X)
	assert.Equal(1.0, result1.Points[2].Y)
	assert.Equal(0.0, result1.Points[3].X)
	assert.Equal(1.0, result1.Points[3].Y)
	rotate45 := NewAffine2DWithRotation(Deg2Rad(45))
	result2 := rotate45.TransformCurve(curve)
	assert.True(Equalf(0.0, result2.Points[0].X))
	assert.True(Equalf(0.0, result2.Points[0].Y))
	assert.True(Equalf(math.Sqrt2/2, result2.Points[1].X))
	assert.True(Equalf(math.Sqrt2/2, result2.Points[1].Y))
	assert.True(Equalf(0.0, result2.Points[2].X))
	assert.True(Equalf(math.Sqrt2, result2.Points[2].Y))
	assert.True(Equalf(-math.Sqrt2/2, result2.Points[3].X))
	assert.True(Equalf(math.Sqrt2/2, result2.Points[3].Y))
	scale := NewAffine2DWithScale(5, 5)
	result3 := scale.TransformCurve(result2)
	assert.True(Equalf(0.0, result3.Points[0].X))
	assert.True(Equalf(0.0, result3.Points[0].Y))
	assert.True(Equalf(5*math.Sqrt2/2, result3.Points[1].X))
	assert.True(Equalf(5*math.Sqrt2/2, result3.Points[1].Y))
	assert.True(Equalf(0.0, result3.Points[2].X))
	assert.True(Equalf(5*math.Sqrt2, result3.Points[2].Y))
	assert.True(Equalf(-5*math.Sqrt2/2, result3.Points[3].X))
	assert.True(Equalf(5*math.Sqrt2/2, result3.Points[3].Y))
}

func TestAffine2D_RotateAboutPoint(t *testing.T) {
	assert := assert.New(t)
	// unit square
	var curve Curve
	curve.Closed = true
	curve.AddPoint(0, 0)
	curve.AddPoint(1, 0)
	curve.AddPoint(1, 1)
	curve.AddPoint(0, 1)
	// first translate center to orign
	centerToOrigin := NewAffine2DWithTranslation(-0.5, -0.5)
	rotation := NewAffine2DWithRotation(Deg2Rad(45))
	backToCenter := NewAffine2DWithTranslation(0.5, 0.5)
	t1 := centerToOrigin.TransformCurve(curve)
	t2 := rotation.TransformCurve(t1)
	t3 := backToCenter.TransformCurve(t2)
	assert.True(Equalf(0.5, t3.Points[0].X))
	assert.True(Equalf(0.5-math.Sqrt2/2, t3.Points[0].Y))
	assert.True(Equalf(0.5+math.Sqrt2/2, t3.Points[1].X))
	assert.True(Equalf(0.5, t3.Points[1].Y))
	assert.True(Equalf(0.5, t3.Points[2].X))
	assert.True(Equalf(0.5+math.Sqrt2/2, t3.Points[2].Y))
	assert.True(Equalf(0.5-math.Sqrt2/2, t3.Points[3].X))
	assert.True(Equalf(0.5, t3.Points[3].Y))
}

func TestAffine2D_RotateCurveAboutPoint(t *testing.T) {
	assert := assert.New(t)
	// same as above but using the builtin function
	var curve Curve
	curve.Closed = true
	curve.AddPoint(0, 0)
	curve.AddPoint(1, 0)
	curve.AddPoint(1, 1)
	curve.AddPoint(0, 1)
	affine := NewAffine2D()
	result := affine.RotateCurveAboutPoint(curve, Deg2Rad(45), Point{X: 0.5, Y: 0.5})
	assert.True(Equalf(0.5, result.Points[0].X))
	assert.True(Equalf(0.5-math.Sqrt2/2, result.Points[0].Y))
	assert.True(Equalf(0.5+math.Sqrt2/2, result.Points[1].X))
	assert.True(Equalf(0.5, result.Points[1].Y))
	assert.True(Equalf(0.5, result.Points[2].X))
	assert.True(Equalf(0.5+math.Sqrt2/2, result.Points[2].Y))
	assert.True(Equalf(0.5-math.Sqrt2/2, result.Points[3].X))
	assert.True(Equalf(0.5, result.Points[3].Y))
}
