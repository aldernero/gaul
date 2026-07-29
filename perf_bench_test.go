package gaul

import (
	"math/rand"
	"testing"
)

// benchPoints builds a deterministic point cloud independent of gaul's own
// PRNG so the same benchmark source runs against any revision.
func benchPoints(n int) []Point {
	rng := rand.New(rand.NewSource(7))
	pts := make([]Point, n)
	for i := range pts {
		pts[i] = Point{X: rng.Float64() * 1000, Y: rng.Float64() * 1000}
	}
	return pts
}

func BenchmarkKDTreeNearestNeighbors(b *testing.B) {
	pts := benchPoints(5000)
	tree := NewKDTree(Rect{X: 0, Y: 0, W: 1000, H: 1000})
	for i, p := range pts {
		tree.Insert(p.ToIndexPoint(i))
	}
	targets := pts[:200]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := targets[i%len(targets)]
		_ = tree.NearestNeighbors(t.ToIndexPoint(i%len(targets)), 8)
	}
}

func BenchmarkQuadTreeNearestNeighbors(b *testing.B) {
	pts := benchPoints(5000)
	tree := NewQuadTree(Rect{X: 0, Y: 0, W: 1000, H: 1000})
	for i, p := range pts {
		tree.Insert(p.ToIndexPoint(i))
	}
	targets := pts[:200]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := targets[i%len(targets)]
		_ = tree.NearestNeighbors(t.ToIndexPoint(i%len(targets)), 8)
	}
}

func BenchmarkKDTreeSearch(b *testing.B) {
	pts := benchPoints(5000)
	tree := NewKDTree(Rect{X: 0, Y: 0, W: 1000, H: 1000})
	for i, p := range pts {
		tree.Insert(p.ToIndexPoint(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := pts[i%len(pts)]
		_ = tree.Search(p.ToIndexPoint(i % len(pts)))
	}
}

func BenchmarkLinspace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Linspace(0, Tau, 4096, false)
	}
}

func BenchmarkNewTrigLUT(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTrigLUT()
	}
}

func BenchmarkDelaunayTriangles400(b *testing.B) {
	pts := benchPoints(400)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DelaunayTriangles(pts)
	}
}

func BenchmarkShuffle(b *testing.B) {
	pts := benchPoints(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Shuffle(&pts)
	}
}

func BenchmarkRegularPolygonEdges(b *testing.B) {
	p, err := NewRegularPolygon(64, 100, Point{X: 0, Y: 0})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Edges()
	}
}
