package gaul

import (
	"fmt"
	"log"
	"math"

	"github.com/tdewolff/canvas"
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

type InteriorAngle int

const (
	CAB InteriorAngle = iota
	ABC
	BCA
)

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

// Scale multiplies point coordinates by a fixed value
func (p Point) Scale(x, y float64) Point {
	return Point{X: p.X * x, Y: p.Y * y}
}

// ScaleX scales the x value of a point
func (p Point) ScaleX(x float64) Point {
	return Point{X: p.X * x, Y: p.Y}
}

// ScaleY scales the y value of a point
func (p Point) ScaleY(y float64) Point {
	return Point{X: p.X, Y: p.Y * y}
}

// Rotate calculates a new point rotated around the origin by a given angle
func (p Point) Rotate(a float64) Point {
	x := p.X*math.Cos(a) - p.Y*math.Sin(a)
	y := p.Y*math.Sin(a) + p.Y*math.Cos(a)
	return Point{X: x, Y: y}
}

// Shear calculates a new point sheared given angles for the x and y directions
func (p Point) Shear(x, y float64) Point {
	newX := p.X + p.Y*math.Tan(x)
	newY := p.X*math.Tan(y) + p.Y
	return Point{X: newX, Y: newY}
}

// ShearX calculates a new point sheared in the x direction by a given angle
func (p Point) ShearX(a float64) Point {
	return Point{X: p.X + p.Y*math.Tan(a), Y: p.Y}
}

// ShearY calculates a new point sheared in the x direction by a given angle
func (p Point) ShearY(a float64) Point {
	return Point{X: p.X, Y: p.X*math.Tan(a) + p.Y}
}

// Reflect calculates a new point that is reflected about both the x and y axes
func (p Point) Reflect() Point {
	return p.Scale(-1, -1)
}

// ReflectX calculates a new point that is reflected about the x-axis
func (p Point) ReflectX() Point {
	return p.ScaleX(-1)
}

// ReflectY calculates a new point that is reflected about the y-axis
func (p Point) ReflectY() Point {
	return p.ScaleY(-1)
}

// Translate calculates a new point with coordinates translated by the given amounts
func (p Point) Translate(x, y float64) Point {
	return Point{X: p.X + x, Y: p.Y + y}
}

// TranslateX calculates a new point with the x coordinated translated by the given amount
func (p Point) TranslateX(x float64) Point {
	return Point{X: p.X + x, Y: p.Y}
}

// TranslateY calculates a new point with the y coordinated translated by the given amount
func (p Point) TranslateY(y float64) Point {
	return Point{X: p.X, Y: p.Y + y}
}

// Copy returns a new point with the same x and y coordinates
func (p Point) Copy() Point {
	return Point{X: p.X, Y: p.Y}
}

// ToVec2 translates the point to a Vec2 struct
func (p Point) ToVec2() Vec2 {
	return Vec2{X: p.X, Y: p.Y}
}

// Distance between two points
func Distance(p Point, q Point) float64 {
	return math.Sqrt(math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2))
}

// SquaredDistance is the square of the distance between two points
func SquaredDistance(p Point, q Point) float64 {
	return math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2)
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
	return Line{P: p, Q: q}
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
func (l Line) Scale(x, y float64) Line {
	return Line{
		P: l.P.Scale(x, y),
		Q: l.Q.Scale(x, y),
	}
}

// Rotate calculates a new line for which both points are rotated by the given angle
func (l Line) Rotate(a float64) Line {
	return Line{
		P: l.P.Rotate(a),
		Q: l.Q.Rotate(a),
	}
}

// Shear calculates a new line for which both points are sheared by the given amount
func (l Line) Shear(x, y float64) Line {
	return Line{
		P: l.P.Shear(x, y),
		Q: l.Q.Shear(x, y),
	}
}

// Translate calculates a new line for which both points are translated by the given amount
func (l Line) Translate(x, y float64) Line {
	return Line{
		P: l.P.Translate(x, y),
		Q: l.Q.Translate(x, y),
	}
}

// Copy returns a new line with the same points P and Q
func (l Line) Copy() Line {
	return Line{P: l.P.Copy(), Q: l.Q.Copy()}
}

// Boundary returns the smallest rect that contains both points in the line
func (l Line) Boundary() Rect {
	minX := math.Min(l.P.X, l.Q.X)
	minY := math.Min(l.P.Y, l.Q.Y)
	maxX := math.Max(l.P.X, l.Q.X)
	maxY := math.Max(l.P.Y, l.Q.Y)
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

// SDF calculates the signed distance from a point to a line segment
func (l Line) SDF(p Point) float64 {
	// Vector from P to Q
	dx := l.Q.X - l.P.X
	dy := l.Q.Y - l.P.Y

	// Vector from P to the point
	px := p.X - l.P.X
	py := p.Y - l.P.Y

	// Project point onto line
	dot := px*dx + py*dy
	lenSq := dx*dx + dy*dy
	t := dot / lenSq

	// Clamp t to [0,1] to stay on segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Projected point
	projX := l.P.X + t*dx
	projY := l.P.Y + t*dy

	// Distance to projected point
	return Distance(p, Point{X: projX, Y: projY})
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
	if n == 0 {
		return
	}
	ctx.MoveTo(c.Points[0].X, c.Points[0].Y)
	for i := 1; i < n; i++ {
		ctx.LineTo(c.Points[i].X, c.Points[i].Y)
	}
	if c.Closed {
		ctx.Close()
		ctx.FillStroke()
		return
	}
	ctx.Stroke()
}

// Scale calculates a new curve for which each point is scaled by the give amount
func (c *Curve) Scale(x, y float64) {
	for i := range c.Points {
		c.Points[i] = c.Points[i].Scale(x, y)
	}
}

// Rotate calculates a new curve for which each point is rotated by the given angle
func (c *Curve) Rotate(a float64) {
	for i := range c.Points {
		c.Points[i] = c.Points[i].Rotate(a)
	}
}

// Shear calculates a new curve for which each point is sheared by the given amount
func (c *Curve) Shear(x, y float64) {
	for i := range c.Points {
		c.Points[i] = c.Points[i].Shear(x, y)
	}
}

// Translate calculates a new curve for which each point is translated by the given amount
func (c *Curve) Translate(x, y float64) {
	for i := range c.Points {
		c.Points[i] = c.Points[i].Translate(x, y)
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

func (c *Curve) Area() float64 {
	if !c.Closed {
		return math.NaN()
	}
	// shoelace formula
	n := len(c.Points)
	area := 0.0
	for i := 0; i < n; i++ {
		area += c.Points[i].X*c.Points[(i+1)%n].Y - c.Points[(i+1)%n].X*c.Points[i].Y
	}
	return 0.5 * math.Abs(area)
}

// Centroid returns the centroid for the curve
func (c *Curve) Centroid() Point {
	if !c.Closed {
		N := float64(len(c.Points))
		var totalX, totalY float64
		for _, p := range c.Points {
			totalX += p.X
			totalY += p.Y
		}
		return Point{X: totalX / N, Y: totalY / N}
	}
	N := len(c.Points)
	var cx, cy float64
	A := c.Area()
	for i := 0; i < N; i++ {
		p := c.Points[i]
		q := c.Points[(i+1)%N]
		cx += (p.X + q.X) * (p.X*q.Y - q.X*p.Y)
		cy += (p.Y + q.Y) * (p.X*q.Y - q.X*p.Y)
	}
	cx /= 6 * A
	cy /= 6 * A
	return Point{X: cx, Y: cy}
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

func (c *Curve) Reverse() {
	n := len(c.Points)
	for i := 0; i < n/2; i++ {
		c.Points[i], c.Points[n-1-i] = c.Points[n-1-i], c.Points[i]
	}
}

func (c *Curve) Stitch(d *Curve) {
	c.Points = append(c.Points, d.Points...)
}

// Circle functions

// Draw draws the circle given a canvas context
func (c Circle) Draw(ctx *canvas.Context) {
	ctx.DrawPath(c.Center.X, c.Center.Y, canvas.Circle(c.Radius))
}

// ToCurve calculates a curve that approximates the circle with a given resolution (number of sides)
func (c Circle) ToCurve(resolution int) Curve {
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
func (c Circle) ContainsPoint(p Point) bool {
	return Distance(c.Center, p) <= c.Radius
}

// PointOnEdge determines if a point lies on the boundary of a circle
func (c Circle) PointOnEdge(p Point) bool {
	return Equalf(Distance(c.Center, p), c.Radius)
}

// Copy returns a new circle with the same center and radius
func (c Circle) Copy() Circle {
	return Circle{
		Center: c.Center.Copy(),
		Radius: c.Radius,
	}
}

// Boundary returns the smallest rect that contains all points on the circle
func (c Circle) Boundary() Rect {
	x := c.Center.X
	y := c.Center.Y
	r := c.Radius
	minX := x + r*math.Cos(Pi)
	minY := y + r*math.Sin(Pi/2)
	return Rect{X: minX, Y: minY, W: 2 * r, H: 2 * r}
}

// Rect functions

// ContainsPoint determines if a point lies within a rectangle
func (r Rect) ContainsPoint(p Point) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

// Contains determines if the rectangle contains a given rectangle
func (r Rect) Contains(rect Rect) bool {
	a := Point{X: r.X, Y: r.Y}
	b := Point{X: r.X + r.W, Y: r.Y + r.H}
	c := Point{X: rect.X, Y: rect.Y}
	d := Point{X: rect.X + rect.W, Y: rect.Y + rect.H}
	return a.X < c.X && a.Y < c.Y && b.X > d.X && b.Y > d.Y
}

// IsDisjoint determines if the rectangle is disjoint (no overlap) from a given rectangle
func (r Rect) IsDisjoint(rect Rect) bool {
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
func (r Rect) Overlaps(rect Rect) bool {
	return !r.IsDisjoint(rect)
}

// Intersects determines if the rectangle intersects the given rectangle
func (r Rect) Intersects(rect Rect) bool {
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
func (r Rect) ToCurve() Curve {
	var curve Curve
	curve.Closed = true
	curve.Points = append(curve.Points, Point{X: r.X, Y: r.Y})
	curve.Points = append(curve.Points, Point{X: r.X + r.W, Y: r.Y})
	curve.Points = append(curve.Points, Point{X: r.X + r.W, Y: r.Y + r.H})
	curve.Points = append(curve.Points, Point{X: r.X, Y: r.Y + r.H})
	return curve
}

// Draw draws the rectangle given a canvas context
func (r Rect) Draw(ctx *canvas.Context) {
	rect := canvas.Rectangle(r.W, r.H)
	ctx.DrawPath(r.X, r.Y, rect)
}

// Copy returns a new rectangle with the same corner, width, and height
func (r Rect) Copy() Rect {
	return Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
}

// Center returns a point at the center of the rectangle
func (r Rect) Center() Point {
	return Point{X: r.X + 0.5*r.W, Y: r.Y + 0.5*r.H}
}

func (r Rect) Subdivide(perc float64) (Rect, Rect) {
	var a, b Rect
	var w, h float64
	if r.W >= r.H {
		w = r.W * perc
		h = r.H
		a = Rect{X: r.X, Y: r.Y, W: w, H: h}
		b = Rect{X: r.X + w, Y: r.Y, W: r.W - w, H: h}
	} else {
		w = r.W
		h = r.H * perc
		a = Rect{X: r.X, Y: r.Y, W: w, H: h}
		b = Rect{X: r.X, Y: r.Y + h, W: w, H: r.H - h}
	}
	return a, b
}

// GoldenSubdivision returns two rectangles using the golden ratio to calculate where to split the rectangle
func (r Rect) GoldenSubdivision() (Rect, Rect) {
	var a, b Rect
	var w, h float64
	if r.W >= r.H {
		w = r.W / Phi
		h = r.H
		a = Rect{X: r.X, Y: r.Y, W: w, H: h}
		b = Rect{X: r.X + w, Y: r.Y, W: r.W - w, H: h}
	} else {
		w = r.W
		h = r.H / Phi
		a = Rect{X: r.X, Y: r.Y, W: w, H: h}
		b = Rect{X: r.X, Y: r.Y + h, W: w, H: r.H - h}
	}
	return a, b
}

// Triangle functions

// ToCurve calculates a closed curve with points corresponding to the vertices of the triangle
func (t Triangle) ToCurve() Curve {
	return Curve{
		Points: []Point{t.A, t.B, t.C},
		Closed: true,
	}
}

func (t Triangle) Reverse() Triangle {
	return Triangle{
		A: t.C,
		B: t.B,
		C: t.A,
	}
}

// Draw draws the triangle given a canvas context
func (t Triangle) Draw(ctx *canvas.Context) {
	ctx.MoveTo(t.A.X, t.A.Y)
	ctx.LineTo(t.B.X, t.B.Y)
	ctx.LineTo(t.C.X, t.C.Y)
	ctx.Close()
	ctx.FillStroke()
}

// Area calculates the area of the triangle
func (t Triangle) Area() float64 {
	// Heron's formula
	a := Line{P: t.A, Q: t.B}.Length()
	b := Line{P: t.B, Q: t.C}.Length()
	c := Line{P: t.C, Q: t.A}.Length()
	s := (a + b + c) / 2
	return math.Sqrt(s * (s - a) * (s - b) * (s - c))
}

// Perimeter calculates the perimeter of the triangle
func (t Triangle) Perimeter() float64 {
	a := Line{P: t.A, Q: t.B}.Length()
	b := Line{P: t.B, Q: t.C}.Length()
	c := Line{P: t.C, Q: t.A}.Length()
	return a + b + c
}

// AB returns the length of the side defined by vertices A and B
// also known as side "c" in typical notation
func (t Triangle) AB() float64 {
	return Distance(t.A, t.B)
}

// AC returns the length of the side defined by vertices A and C
// also known as side "b" in typical notation
func (t Triangle) AC() float64 {
	return Distance(t.A, t.C)
}

// BC returns the length of the side defined by vertices B and C
// also known as side "a" in typical notation
func (t Triangle) BC() float64 {
	return Distance(t.B, t.C)
}

// Angles calculates the interior angles of the triangle
func (t Triangle) Angles() (alpha, beta, gamma float64) {
	a := t.AB()
	b := t.AC()
	c := t.BC()

	// Use law of cosines with careful clamping
	cosAlpha := (b*b + c*c - a*a) / (2 * b * c)
	cosBeta := (a*a + c*c - b*b) / (2 * a * c)
	cosGamma := (a*a + b*b - c*c) / (2 * a * b)

	fmt.Printf("Angles: cosines %.10f, %.10f, %.10f\n", cosAlpha, cosBeta, cosGamma)

	// Clamp cosine values to valid range
	cosAlpha = math.Max(-1, math.Min(1, cosAlpha))
	cosBeta = math.Max(-1, math.Min(1, cosBeta))
	cosGamma = math.Max(-1, math.Min(1, cosGamma))

	alpha = math.Acos(cosAlpha)
	beta = math.Acos(cosBeta)
	gamma = math.Acos(cosGamma)

	fmt.Printf("Angles: %.10f, %.10f, %.10f\n", alpha, beta, gamma)

	return
}

// IsIsosceles determines whether the triangle has at least two sides of equal length, indicating an isosceles triangle.
func (t Triangle) IsIsosceles() bool {
	a := t.AB()
	b := t.AC()
	c := t.BC()

	fmt.Printf("IsIsosceles: sides a=%.10f, b=%.10f, c=%.10f\n", a, b, c)

	// Compare ratios with a relative tolerance
	ratio1 := math.Abs(a/b - 1)
	ratio2 := math.Abs(b/c - 1)
	ratio3 := math.Abs(a/c - 1)

	fmt.Printf("IsIsosceles: ratios %.10f, %.10f, %.10f (tolerance %.10f)\n", ratio1, ratio2, ratio3, Smol)

	return ratio1 <= Smol || ratio2 <= Smol || ratio3 <= Smol
}

// IsEquilateral determines whether the triangle has all three sides of equal length, indicating an equilateral triangle.
func (t Triangle) IsEquilateral() bool {
	a := t.AB()
	b := t.AC()
	c := t.BC()

	fmt.Printf("IsEquilateral: sides a=%.10f, b=%.10f, c=%.10f\n", a, b, c)

	// All ratios must be close to 1
	ratio1 := math.Abs(a/b - 1)
	ratio2 := math.Abs(b/c - 1)
	ratio3 := math.Abs(a/c - 1)

	fmt.Printf("IsEquilateral: ratios %.10f, %.10f, %.10f (tolerance %.10f)\n", ratio1, ratio2, ratio3, Smol)

	return ratio1 <= Smol && ratio2 <= Smol && ratio3 <= Smol
}

// IsRight determines whether the triangle has one right angle, indicating a right triangle.
func (t Triangle) IsRight() bool {
	alpha, beta, gamma := t.Angles()
	rightAngle := math.Pi / 2

	// Check if any angle is approximately 90 degrees
	return math.Abs(alpha-rightAngle) <= Smol ||
		math.Abs(beta-rightAngle) <= Smol ||
		math.Abs(gamma-rightAngle) <= Smol
}

// IsAcute determines whether the triangle has all angles less than 90 degrees, indicating an acute triangle.
func (t Triangle) IsAcute() bool {
	alpha, beta, gamma := t.Angles()
	rightAngle := math.Pi / 2

	// All angles must be less than 90 degrees
	return alpha < rightAngle-Smol &&
		beta < rightAngle-Smol &&
		gamma < rightAngle-Smol
}

// IsObtuse determines whether the triangle has one angle greater than 90 degrees, indicating an obtuse triangle.
func (t Triangle) IsObtuse() bool {
	alpha, beta, gamma := t.Angles()
	rightAngle := math.Pi / 2

	// Any angle greater than 90 degrees
	return alpha > rightAngle+Smol ||
		beta > rightAngle+Smol ||
		gamma > rightAngle+Smol
}

func (t Triangle) ContainsPoint(p Point) bool {
	// Calculate barycentric coordinates
	denominator := ((t.B.Y-t.C.Y)*(t.A.X-t.C.X) + (t.C.X-t.B.X)*(t.A.Y-t.C.Y))
	alpha := ((t.B.Y-t.C.Y)*(p.X-t.C.X) + (t.C.X-t.B.X)*(p.Y-t.C.Y)) / denominator
	beta := ((t.C.Y-t.A.Y)*(p.X-t.C.X) + (t.A.X-t.C.X)*(p.Y-t.C.Y)) / denominator
	gamma := 1.0 - alpha - beta

	// Point is inside if all coordinates are between 0 and 1
	return alpha >= -Smol && beta >= -Smol && gamma >= -Smol
}

// IncircleRadius calculates the radius of the incircle
func (t Triangle) IncircleRadius() float64 {
	return 2 * t.Area() / t.Perimeter()
}

// CircumcircleRadius calculates the radius of the circumcircle
func (t Triangle) CircumcircleRadius() float64 {
	return (t.AB() * t.BC() * t.AC()) / (4 * t.Area())
}

// NinePointRadius calculates the radius of the nine-point circle
func (t Triangle) NinePointRadius() float64 {
	return t.CircumcircleRadius() / 2
}

// Centroid calculates the centroid of the triangle
func (t Triangle) Centroid() Point {
	x := (t.A.X + t.B.X + t.C.X) / 3
	y := (t.A.Y + t.B.Y + t.C.Y) / 3
	return Point{X: x, Y: y}
}

// Incenter calculates the incenter of the triangle
func (t Triangle) Incenter() Point {
	a := t.BC()
	b := t.AC()
	c := t.AB()
	x := (a*t.A.X + b*t.B.X + c*t.C.X) / (a + b + c)
	y := (a*t.A.Y + b*t.B.Y + c*t.C.Y) / (a + b + c)
	return Point{X: x, Y: y}
}

// Circumcenter calculates the circumcenter of the triangle
func (t Triangle) Circumcenter() Point {
	D := 2 * (t.A.X*(t.B.Y-t.C.Y) + t.B.X*(t.C.Y-t.A.Y) + t.C.X*(t.A.Y-t.B.Y))
	A2 := t.A.X*t.A.X + t.A.Y*t.A.Y
	B2 := t.B.X*t.B.X + t.B.Y*t.B.Y
	C2 := t.C.X*t.C.X + t.C.Y*t.C.Y
	return Point{
		X: (A2*(t.B.Y-t.C.Y) + B2*(t.C.Y-t.A.Y) + C2*(t.A.Y-t.B.Y)) / D,
		Y: (A2*(t.C.X-t.B.X) + B2*(t.A.X-t.C.X) + C2*(t.B.X-t.A.X)) / D,
	}
}

// Orthocenter calculates the orthocenter of the triangle
func (t Triangle) Orthocenter() Point {
	alpha, beta, gamma := t.Angles()
	ta := math.Tan(alpha)
	tb := math.Tan(beta)
	tc := math.Tan(gamma)
	return Point{
		X: (ta*t.A.X + tb*t.B.X + tc*t.C.X) / (ta + tb + tc),
		Y: (ta*t.A.Y + tb*t.B.Y + tc*t.C.Y) / (ta + tb + tc),
	}
}

// NinePointCenter calculates the center point of the nine-point circle
func (t Triangle) NinePointCenter() Point {
	line := Line{
		P: t.Circumcenter(),
		Q: t.Orthocenter(),
	}
	return line.Midpoint()
}

// Incircle returns the incircle of the triangle
func (t Triangle) Incircle() Circle {
	return Circle{
		Center: t.Incenter(),
		Radius: t.IncircleRadius(),
	}
}

// Circumcircle returns the circumcircle of the triangle
func (t Triangle) Circumcircle() Circle {
	return Circle{
		Center: t.Circumcenter(),
		Radius: t.CircumcircleRadius(),
	}
}

// NinePointCircle returns the nine-point circle of the triangle
func (t Triangle) NinePointCircle() Circle {
	return Circle{
		Center: t.NinePointCenter(),
		Radius: t.NinePointRadius(),
	}
}

func (t Triangle) SDF(p Point) float64 {
	// Calculate vectors for each edge
	ab := Line{P: t.A, Q: t.B}
	bc := Line{P: t.B, Q: t.C}
	ca := Line{P: t.C, Q: t.A}

	// Calculate distances to each edge
	d1 := ab.SDF(p)
	d2 := bc.SDF(p)
	d3 := ca.SDF(p)

	// Calculate cross products to determine which side of each line the point is on
	cp1 := (t.B.X-t.A.X)*(p.Y-t.A.Y) - (t.B.Y-t.A.Y)*(p.X-t.A.X)
	cp2 := (t.C.X-t.B.X)*(p.Y-t.B.Y) - (t.C.Y-t.B.Y)*(p.X-t.B.X)
	cp3 := (t.A.X-t.C.X)*(p.Y-t.C.Y) - (t.A.Y-t.C.Y)*(p.X-t.C.X)

	// Point is inside if all cross products have the same sign
	inside := (cp1 >= 0 && cp2 >= 0 && cp3 >= 0) || (cp1 <= 0 && cp2 <= 0 && cp3 <= 0)

	// Find minimum distance to any edge
	minDist := math.Min(math.Min(d1, d2), d3)

	if inside {
		return -minDist
	}
	return minDist
}

// ClosestPoint returns the closest point on the triangle to the given point
func (t Triangle) ClosestPoint(p Point) Point {
	// Check if point is inside triangle
	if t.ContainsPoint(p) {
		return p
	}

	// Find closest point on each edge
	closest1 := Line{P: t.A, Q: t.B}.ClosestPoint(p)
	closest2 := Line{P: t.B, Q: t.C}.ClosestPoint(p)
	closest3 := Line{P: t.C, Q: t.A}.ClosestPoint(p)

	// Calculate distances to each closest point
	dist1 := Distance(p, closest1)
	dist2 := Distance(p, closest2)
	dist3 := Distance(p, closest3)

	// Return the closest point
	if dist1 <= dist2 && dist1 <= dist3 {
		return closest1
	}
	if dist2 <= dist1 && dist2 <= dist3 {
		return closest2
	}
	return closest3
}

// ClosestPoint returns the closest point on the line segment to the given point
func (l Line) ClosestPoint(p Point) Point {
	// Vector from P to Q
	dx := l.Q.X - l.P.X
	dy := l.Q.Y - l.P.Y

	// Vector from P to the point
	px := p.X - l.P.X
	py := p.Y - l.P.Y

	// Project point onto line
	dot := px*dx + py*dy
	lenSq := dx*dx + dy*dy
	t := dot / lenSq

	// Clamp t to [0,1] to stay on segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Return projected point
	return Point{
		X: l.P.X + t*dx,
		Y: l.P.Y + t*dy,
	}
}

// Copy returns a new triangle with the same vertices
func (t Triangle) Copy() Triangle {
	return Triangle{
		A: t.A.Copy(),
		B: t.B.Copy(),
		C: t.C.Copy(),
	}
}

// Boundary returns the smallest rect that contains all three vertices
func (t Triangle) Boundary() Rect {
	curve := t.ToCurve()
	return curve.Boundary()
}

// BisectSubdivision returns a slice of triangles that are the result of subdividing the triangle into two
// smaller triangles by bisecting one angle with a variable percent offset along the opposite side
func (t Triangle) BisectSubdivision(angle InteriorAngle, percent float64) []Triangle {
	result := make([]Triangle, 2)
	var vertex1, vertex2, vertex3, vertex4 Point
	var side Line
	switch angle {
	case CAB:
		side = Line{P: t.B, Q: t.C}
		vertex1 = t.A
		vertex3 = t.B
		vertex4 = t.C
	case ABC:
		side = Line{P: t.C, Q: t.A}
		vertex1 = t.B
		vertex3 = t.C
		vertex4 = t.A
	case BCA:
		side = Line{P: t.A, Q: t.B}
		vertex1 = t.C
		vertex3 = t.A
		vertex4 = t.B
	}
	vertex2 = side.Lerp(percent)
	result[0] = Triangle{A: vertex1, B: vertex3, C: vertex2}
	result[1] = Triangle{A: vertex1, B: vertex4, C: vertex2}
	return result
}

// BarycentricSubdivision returns a slice of triangles that are the result of subdividing the triangle into three
// smaller triangles using the centroid and each vertex as the new vertices
func (t Triangle) BarycentricSubdivision() []Triangle {
	result := make([]Triangle, 3)
	centroid := t.Centroid()
	result[0] = Triangle{A: t.A, B: t.B, C: centroid}
	result[1] = Triangle{A: t.B, B: t.C, C: centroid}
	result[2] = Triangle{A: t.C, B: t.A, C: centroid}
	return result
}

// IncenterSubdivision returns a slice of triangles that are the result of subdividing the triangle into three
// smaller triangles using the incenter and each vertex as the new vertices
func (t Triangle) IncenterSubdivision() []Triangle {
	result := make([]Triangle, 3)
	incenter := t.Incenter()
	result[0] = Triangle{A: t.A, B: t.B, C: incenter}
	result[1] = Triangle{A: t.B, B: t.C, C: incenter}
	result[2] = Triangle{A: t.C, B: t.A, C: incenter}
	return result
}

// CircumcenterSubdivision returns a slice of triangles that are the result of subdividing the triangle into three
// smaller triangles using the circumcenter and each vertex as the new vertices
func (t Triangle) CircumcenterSubdivision() []Triangle {
	result := make([]Triangle, 3)
	circumcenter := t.Circumcenter()
	result[0] = Triangle{A: t.A, B: t.B, C: circumcenter}
	result[1] = Triangle{A: t.B, B: t.C, C: circumcenter}
	result[2] = Triangle{A: t.C, B: t.A, C: circumcenter}
	return result
}

// OrthocenterSubdivision returns a slice of triangles that are the result of subdividing the triangle into three
// smaller triangles using the orthocenter and each vertex as the new vertices
func (t Triangle) OrthocenterSubdivision() []Triangle {
	result := make([]Triangle, 3)
	orthocenter := t.Orthocenter()
	result[0] = Triangle{A: t.A, B: t.B, C: orthocenter}
	result[1] = Triangle{A: t.B, B: t.C, C: orthocenter}
	result[2] = Triangle{A: t.C, B: t.A, C: orthocenter}
	return result
}

// MidpointSubdivision returns a slice of triangles that are the result of subdividing the triangle into four
// smaller triangles using the midpoints of each side as the set of new vertices
func (t Triangle) MidpointSubdivision() []Triangle {
	result := make([]Triangle, 4)
	c := Line{P: t.A, Q: t.B}.Midpoint()
	a := Line{P: t.B, Q: t.C}.Midpoint()
	b := Line{P: t.C, Q: t.A}.Midpoint()
	result[0] = Triangle{A: t.A, B: c, C: b}
	result[1] = Triangle{A: c, B: t.B, C: a}
	result[2] = Triangle{A: b, B: a, C: t.C}
	result[3] = Triangle{A: a, B: b, C: c}
	return result
}

func (t Triangle) MidAndThirdsSubdivision(angle InteriorAngle) []Triangle {
	result := make([]Triangle, 5)
	var p1, p2, q1, q2, v1, v2, v3 Point
	switch angle {
	case BCA:
		v1 = t.C
		v2 = t.A
		v3 = t.B
	case CAB:
		v1 = t.A
		v2 = t.B
		v3 = t.C
	case ABC:
		v1 = t.B
		v2 = t.C
		v3 = t.A
	}
	p1 = Line{P: v2, Q: v3}.Lerp(1.0 / 3.0)
	p2 = Line{P: v2, Q: v3}.Lerp(2.0 / 3.0)
	q1 = Line{P: v1, Q: v2}.Midpoint()
	q2 = Line{P: v1, Q: v3}.Midpoint()
	result[0] = Triangle{A: v1, B: q1, C: q2}
	result[1] = Triangle{A: q1, B: v2, C: p1}
	result[2] = Triangle{A: q2, B: p2, C: v3}
	result[3] = Triangle{A: p1, B: p2, C: q1}
	result[4] = Triangle{A: p2, B: q2, C: q1}
	return result
}
