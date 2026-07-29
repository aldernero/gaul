package gaul

import (
	"image/color"
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The tests in this file pin down behavior that was previously wrong. Each one
// fails against the pre-fix implementation.

func TestPointRotate(t *testing.T) {
	assert := assert.New(t)
	got := Point{X: 1, Y: 0}.Rotate(Pi / 2)
	assert.InDelta(0.0, got.X, Smol)
	assert.InDelta(1.0, got.Y, Smol)

	// Rotating by a full turn is the identity.
	p := Point{X: 3, Y: -7}
	full := p.Rotate(Tau)
	assert.InDelta(p.X, full.X, Smol)
	assert.InDelta(p.Y, full.Y, Smol)

	// Rotation preserves magnitude.
	q := Point{X: 3, Y: 4}.Rotate(0.7)
	assert.InDelta(5.0, math.Hypot(q.X, q.Y), Smol)
}

func TestRectIntersects(t *testing.T) {
	assert := assert.New(t)
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	assert.True(a.Intersects(Rect{X: 5, Y: 5, W: 10, H: 10}), "overlapping corner")
	assert.True(a.Intersects(Rect{X: 2, Y: 2, W: 3, H: 3}), "fully contained")
	assert.False(a.Intersects(Rect{X: 100, Y: 100, W: 1, H: 1}), "far away")
	assert.False(a.Intersects(Rect{X: 0, Y: 20, W: 10, H: 5}), "disjoint in y only")
	assert.False(a.Intersects(Rect{X: 20, Y: 0, W: 5, H: 10}), "disjoint in x only")
	// Intersects is symmetric.
	b := Rect{X: 5, Y: 5, W: 10, H: 10}
	assert.Equal(a.Intersects(b), b.Intersects(a))
}

func TestCircleBoundary(t *testing.T) {
	assert := assert.New(t)
	b := Circle{Center: Point{X: 3, Y: -4}, Radius: 2}.Boundary()
	assert.Equal(Rect{X: 1, Y: -6, W: 4, H: 4}, b)
	// The boundary must actually contain the circle's extreme points.
	c := Circle{Center: Point{X: 0, Y: 0}, Radius: 1}
	bb := c.Boundary()
	for _, p := range []Point{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}} {
		assert.True(bb.ContainsPoint(p), "boundary should contain %v", p)
	}
}

func TestCurveCopyPreservesClosed(t *testing.T) {
	assert := assert.New(t)
	open := Curve{Points: []Point{{X: 0, Y: 0}, {X: 1, Y: 1}}, Closed: false}
	assert.False(open.Copy().Closed)
	closed := Curve{Points: []Point{{X: 0, Y: 0}, {X: 1, Y: 1}}, Closed: true}
	assert.True(closed.Copy().Closed)

	// The copy must be independent of the original.
	c := open.Copy()
	c.Points[0].X = 99
	assert.Equal(0.0, open.Points[0].X)
}

func TestCentroidIndependentOfWinding(t *testing.T) {
	assert := assert.New(t)
	ccw := Curve{Points: []Point{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 2, Y: 2}, {X: 2, Y: 0}}, Closed: true}
	cw := Curve{Points: []Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}, Closed: true}
	for _, c := range []Curve{ccw, cw} {
		got := c.Centroid()
		assert.InDelta(1.0, got.X, Smol)
		assert.InDelta(1.0, got.Y, Smol)
	}

	pccw := Polygon{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 2, Y: 2}, {X: 2, Y: 0}}
	pcw := Polygon{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}
	for _, p := range []Polygon{pccw, pcw} {
		got := p.Centroid()
		assert.InDelta(1.0, got.X, Smol)
		assert.InDelta(1.0, got.Y, Smol)
	}
}

func TestCurveDegenerateInputs(t *testing.T) {
	assert := assert.New(t)
	var empty Curve
	assert.NotPanics(func() {
		assert.Equal(0.0, empty.Length())
		assert.Equal(Point{}, empty.Centroid())
		assert.Equal(Point{}, empty.Lerp(0.5))
	})
	closedEmpty := Curve{Closed: true}
	assert.NotPanics(func() { _ = closedEmpty.Length() })

	// Out-of-range percentages clamp instead of killing the process.
	c := Curve{Points: []Point{{X: 0, Y: 0}, {X: 10, Y: 0}}}
	assert.NotPanics(func() {
		assert.Equal(Point{X: 0, Y: 0}, c.Lerp(-1))
		assert.Equal(Point{X: 10, Y: 0}, c.Lerp(2))
		_, _ = c.LineAt(-1)
		_, _ = c.LineAt(2)
	})
	assert.InDelta(5.0, c.Lerp(0.5).X, Smol)
}

func TestPolygonDegenerateInputs(t *testing.T) {
	assert := assert.New(t)
	var empty Polygon
	assert.NotPanics(func() {
		assert.Equal(0.0, empty.Perimeter())
		assert.Equal(0.0, empty.Area())
		assert.Equal(Point{}, empty.Centroid())
	})
}

func TestLineAngleUsesAtan2(t *testing.T) {
	assert := assert.New(t)
	// A vertical line must not divide by zero, and direction must be kept.
	up := Line{P: Point{X: 0, Y: 0}, Q: Point{X: 0, Y: 1}}
	assert.InDelta(Pi/2, up.Angle(), Smol)
	down := Line{P: Point{X: 0, Y: 0}, Q: Point{X: 0, Y: -1}}
	assert.InDelta(-Pi/2, down.Angle(), Smol)
	// Opposite directions are distinguishable; Atan(dy/dx) collapsed them.
	right := Line{P: Point{X: 0, Y: 0}, Q: Point{X: 1, Y: 0}}
	left := Line{P: Point{X: 0, Y: 0}, Q: Point{X: -1, Y: 0}}
	assert.InDelta(0.0, right.Angle(), Smol)
	assert.InDelta(Pi, left.Angle(), Smol)
}

func TestPointHeapOrdering(t *testing.T) {
	assert := assert.New(t)
	// Seven elements in this order break a heap whose Pop re-slices off the
	// front; five do not, which is why the original test passed.
	metrics := []float64{10, 9, 8, 1, 2, 3, 4}
	maxHeap := NewMaxPointHeap()
	minHeap := NewMinPointHeap()
	for i, m := range metrics {
		p := MetricPoint{Metric: m, Index: i}
		maxHeap.Push(p)
		minHeap.Push(p)
	}
	ascending := append([]float64(nil), metrics...)
	sort.Float64s(ascending)
	n := len(metrics)
	for i := range ascending {
		assert.Equal(ascending[i], minHeap.Pop().Metric, "min-heap position %d", i)
		assert.Equal(ascending[n-1-i], maxHeap.Pop().Metric, "max-heap position %d", i)
	}
	assert.Equal(0, minHeap.Len())
	assert.Equal(0, maxHeap.Len())
}

func TestPointHeapRandomizedAgainstSort(t *testing.T) {
	assert := assert.New(t)
	rng := NewLFSRLargeWithSeed(12345)
	for trial := 0; trial < 50; trial++ {
		count := int(rng.Uint64n(40)) + 1
		var metrics []float64
		h := NewMinPointHeap()
		for i := 0; i < count; i++ {
			m := rng.Float64() * 100
			metrics = append(metrics, m)
			h.Push(MetricPoint{Metric: m, Index: i})
		}
		sort.Float64s(metrics)
		got := h.Report()
		assert.Len(got, count)
		for i := range metrics {
			assert.Equal(metrics[i], got[i].Metric, "trial %d position %d", trial, i)
		}
	}
}

func TestQuadTreeQueryCircleCoversWholeCircle(t *testing.T) {
	assert := assert.New(t)
	q := NewQuadTree(Rect{X: 0, Y: 0, W: 100, H: 100})
	center := Point{X: 50, Y: 50}
	// A ring of points well inside the query radius, one per direction.
	const ring = 16
	for i := 0; i < ring; i++ {
		angle := Tau * float64(i) / ring
		p := Point{X: center.X + 5*math.Cos(angle), Y: center.Y + 5*math.Sin(angle)}
		q.Insert(p.ToIndexPoint(i))
	}
	// A point outside the radius must be excluded.
	q.Insert(Point{X: 90, Y: 90}.ToIndexPoint(ring))

	assert.Len(q.QueryCircle(center, 10), ring)
	assert.Len(q.QueryCircleExcludeIndex(center, 10, 0), ring-1)
}

func TestKDTreeSizeEmpty(t *testing.T) {
	assert := assert.New(t)
	k := NewKDTree(Rect{X: 0, Y: 0, W: 10, H: 10})
	assert.Equal(0, k.Size())
	k.Insert(Point{X: 1, Y: 1}.ToIndexPoint(0))
	assert.Equal(1, k.Size())
	k.Insert(Point{X: 2, Y: 2}.ToIndexPoint(1))
	assert.Equal(2, k.Size())
}

// nearestBruteForce is the reference answer for the tree neighbor queries.
func nearestBruteForce(pts []Point, target IndexPoint, k int) []float64 {
	var d []float64
	for i, p := range pts {
		if i == target.Index {
			continue
		}
		d = append(d, SquaredDistance(target.Point, p))
	}
	sort.Float64s(d)
	if len(d) > k {
		d = d[:k]
	}
	return d
}

func TestNearestNeighborsMatchBruteForce(t *testing.T) {
	assert := assert.New(t)
	bounds := Rect{X: 0, Y: 0, W: 100, H: 100}
	rng := NewLFSRLargeWithSeed(99)
	var pts []Point
	kd := NewKDTree(bounds)
	qt := NewQuadTree(bounds)
	for i := 0; i < 200; i++ {
		p := Point{X: rng.Float64() * 100, Y: rng.Float64() * 100}
		pts = append(pts, p)
		kd.Insert(p.ToIndexPoint(i))
		qt.Insert(p.ToIndexPoint(i))
	}
	for _, idx := range []int{0, 7, 42, 150, 199} {
		target := pts[idx].ToIndexPoint(idx)
		for _, k := range []int{1, 5, 12} {
			want := nearestBruteForce(pts, target, k)

			kdGot := kd.NearestNeighbors(target, k)
			assert.Len(kdGot, k)
			for i, n := range kdGot {
				assert.InDelta(want[i], SquaredDistance(target.Point, n.Point), Smol,
					"kdtree idx=%d k=%d position=%d", idx, k, i)
			}

			qtGot := qt.NearestNeighbors(target, k)
			assert.Len(qtGot, k)
			for i, n := range qtGot {
				assert.InDelta(want[i], SquaredDistance(target.Point, n.Point), Smol,
					"quadtree idx=%d k=%d position=%d", idx, k, i)
			}
		}
	}
}

func TestAffineSetRotation(t *testing.T) {
	assert := assert.New(t)
	a := NewAffine2D()
	a.SetRotation(Pi / 2)
	got := a.MultVec(Vec2{X: 1, Y: 0})
	assert.InDelta(0.0, got.X, Smol)
	assert.InDelta(1.0, got.Y, Smol)

	// SetRotation must agree with the constructor.
	b := NewAffine2DWithRotation(0.42)
	c := NewAffine2D()
	c.SetRotation(0.42)
	v := Vec2{X: 2, Y: -3}
	assert.InDelta(b.MultVec(v).X, c.MultVec(v).X, Smol)
	assert.InDelta(b.MultVec(v).Y, c.MultVec(v).Y, Smol)
}

func TestAffineSetShearFactor(t *testing.T) {
	assert := assert.New(t)
	a := NewAffine2D()
	a.SetShearFactor(0.5, 0.25)
	// x' = x + 0.5y, y' = 0.25x + y
	got := a.MultVec(Vec2{X: 2, Y: 4})
	assert.InDelta(4.0, got.X, Smol)
	assert.InDelta(4.5, got.Y, Smol)
}

func TestRotateAboutPoint(t *testing.T) {
	assert := assert.New(t)
	// Rotating a point about itself leaves it in place.
	p := Point{X: 5, Y: 5}
	got := RotatePointAboutPoint(p, 1.234, p)
	assert.InDelta(p.X, got.X, Smol)
	assert.InDelta(p.Y, got.Y, Smol)
	// Quarter turn about the origin.
	got = RotatePointAboutPoint(Point{X: 1, Y: 0}, Pi/2, Point{X: 0, Y: 0})
	assert.InDelta(0.0, got.X, Smol)
	assert.InDelta(1.0, got.Y, Smol)
}

func TestRngIsDeterministic(t *testing.T) {
	assert := assert.New(t)
	rect := Rect{X: 0, Y: 0, W: 1, H: 1}
	r1 := NewRng(42)
	r2 := NewRng(42)
	assert.Equal(r1.UniformRandomPoints(20, rect), r2.UniformRandomPoints(20, rect))

	g1 := NewRng(7)
	g2 := NewRng(7)
	for i := 0; i < 10; i++ {
		assert.Equal(g1.Gaussian(0, 1), g2.Gaussian(0, 1))
	}

	// A different seed gives a different stream.
	d1, d2 := NewRng(1), NewRng(2)
	assert.NotEqual(d1.UniformRandomPoints(5, rect), d2.UniformRandomPoints(5, rect))

	// SetSeed(s) matches a fresh NewRng(s).
	r3 := NewRng(1)
	r3.UniformRandomPoints(3, rect)
	r3.SetSeed(42)
	fresh := NewRng(42)
	assert.Equal(fresh.UniformRandomPoints(20, rect), r3.UniformRandomPoints(20, rect))
}

func TestRngPointsStayInRect(t *testing.T) {
	assert := assert.New(t)
	rect := Rect{X: -5, Y: 10, W: 3, H: 7}
	r := NewRng(3)
	for _, p := range r.UniformRandomPoints(500, rect) {
		assert.True(rect.ContainsPoint(p), "point %v outside %v", p, rect)
	}
}

func TestLFSRZeroSeed(t *testing.T) {
	assert := assert.New(t)
	// A zero seed must not lock the register at zero.
	l := NewLFSRLargeWithSeed(0)
	assert.NotEqual(uint64(0), l.Next())
	m := NewLFSRMediumWithSeed(0)
	assert.NotEqual(uint32(0), m.Next())
	s := NewLFSRSmallWithSeed(0)
	assert.NotEqual(uint16(0), s.Next())

	// Float64 stays in [0, 1).
	big := NewLFSRLargeWithSeed(1)
	for i := 0; i < 10000; i++ {
		f := big.Float64()
		assert.GreaterOrEqual(f, 0.0)
		assert.Less(f, 1.0)
	}
}

func TestNoiseRespectsPersistence(t *testing.T) {
	assert := assert.New(t)
	// With the amplitude factor applied, changing persistence must change the
	// result; previously every octave was weighted equally and it did not.
	a := NewRng(11)
	a.SetNoiseOctaves(4)
	a.SetNoisePersistence(0.2)
	b := NewRng(11)
	b.SetNoiseOctaves(4)
	b.SetNoisePersistence(0.9)
	assert.NotEqual(a.Noise2D(12, 34), b.Noise2D(12, 34))

	// Zero octaves must not produce NaN.
	z := NewRng(1)
	z.SetNoiseOctaves(0)
	assert.Equal(0.0, z.Noise2D(1, 2))
}

func TestNoise4DUsesWAxis(t *testing.T) {
	assert := assert.New(t)
	// wscale defaulted to zero, which pinned the w term at 0.
	r := NewRng(5)
	assert.NotEqual(r.Noise4D(1, 1, 1, 0), r.Noise4D(1, 1, 1, 5000))
}

func TestTrigLUTInterpolates(t *testing.T) {
	assert := assert.New(t)
	lut := NewTrigLUT()
	// A raw table lookup is accurate to about one table step (9.6e-5). Real
	// interpolation is orders of magnitude better; require well past the
	// lookup-only error to prove the lerp is live.
	var worst float64
	for _, a := range Linspace(-Tau, 2*Tau, 50000, false) {
		d := math.Abs(math.Sin(a) - lut.Sin(a))
		if d > worst {
			worst = d
		}
	}
	assert.Less(worst, 1e-7, "worst sine error %g suggests interpolation is not applied", worst)

	var worstCos float64
	for _, a := range Linspace(-Tau, 2*Tau, 50000, false) {
		d := math.Abs(math.Cos(a) - lut.Cos(a))
		if d > worstCos {
			worstCos = d
		}
	}
	assert.Less(worstCos, 1e-7, "worst cosine error %g", worstCos)
}

func TestParseColor(t *testing.T) {
	assert := assert.New(t)
	c, err := ParseColor("#ff0000")
	assert.NoError(err)
	r, g, b, _ := c.RGBA()
	assert.Equal(uint32(0xffff), r)
	assert.Equal(uint32(0), g)
	assert.Equal(uint32(0), b)

	_, err = ParseColor("nonsense")
	assert.ErrorIs(err, ErrInvalidColor)

	// The hex pattern is anchored: a string merely containing a hex code is
	// not a hex color, and must not be silently accepted.
	_, err = ParseColor("junk#aabbcc")
	assert.ErrorIs(err, ErrInvalidColor)

	// Named colors still work, case insensitively.
	_, err = ParseColor("SteelBlue")
	assert.NoError(err)

	empty, err := ParseColor("")
	assert.NoError(err)
	_, _, _, alpha := empty.RGBA()
	assert.Equal(uint32(0), alpha)
}

func TestGradientDegenerateInputs(t *testing.T) {
	assert := assert.New(t)
	var empty Gradient
	assert.NotPanics(func() { _ = empty.Color(0.5) })

	// Fully transparent stops cannot be converted by colorful.MakeColor;
	// LinearPaletteStrings used to panic on them.
	transparent := NewGradientFromColors([]color.Color{color.Transparent, color.Transparent})
	assert.NotPanics(func() {
		got := transparent.LinearPaletteStrings(3)
		assert.Len(got, 3)
	})
}
