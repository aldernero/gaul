package gaul

import (
	"fmt"
	"github.com/tdewolff/canvas"
	"log"
	"math"
)

// Primitive types

// Point is a simple point in 2D space
type Point struct {
	X float64
	Y float64
}

// Line is two points that form a line
type Line struct {
	P Point
	Q Point
}

// Curve A curve is a list of points, may be closed
type Curve struct {
	Points []Point
	Closed bool
}

// A Circle represented by a center point and radius
type Circle struct {
	Center Point
	Radius float64
}

// Rect is a simple rectangle
type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

// A Triangle specified by vertices as points
type Triangle struct {
	A Point
	B Point
	C Point
}

// Point functions

// Tuple representation of a point, useful for debugging
func (p Point) String() string {
	return fmt.Sprintf("(%f, %f)", p.X, p.Y)
}

// Lerp is a linear interpolation between two points
func (p Point) Lerp(a Point, i float64) Point {
	return Point{
		X: Lerp(p.X, a.X, i),
		Y: Lerp(p.Y, a.Y, i),
	}
}

// IsEqual determines if two points are equal
func (p Point) IsEqual(q Point) bool {
	return p.X == q.X && p.Y == q.Y
}

// Draw draws the point as a circle with a given radius using a canvas context
func (p Point) Draw(s float64, ctx *canvas.Context) {
	ctx.DrawPath(p.X, p.Y, canvas.Circle(s))
}

// Distance between two points
func Distance(p Point, q Point) float64 {
	return math.Sqrt(math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2))
}

// SquaredDistance is the square of the distance between two points
func SquaredDistance(p Point, q Point) float64 {
	return math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2)
}

// Scale multiplies point coordinates by a fixed value
func (p *Point) Scale(x, y float64) {
	p.X *= x
	p.Y *= y
}

// ScaleX scales the x value of a point
func (p *Point) ScaleX(x float64) {
	p.X *= x
}

// ScaleY scales the y value of a point
func (p *Point) ScaleY(y float64) {
	p.Y *= y
}

// Rotate calculates a new point rotated around the origin by a given angle
func (p *Point) Rotate(a float64) {
	x := p.X*math.Cos(a) - p.Y*math.Sin(a)
	y := p.Y*math.Sin(a) + p.Y*math.Cos(a)
	p.X = x
	p.Y = y
}

// Shear calculates a new point sheared given angles for the x and y directions
func (p *Point) Shear(x, y float64) {
	newX := p.X + p.Y*math.Tan(x)
	newY := p.X*math.Tan(y) + p.Y
	p.X = newX
	p.Y = newY
}

// ShearX calculates a new point sheared in the x direction by a given angle
func (p *Point) ShearX(a float64) {
	p.X = p.X + p.Y*math.Tan(a)
}

// ShearY calculates a new point sheared in the x direction by a given angle
func (p *Point) ShearY(a float64) {
	p.Y = p.X*math.Tan(a) + p.Y
}

// Reflect calculates a new point that is reflected about both the x and y axes
func (p *Point) Reflect() {
	p.Scale(-1, -1)
}

// ReflectX calculates a new point that is reflected about the x-axis
func (p *Point) ReflectX() {
	p.ScaleX(-1)
}

// ReflectY calculates a new point that is reflected about the y-axis
func (p *Point) ReflectY() {
	p.ScaleY(-1)
}

// Translate calculates a new point with coordinates translated by the given amounts
func (p *Point) Translate(x, y float64) {
	p.X += x
	p.Y += y
}

// TranslateX calculates a new point with the x coordinated translated by the given amount
func (p *Point) TranslateX(x float64) {
	p.X += x
}

// TranslateY calculates a new point with the y coordinated translated by the given amount
func (p *Point) TranslateY(y float64) {
	p.Y += y
}

// Copy returns a new point with the same x and y coordinates
func (p *Point) Copy() Point {
	return Point{X: p.X, Y: p.Y}
}

// Line functions
// String representation of a line, useful for debugging
func (l Line) String() string {
	return fmt.Sprintf("(%f, %f) -> (%f, %f)", l.P.X, l.P.Y, l.Q.X, l.Q.Y)
}

// IsEqual determines if two lines are equal to each other
func (l Line) IsEqual(k Line) bool {
	return l.P.IsEqual(k.P) && l.Q.IsEqual(k.Q)
}

// Angle calculates the angle between the line and the x-axis
func (l Line) Angle() float64 {
	dy := l.Q.Y - l.P.Y
	dx := l.Q.X - l.P.X
	angle := math.Atan(dy / dx)
	return angle
}

// Slope computes the slope of the line
func (l Line) Slope() float64 {
	dy := l.Q.Y - l.P.Y
	dx := l.Q.X - l.P.X
	if math.Abs(dx) < Smol {
		if dx < 0 {
			if dy > 0 {
				return math.Inf(-1)
			} else {
				return math.Inf(1)
			}
		} else {
			if dy > 0 {
				return math.Inf(1)
			} else {
				return math.Inf(-1)
			}
		}
	}
	return dy / dx
}

// InvertedSlope calculates one over the slope of the line
func (l Line) InvertedSlope() float64 {
	slope := l.Slope()
	if math.IsInf(slope, 1) || math.IsInf(slope, -1) {
		return 0
	}
	return -1 / slope
}

// PerpendicularAt calculates a line at a given percentage along the line with a given length that is perpendicular
// to the original line
func (l Line) PerpendicularAt(percentage float64, length float64) Line {
	angle := l.Angle()
	point := l.P.Lerp(l.Q, percentage)
	sinOffset := 0.5 * length * math.Sin(angle)
	cosOffset := 0.5 * length * math.Cos(angle)
	p := Point{
		X: NoTinyVals(point.X - sinOffset),
		Y: NoTinyVals(point.Y + cosOffset),
	}
	q := Point{
		X: NoTinyVals(point.X + sinOffset),
		Y: NoTinyVals(point.Y - cosOffset),
	}
	return Line{
		P: p,
		Q: q,
	}
}

// PerpendicularBisector calculates a line with a given length at the midpoint of the original line that is also
// perpendicular to the line
func (l Line) PerpendicularBisector(length float64) Line {
	return l.PerpendicularAt(0.5, length)
}

// Lerp is an interpolation between the two points of a line
func (l Line) Lerp(i float64) Point {
	return Point{
		X: Lerp(l.P.X, l.Q.X, i),
		Y: Lerp(l.P.Y, l.Q.Y, i),
	}
}

// Draw draws the line given a canvas context
func (l Line) Draw(ctx *canvas.Context) {
	ctx.MoveTo(l.P.X, l.P.Y)
	ctx.LineTo(l.Q.X, l.Q.Y)
	ctx.Stroke()
}

// Midpoint Calculates the midpoint between two points
func Midpoint(p Point, q Point) Point {
	return Point{X: 0.5 * (p.X + q.X), Y: 0.5 * (p.Y + q.Y)}
}

// Midpoint Calculates the midpoint of a line
func (l Line) Midpoint() Point {
	return Midpoint(l.P, l.Q)
}

// Length Calculates the length of a line
func (l Line) Length() float64 {
	return Distance(l.P, l.Q)
}

// Intersects determines if two lines intersect each other
func (l Line) Intersects(k Line) bool {
	a1 := l.Q.X - l.P.X
	b1 := k.P.X - k.Q.X
	c1 := k.P.X - l.P.X
	a2 := l.Q.Y - l.P.Y
	b2 := k.P.Y - k.Q.Y
	c2 := k.P.Y - l.P.Y
	d := a1*b2 - a2*b1
	if d == 0 {
		// lines are parallel
		return false
	}
	// Cramer's rule
	s := (c1*b2 - c2*b1) / d
	t := (a1*c2 - a2*c1) / d
	return s >= 0 && t >= 0 && s <= 1 && t <= 1
}

// ParallelTo determines if two lines are parallel
func (l Line) ParallelTo(k Line) bool {
	a1 := l.Q.X - l.P.X
	b1 := k.P.X - k.Q.X
	a2 := l.Q.Y - l.P.Y
	b2 := k.P.Y - k.Q.Y
	d := a1*b2 - a2*b1
	return d == 0
}

// Scale calculates a new line for which both points are scaled by the given amount
func (l *Line) Scale(x, y float64) {
	l.P.Scale(x, y)
	l.Q.Scale(x, y)
}

// Rotate calculates a new line for which both points are rotated by the given angle
func (l *Line) Rotate(a float64) {
	l.P.Rotate(a)
	l.Q.Rotate(a)
}

// Shear calculates a new line for which both points are sheared by the given amount
func (l *Line) Shear(x, y float64) {
	l.P.Shear(x, y)
	l.Q.Shear(x, y)
}

// Translate calculates a new line for which both points are translated by the given amount
func (l *Line) Translate(x, y float64) {
	l.P.Translate(x, y)
	l.Q.Translate(x, y)
}

// Copy returns a new line with the same points P and Q
func (l *Line) Copy() Line {
	return Line{P: l.P.Copy(), Q: l.Q.Copy()}
}

// Boundary returns the smallest rect that contains both points in the line
func (l *Line) Boundary() Rect {
	minX := math.Min(l.P.X, l.Q.X)
	minY := math.Min(l.P.Y, l.Q.Y)
	maxX := math.Max(l.P.X, l.Q.X)
	maxY := math.Max(l.P.Y, l.Q.Y)
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

// Curve functions

// Length Calculates the length of the line segments of a curve
func (c *Curve) Length() float64 {
	result := 0.0
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		result += Distance(c.Points[i], c.Points[i+1])
	}
	if c.Closed {
		result += Distance(c.Points[0], c.Points[n-1])
	}
	return result
}

// Last returns the last point in a curve
func (c *Curve) Last() Point {
	n := len(c.Points)
	switch n {
	case 0:
		return Point{
			X: 0,
			Y: 0,
		}
	case 1:
		return c.Points[0]
	}
	if c.Closed {
		return c.Points[0]
	}
	return c.Points[n-1]
}

// LastLine returns the last line in a curve
func (c *Curve) LastLine() Line {
	n := len(c.Points)
	switch n {
	case 0:
		return Line{
			P: Point{X: 0, Y: 0},
			Q: Point{X: 0, Y: 0},
		}
	case 1:
		return Line{
			P: c.Points[0],
			Q: c.Points[0],
		}
	}
	if c.Closed {
		return Line{
			P: c.Points[n-1],
			Q: c.Points[0],
		}
	}
	return Line{
		P: c.Points[n-2],
		Q: c.Points[n-1],
	}
}

// AddPoint appends a point to the curve
func (c *Curve) AddPoint(x, y float64) {
	c.Points = append(c.Points, Point{X: x, Y: y})
}

// Lerp calculates a point a given percentage along a curve
func (c *Curve) Lerp(percentage float64) Point {
	var point Point
	if percentage < 0 || percentage > 1 {
		log.Fatalf("percentage in Lerp not between 0 and 1: %v\n", percentage)
	}
	if NoTinyVals(percentage) == 0 {
		return c.Points[0]
	}
	if math.Abs(percentage-1) < Smol {
		return c.Last()
	}
	totalDist := c.Length()
	targetDist := percentage * totalDist
	partialDist := 0.0
	var foundPoint bool
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		dist := Distance(c.Points[i], c.Points[i+1])
		if partialDist+dist >= targetDist {
			remainderDist := targetDist - partialDist
			pct := remainderDist / dist
			point = c.Points[i].Lerp(c.Points[i+1], pct)
			foundPoint = true
			break
		}
		partialDist += dist
	}
	if !foundPoint {
		if c.Closed {
			dist := Distance(c.Points[n-1], c.Points[0])
			remainderDist := targetDist - partialDist
			pct := remainderDist / dist
			point = c.Points[n-1].Lerp(c.Points[0], pct)
		} else {
			panic("couldn't find curve lerp point")
		}
	}
	return point
}

// LineAt returns the line segment in a curve that is closest to the given percentage along the curve's length
func (c *Curve) LineAt(percentage float64) (Line, float64) {
	var line Line
	var linePct float64
	if percentage < 0 || percentage > 1 {
		log.Fatalf("percentage in Lerp not between 0 and 1: %v\n", percentage)
	}
	if NoTinyVals(percentage) == 0 {
		return Line{P: c.Points[0], Q: c.Points[1]}, 0
	}
	if math.Abs(percentage-1) < Smol {
		return c.LastLine(), 1
	}
	totalDist := c.Length()
	targetDist := percentage * totalDist
	partialDist := 0.0
	var foundPoint bool
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		dist := Distance(c.Points[i], c.Points[i+1])
		if partialDist+dist >= targetDist {
			remainderDist := targetDist - partialDist
			linePct = remainderDist / dist
			line.P = c.Points[i]
			line.Q = c.Points[i+1]
			foundPoint = true
			break
		}
		partialDist += dist
	}
	if !foundPoint {
		if c.Closed {
			dist := Distance(c.Points[n-1], c.Points[0])
			remainderDist := targetDist - partialDist
			linePct = remainderDist / dist
			line.P = c.Points[n-1]
			line.Q = c.Points[0]
		} else {
			panic("couldn't find curve lerp point")
		}
	}
	return line, linePct
}

// PerpendicularAt calculates a line that is perpendicular to the curve at a given percentage along the curve's length
func (c *Curve) PerpendicularAt(percentage float64, length float64) Line {
	line, linePct := c.LineAt(percentage)
	return line.PerpendicularAt(linePct, length)
}

// Draw draws the curve given a canvas context
func (c *Curve) Draw(ctx *canvas.Context) {
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		ctx.MoveTo(c.Points[i].X, c.Points[i].Y)
		ctx.LineTo(c.Points[i+1].X, c.Points[i+1].Y)
	}
	if c.Closed {
		ctx.MoveTo(c.Points[n-1].X, c.Points[n-1].Y)
		ctx.LineTo(c.Points[0].X, c.Points[0].Y)
	}
	ctx.Stroke()
}

// Scale calculates a new curve for which each point is scaled by the give amount
func (c *Curve) Scale(x, y float64) {
	for _, p := range c.Points {
		p.Scale(x, y)
	}
}

// Rotate calculates a new curve for which each point is rotated by the given angle
func (c *Curve) Rotate(a float64) {
	for _, p := range c.Points {
		p.Rotate(a)
	}
}

// Shear calculates a new curve for which each point is sheared by the given amount
func (c *Curve) Shear(x, y float64) {
	for _, p := range c.Points {
		p.Shear(x, y)
	}
}

// Translate calculates a new curve for which each point is translated by the given amount
func (c *Curve) Translate(x, y float64) {
	for _, p := range c.Points {
		p.Translate(x, y)
	}
}

// Copy returns a new line with the same points and closed property
func (c *Curve) Copy() Curve {
	var curve Curve
	curve.Closed = true
	for _, p := range c.Points {
		curve.AddPoint(p.X, p.Y)
	}
	return curve
}

// Centroid returns the centroid for the curve
func (c *Curve) Centroid() Point {
	N := float64(len(c.Points))
	var totalX, totalY float64
	for _, p := range c.Points {
		totalX += p.X
		totalY += p.Y
	}
	return Point{X: totalX / N, Y: totalY / N}
}

// Boundary returns the smallest rect that contains all the points in the curve
func (c *Curve) Boundary() Rect {
	minX := math.Inf(1)
	minY := math.Inf(1)
	maxX := math.Inf(-1)
	maxY := math.Inf(-1)
	for _, p := range c.Points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

// Circle functions

// Draw draws the circle given a canvas context
func (c *Circle) Draw(ctx *canvas.Context) {
	ctx.DrawPath(c.Center.X, c.Center.Y, canvas.Circle(c.Radius))
}

// ToCurve calculates a curve that approximates the circle with a given resolution (number of sides)
func (c *Circle) ToCurve(resolution int) Curve {
	points := make([]Point, resolution)
	theta := Linspace(0, Tau, resolution, false)
	for i, t := range theta {
		x := c.Center.X + c.Radius*math.Cos(t)
		y := c.Center.Y + c.Radius*math.Sin(t)
		points[i] = Point{X: x, Y: y}
	}
	return Curve{Points: points, Closed: true}
}

// ContainsPoint determines if a point lies inside the circle, including the boundary
func (c *Circle) ContainsPoint(p Point) bool {
	return Distance(c.Center, p) <= c.Radius
}

// PointOnEdge determines if a point lies on the boundary of a circle
func (c *Circle) PointOnEdge(p Point) bool {
	return Equalf(Distance(c.Center, p), c.Radius)
}

// Copy returns a new circle with the same center and radius
func (c *Circle) Copy() Circle {
	return Circle{
		Center: c.Center.Copy(),
		Radius: c.Radius,
	}
}

// Boundary returns the smallest rect that contains all points on the circle
func (c *Circle) Boundary() Rect {
	x := c.Center.X
	y := c.Center.Y
	r := c.Radius
	minX := x + r*math.Cos(Pi)
	minY := y + r*math.Sin(Pi/2)
	return Rect{X: minX, Y: minY, W: 2 * r, H: 2 * r}
}

// Rect functions

// ContainsPoint determines if a point lies within a rectangle
func (r *Rect) ContainsPoint(p Point) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

// Contains determines if the rectangle contains a given rectangle
func (r *Rect) Contains(rect Rect) bool {
	a := Point{X: r.X, Y: r.Y}
	b := Point{X: r.X + r.W, Y: r.Y + r.H}
	c := Point{X: rect.X, Y: rect.Y}
	d := Point{X: rect.X + rect.W, Y: rect.Y + rect.H}
	return a.X < c.X && a.Y < c.Y && b.X > d.X && b.Y > d.Y
}

// IsDisjoint determines if the rectangle is disjoint (no overlap) from a given rectangle
func (r *Rect) IsDisjoint(rect Rect) bool {
	aLeft := r.X
	aRight := r.X + r.W
	aTop := r.Y + r.H
	aBottom := r.Y
	bLeft := rect.X
	bRight := rect.X + rect.W
	bTop := rect.Y + rect.H
	bBottom := rect.Y

	if aLeft > bRight || aBottom > bTop || aRight < bLeft || aTop < bBottom {
		return true
	}
	return false
}

// Overlaps determines if the rectangle overlaps the given rectangle
func (r *Rect) Overlaps(rect Rect) bool {
	return !r.IsDisjoint(rect)
}

// Intersects determines if the rectangle intersects the given rectangle
func (r *Rect) Intersects(rect Rect) bool {
	a := Point{X: r.X, Y: r.Y}
	b := Point{X: r.X + r.W, Y: r.Y + r.H}
	c := Point{X: rect.X, Y: rect.Y}
	d := Point{X: rect.X + rect.W, Y: rect.Y + rect.H}

	if a.X >= d.X || c.X >= b.X {
		return false
	}

	if b.Y >= c.Y || d.Y >= a.Y {
		return false
	}

	return true
}

// ToCurve calculates a closed curve with points corresponding to the vertices of the rectangle
func (r *Rect) ToCurve() Curve {
	var curve Curve
	curve.Closed = true
	curve.Points = append(curve.Points, Point{X: r.X, Y: r.Y})
	curve.Points = append(curve.Points, Point{X: r.X + r.W, Y: r.Y})
	curve.Points = append(curve.Points, Point{X: r.X + r.W, Y: r.Y + r.H})
	curve.Points = append(curve.Points, Point{X: r.X, Y: r.Y + r.H})
	return curve
}

// Draw draws the rectangle given a canvas context
func (r *Rect) Draw(ctx *canvas.Context) {
	rect := canvas.Rectangle(r.W, r.H)
	ctx.DrawPath(r.X, r.Y, rect)
}

// Copy returns a new rectangle with the same corner, width, and height
func (r *Rect) Copy() Rect {
	return Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
}

// Center returns a point at the center of the rectangle
func (r *Rect) Center() Point {
	return Point{X: r.X + 0.5*r.W, Y: r.Y + 0.5*r.H}
}

// Triangle functions

// ToCurve calculates a closed curve with points corresponding to the vertices of the triangle
func (t *Triangle) ToCurve() Curve {
	return Curve{
		Points: []Point{t.A, t.B, t.C},
		Closed: true,
	}
}

// Draw draws the triangle given a canvas context
func (t *Triangle) Draw(ctx *canvas.Context) {
	ctx.MoveTo(t.A.X, t.A.Y)
	ctx.LineTo(t.B.X, t.B.Y)
	ctx.LineTo(t.C.X, t.C.Y)
	ctx.Close()
}

// Area calculates the area of the triangle
func (t *Triangle) Area() float64 {
	// Heron's formula
	a := Line{P: t.A, Q: t.B}.Length()
	b := Line{P: t.B, Q: t.C}.Length()
	c := Line{P: t.C, Q: t.A}.Length()
	s := (a + b + c) / 2
	return math.Sqrt(s * (s - a) * (s - b) * (s - c))
}

// Perimeter calculates the perimeter of the triangle
func (t *Triangle) Perimeter() float64 {
	a := Line{P: t.A, Q: t.B}.Length()
	b := Line{P: t.B, Q: t.C}.Length()
	c := Line{P: t.C, Q: t.A}.Length()
	return a + b + c
}

// Centroid calculates the centroid of the triangle
func (t *Triangle) Centroid() Point {
	x := (t.A.X + t.B.X + t.C.X) / 3
	y := (t.A.Y + t.B.Y + t.C.Y) / 3
	return Point{X: x, Y: y}
}

// Copy returns a new triangle with the same vertices
func (t *Triangle) Copy() Triangle {
	return Triangle{
		A: t.A.Copy(),
		B: t.B.Copy(),
		C: t.C.Copy(),
	}
}

// Boundary returns the smallest rect that contains all three vertices
func (t *Triangle) Boundary() Rect {
	curve := t.ToCurve()
	return curve.Boundary()
}
