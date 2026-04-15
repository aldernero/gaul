package gaul

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelaunayTriangles_empty(t *testing.T) {
	tr := DelaunayTriangles(nil)
	assert.Nil(t, tr)
}

func TestDelaunayTriangles_fewPoints(t *testing.T) {
	assert.Nil(t, DelaunayTriangles([]Point{{0.5, 0.5}}))
	assert.Nil(t, DelaunayTriangles([]Point{{0.2, 0.2}, {0.8, 0.8}}))
}

func TestDelaunayTriangles_square_twoTriangles(t *testing.T) {
	sites := []Point{{0.1, 0.1}, {0.9, 0.1}, {0.9, 0.9}, {0.1, 0.9}}
	tr := DelaunayTriangles(sites)
	require.Len(t, tr, 2)
}

func TestDelaunayTriangles_equilateral(t *testing.T) {
	sites := []Point{{5, 1}, {2, 6}, {8, 6}}
	tr := DelaunayTriangles(sites)
	require.Len(t, tr, 1)
	assert.True(t, tr[0].Area() > 0)
}

func TestDelaunayTriangles_emptyCircumcircleProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	var sites []Point
	for i := 0; i < 30; i++ {
		sites = append(sites, Point{X: 0.05 + 0.9*rng.Float64(), Y: 0.05 + 0.9*rng.Float64()})
	}
	tr := DelaunayTriangles(sites)
	require.NotEmpty(t, tr)

	for _, tri := range tr {
		assert.True(t, orient2(tri.A, tri.B, tri.C) >= -Smol, "triangle should be CCW or flat")
	}

	uniq := dedupeSitesDelaunay(sites)
	for _, tri := range tr {
		for _, p := range uniq {
			if pointsEqualTri(tri.A, p) || pointsEqualTri(tri.B, p) || pointsEqualTri(tri.C, p) {
				continue
			}
			assert.False(t, inCircumcircle(tri.A, tri.B, tri.C, p),
				"site %v should not lie inside circumcircle of %v %v %v", p, tri.A, tri.B, tri.C)
		}
	}
}

func pointsEqualTri(a, b Point) bool {
	return Equalf(a.X, b.X) && Equalf(a.Y, b.Y)
}

func TestDelaunayTriangles_duplicatesMerged(t *testing.T) {
	withDups := []Point{
		{0.5, 0.5}, {0.5, 0.5}, {0.5, 0.5},
		{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9},
	}
	uniqueOnly := []Point{{0.5, 0.5}, {0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}}
	tr := DelaunayTriangles(withDups)
	tr2 := DelaunayTriangles(uniqueOnly)
	assert.Equal(t, len(tr2), len(tr))
	var sum, sum2 float64
	for _, x := range tr {
		sum += x.Area()
	}
	for _, x := range tr2 {
		sum2 += x.Area()
	}
	assert.InDelta(t, sum2, sum, 1e-9)
}

// BenchmarkDelaunayTriangles measures [DelaunayTriangles] with deterministic
// pseudo-random sites in a fixed square window per sub-benchmark size.
//
// Example: go test -bench=BenchmarkDelaunayTriangles -run=^$ -benchmem -count=5
func BenchmarkDelaunayTriangles(b *testing.B) {
	// Bowyer–Watson is roughly O(n²); keep large sizes out of default bench runs.
	sizes := []int{10, 50, 100, 500, 1000, 2500}
	for _, n := range sizes {
		rng := rand.New(rand.NewSource(42))
		sites := make([]Point, n)
		for i := 0; i < n; i++ {
			sites[i] = Point{X: rng.Float64() * 1000, Y: rng.Float64() * 1000}
		}
		b.Run(fmt.Sprintf("sites_%d", n), func(b *testing.B) {
			b.ReportAllocs()
			var tr []Triangle
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tr = DelaunayTriangles(sites)
			}
			_ = tr
		})
	}
}
