package render

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func pathDump(p *Path) (verbs []PathVerb, coords []float64) {
	p.Iterate(func(v PathVerb, pts []float64) {
		verbs = append(verbs, v)
		coords = append(coords, pts...)
	})
	return verbs, coords
}

func TestPathBuilding(t *testing.T) {
	p := &Path{}
	assert.True(t, p.Empty())
	p.MoveTo(1, 2)
	p.LineTo(3, 4)
	p.QuadTo(5, 6, 7, 8)
	p.CubicTo(9, 10, 11, 12, 13, 14)
	p.Close()
	assert.False(t, p.Empty())

	verbs, coords := pathDump(p)
	assert.Equal(t, []PathVerb{VerbMoveTo, VerbLineTo, VerbQuadTo, VerbCubicTo, VerbClose}, verbs)
	assert.Equal(t, []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, coords)
}

func TestPathLineToWithoutCurrent(t *testing.T) {
	p := &Path{}
	p.LineTo(5, 5)
	verbs, _ := pathDump(p)
	assert.Equal(t, []PathVerb{VerbMoveTo}, verbs)
}

func TestPathReset(t *testing.T) {
	p := &Path{}
	p.MoveTo(1, 1)
	p.LineTo(2, 2)
	p.Reset()
	assert.True(t, p.Empty())
	p.LineTo(3, 3) // behaves like MoveTo again after reset
	verbs, _ := pathDump(p)
	assert.Equal(t, []PathVerb{VerbMoveTo}, verbs)
}

func TestPathTransform(t *testing.T) {
	p := &Path{}
	p.MoveTo(0, 0)
	p.CubicTo(1, 0, 2, 0, 3, 0)
	q := p.Transform(Translation(10, 20))
	_, coords := pathDump(q)
	assert.Equal(t, []float64{10, 20, 11, 20, 12, 20, 13, 20}, coords)
	// Original untouched.
	_, coords = pathDump(p)
	assert.Equal(t, []float64{0, 0, 1, 0, 2, 0, 3, 0}, coords)
}

func TestPathBounds(t *testing.T) {
	p := RectPath(10, 20, 30, 40)
	minX, minY, maxX, maxY := p.Bounds()
	assert.Equal(t, 10.0, minX)
	assert.Equal(t, 20.0, minY)
	assert.Equal(t, 40.0, maxX)
	assert.Equal(t, 60.0, maxY)
}

func TestCirclePathAccuracy(t *testing.T) {
	const r = 100.0
	p := CirclePath(0, 0, r)
	// Sample each cubic segment densely; a 4-bezier circle should stay
	// within 0.03% of the true radius.
	var curX, curY float64
	p.Iterate(func(v PathVerb, pts []float64) {
		switch v {
		case VerbMoveTo:
			curX, curY = pts[0], pts[1]
		case VerbCubicTo:
			for i := 1; i <= 50; i++ {
				x, y := cubicPoint(curX, curY, pts[0], pts[1], pts[2], pts[3], pts[4], pts[5], float64(i)/50)
				assert.InDelta(t, r, math.Hypot(x, y), 0.05)
			}
			curX, curY = pts[4], pts[5]
		}
	})
}
