package render

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenRect(t *testing.T) {
	polys, closed := RectPath(0, 0, 10, 10).Flatten()
	assert.Len(t, polys, 1)
	assert.Equal(t, []bool{true}, closed)
	assert.Equal(t, []float64{0, 0, 10, 0, 10, 10, 0, 10}, polys[0])
}

func TestFlattenCircle(t *testing.T) {
	const r = 100.0
	polys, closed := CirclePath(0, 0, r).Flatten()
	assert.Len(t, polys, 1)
	assert.True(t, closed[0])
	poly := polys[0]
	// Every flattened vertex lies on the (bezier-approximated) circle.
	for i := 0; i < len(poly); i += 2 {
		assert.InDelta(t, r, math.Hypot(poly[i], poly[i+1]), 0.05)
	}
	// Step count is bounded: control polygon of each quarter is ~2.2*r,
	// so at 2 px steps a r=100 circle should have on the order of
	// 4*110 points, and never fewer than 4 per segment.
	n := len(poly) / 2
	assert.Greater(t, n, 100)
	assert.Less(t, n, 1000)
}

func TestFlattenOpenVsClosed(t *testing.T) {
	p := &Path{}
	p.MoveTo(0, 0)
	p.LineTo(10, 0)
	p.MoveTo(20, 0)
	p.LineTo(30, 0)
	p.Close()
	polys, closed := p.Flatten()
	assert.Len(t, polys, 2)
	assert.Equal(t, []bool{false, true}, closed)
}

func TestFlattenContinuesAfterClose(t *testing.T) {
	// Drawing after Close continues from the subpath start point.
	p := &Path{}
	p.MoveTo(0, 0)
	p.LineTo(10, 0)
	p.Close()
	p.LineTo(0, 10)
	polys, closed := p.Flatten()
	assert.Len(t, polys, 2)
	assert.True(t, closed[0])
	assert.False(t, closed[1])
	assert.Equal(t, []float64{0, 0, 0, 10}, polys[1])
}

func TestDashSimple(t *testing.T) {
	polys := [][]float64{{0, 0, 100, 0}}
	out := dashPolylines(polys, []bool{false}, []float64{10, 10}, 0)
	assert.Len(t, out, 5)
	for i, run := range out {
		start := float64(i * 20)
		assert.InDelta(t, start, run[0], eps)
		assert.InDelta(t, start+10, run[len(run)-2], eps)
	}
}

func TestDashOffset(t *testing.T) {
	polys := [][]float64{{0, 0, 100, 0}}
	// Offset 5 starts mid-dash: first run is only 5 long.
	out := dashPolylines(polys, []bool{false}, []float64{10, 10}, 5)
	assert.Len(t, out, 6)
	assert.InDelta(t, 0.0, out[0][0], eps)
	assert.InDelta(t, 5.0, out[0][2], eps)
	// Negative offsets wrap.
	out2 := dashPolylines(polys, []bool{false}, []float64{10, 10}, 5-20)
	assert.Equal(t, out, out2)
}

func TestDashOddPattern(t *testing.T) {
	// SVG semantics: [10] repeats as [10 10].
	polys := [][]float64{{0, 0, 40, 0}}
	out := dashPolylines(polys, []bool{false}, []float64{10}, 0)
	assert.Len(t, out, 2)
}

func TestDashClosedLoop(t *testing.T) {
	// A 40-perimeter square with dash [5 5] gets exactly 4 runs.
	polys, closed := RectPath(0, 0, 10, 10).Flatten()
	out := dashPolylines(polys, closed, []float64{5, 5}, 0)
	assert.Len(t, out, 4)
}

func TestDashDegeneratePatternIsSolid(t *testing.T) {
	polys := [][]float64{{0, 0, 100, 0}}
	out := dashPolylines(polys, []bool{false}, []float64{0, 0}, 0)
	assert.Equal(t, polys, out)
}
