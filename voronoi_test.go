package gaul

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unitSquare() Rect {
	return Rect{X: 0, Y: 0, W: 1, H: 1}
}

func TestVoronoiCells_empty(t *testing.T) {
	curves, err := VoronoiCells(unitSquare(), nil)
	require.NoError(t, err)
	assert.Nil(t, curves)
}

func TestVoronoiCells_badBounds(t *testing.T) {
	_, err := VoronoiCells(Rect{X: 0, Y: 0, W: -1, H: 1}, []Point{{0.5, 0.5}})
	require.Error(t, err)
	_, err = VoronoiCells(Rect{X: 0, Y: 0, W: 1, H: 0}, []Point{{0.5, 0.5}})
	require.Error(t, err)
}

func TestVoronoiCells_siteOutside(t *testing.T) {
	_, err := VoronoiCells(unitSquare(), []Point{{1.5, 0.5}})
	require.Error(t, err)
}

func TestVoronoiCells_singleSite(t *testing.T) {
	b := unitSquare()
	curves, err := VoronoiCells(b, []Point{{0.5, 0.5}})
	require.NoError(t, err)
	require.Len(t, curves, 1)
	assert.True(t, curves[0].Closed)
	assert.InDelta(t, b.W*b.H, curves[0].Area(), 1e-6)
	tc := b.ToCurve()
	assert.InDelta(t, b.W*b.H, tc.Area(), 1e-6)
}

func TestVoronoiCells_twoSites_areaSum(t *testing.T) {
	b := unitSquare()
	curves, err := VoronoiCells(b, []Point{{0.25, 0.5}, {0.75, 0.5}})
	require.NoError(t, err)
	require.Len(t, curves, 2)
	sum := curves[0].Area() + curves[1].Area()
	assert.InDelta(t, b.W*b.H, sum, 1e-5)
}

// Voronoi cells are closed polygons (not triangles). Each should be wound CCW for +Y up.
func TestVoronoiCells_cellPolygonsAreCCW(t *testing.T) {
	b := unitSquare()
	cases := [][]Point{
		{{0.25, 0.5}, {0.75, 0.5}},
		{{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}},
	}
	for _, sites := range cases {
		curves, err := VoronoiCells(b, sites)
		require.NoError(t, err)
		for _, c := range curves {
			if len(c.Points) < 3 {
				continue
			}
			assert.GreaterOrEqual(t, voronoiPolygonSignedArea2(c.Points), -Smol,
				"cell polygon should be counterclockwise (non-negative signed shoelace sum)")
		}
	}
}

func TestVoronoiCells_verticesInsideBounds(t *testing.T) {
	b := Rect{X: 2, Y: 3, W: 10, H: 7}
	sites := []Point{
		{3, 4}, {8, 5}, {7, 8}, {4, 7},
	}
	curves, err := VoronoiCells(b, sites)
	require.NoError(t, err)
	require.Len(t, curves, len(sites))
	for _, c := range curves {
		for _, p := range c.Points {
			assert.True(t, b.ContainsPoint(p), "vertex %v outside bounds", p)
		}
	}
}

func TestVoronoiCells_grid_areaSum(t *testing.T) {
	b := unitSquare()
	var sites []Point
	for x := 0.25; x < 1; x += 0.5 {
		for y := 0.25; y < 1; y += 0.5 {
			sites = append(sites, Point{X: x, Y: y})
		}
	}
	curves, err := VoronoiCells(b, sites)
	require.NoError(t, err)
	var sum float64
	for _, c := range curves {
		sum += c.Area()
	}
	assert.InDelta(t, 1.0, sum, 1e-4)
}

func TestVoronoiCells_duplicates(t *testing.T) {
	b := unitSquare()
	p := Point{X: 0.5, Y: 0.5}
	curves, err := VoronoiCells(b, []Point{p, p})
	require.NoError(t, err)
	require.Len(t, curves, 2)
	assert.InDelta(t, curves[0].Area(), curves[1].Area(), 1e-9)
	require.Equal(t, len(curves[0].Points), len(curves[1].Points))
	for i := range curves[0].Points {
		assert.True(t, voronoiPointEqual(curves[0].Points[i], curves[1].Points[i]))
	}
}

func TestVoronoiCells_twoSites_nearestSite(t *testing.T) {
	b := unitSquare()
	a := Point{X: 0.2, Y: 0.5}
	c := Point{X: 0.8, Y: 0.5}
	curves, err := VoronoiCells(b, []Point{a, c})
	require.NoError(t, err)
	probe := Point{X: 0.4, Y: 0.5}
	assert.True(t, pointInPolygonOrOnEdge(probe, curves[0]))
	assert.False(t, pointInPolygonOrOnEdge(probe, curves[1]))
	probe2 := Point{X: 0.6, Y: 0.5}
	assert.False(t, pointInPolygonOrOnEdge(probe2, curves[0]))
	assert.True(t, pointInPolygonOrOnEdge(probe2, curves[1]))
}

// pointInPolygonOrOnEdge uses ray casting; boundary counts as inside for convex cells.
func pointInPolygonOrOnEdge(p Point, c Curve) bool {
	if !c.Closed || len(c.Points) < 3 {
		return false
	}
	n := len(c.Points)
	inside := false
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi := c.Points[i].X, c.Points[i].Y
		xj, yj := c.Points[j].X, c.Points[j].Y
		if ((yi > p.Y) != (yj > p.Y)) && (p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
	}
	if inside {
		return true
	}
	for i := 0; i < n; i++ {
		q := c.Points[(i+1)%n]
		if distancePointToSegment(p, c.Points[i], q) <= Smol*10 {
			return true
		}
	}
	return false
}

func distancePointToSegment(p, a, b Point) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return Distance(p, a)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	proj := Point{X: a.X + t*dx, Y: a.Y + t*dy}
	return Distance(p, proj)
}

// BenchmarkVoronoiCells measures [VoronoiCells] with a fixed rectangle and a
// deterministic pseudo-random site set per sub-benchmark size. Sites are
// generated once; the benchmark loop only runs the diagram construction.
//
// Example: go test -bench=BenchmarkVoronoiCells -run=^$ -benchmem -count=5
func BenchmarkVoronoiCells(b *testing.B) {
	bounds := Rect{X: 0, Y: 0, W: 1000, H: 1000}
	sizes := []int{10, 50, 100, 500, 1000, 2500, 5000, 10000}
	for _, n := range sizes {
		rng := rand.New(rand.NewSource(42))
		sites := make([]Point, n)
		for i := 0; i < n; i++ {
			sites[i] = Point{
				X: bounds.X + rng.Float64()*bounds.W,
				Y: bounds.Y + rng.Float64()*bounds.H,
			}
		}
		b.Run(fmt.Sprintf("sites_%d", n), func(b *testing.B) {
			b.ReportAllocs()
			var err error
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err = VoronoiCells(bounds, sites)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
