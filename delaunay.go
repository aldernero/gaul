package gaul

import (
	"math"
	"sort"
)

// delaunayPointKey identifies duplicate coordinates (exact float64 equality).
type delaunayPointKey struct {
	x, y float64
}

func dedupeSitesDelaunay(sites []Point) []Point {
	seen := make(map[delaunayPointKey]struct{}, len(sites))
	out := make([]Point, 0, len(sites))
	for _, p := range sites {
		k := delaunayPointKey{p.X, p.Y}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, p)
	}
	return out
}

// DelaunayTriangles returns the Delaunay triangulation of the unique sites (duplicate
// coordinates are merged). Each output [Triangle] is oriented counterclockwise. The
// result is nil if there are fewer than three unique sites, or if no triangles remain
// after removing the artificial super-triangle (e.g. all sites collinear).
//
// Implementation: Bowyer–Watson with a bounding super-triangle. The straightforward
// formulation scans all triangles for each insertion, which is O(n²) in the worst case.
// Divide-and-conquer or sweepline Delaunay constructions achieve O(n log n) time but
// are more intricate to implement and maintain. CCW winding is not a property of any
// particular asymptotic class: it is applied when building each triangle, and any
// correct Delaunay routine can enforce the same convention.
func DelaunayTriangles(sites []Point) []Triangle {
	if len(sites) == 0 {
		return nil
	}
	unique := dedupeSitesDelaunay(sites)
	if len(unique) < 3 {
		return nil
	}
	return bowyerWatson(unique)
}

type triInt struct {
	i, j, k int
}

func sort3(a, b, c int) (int, int, int) {
	x := []int{a, b, c}
	sort.Ints(x)
	return x[0], x[1], x[2]
}

type edgeInt struct {
	a, b int // a < b
}

func edgeIntKey(i, j int) edgeInt {
	if i < j {
		return edgeInt{i, j}
	}
	return edgeInt{j, i}
}

func orient2(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

// inCircumcircle reports whether p lies strictly inside the circumcircle of triangle ABC.
// ABC may be clockwise or counterclockwise; degenerate (collinear) ABC yields false.
func inCircumcircle(a, b, c, p Point) bool {
	o := orient2(a, b, c)
	if math.Abs(o) <= Smol {
		return false
	}
	val := (a.X*a.X+a.Y*a.Y)*orient2(b, c, p) +
		(b.X*b.X+b.Y*b.Y)*orient2(c, a, p) +
		(c.X*c.X+c.Y*c.Y)*orient2(a, b, p) -
		(p.X*p.X+p.Y*p.Y)*o
	if o > 0 {
		return val > Smol
	}
	return val < -Smol
}

func superTriangle(pts []Point) (Point, Point, Point) {
	minX, minY := pts[0].X, pts[0].Y
	maxX, maxY := pts[0].X, pts[0].Y
	for _, p := range pts[1:] {
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
	dx := maxX - minX
	dy := maxY - minY
	dmax := math.Max(dx, dy)
	if dmax < Smol {
		dmax = 1
	}
	mx := (minX + maxX) * 0.5
	my := (minY + maxY) * 0.5
	r := dmax * 4
	// Wide enclosing triangle (Bourke / Sloan style).
	return Point{X: mx - 2*r, Y: my - r},
		Point{X: mx, Y: my + 2*r},
		Point{X: mx + 2*r, Y: my - r}
}

func ccwOrder(i, j, k int, pts []Point) (int, int, int) {
	a, b, c := pts[i], pts[j], pts[k]
	if orient2(a, b, c) >= 0 {
		return i, j, k
	}
	return i, k, j
}

func bowyerWatson(unique []Point) []Triangle {
	s0, s1, s2 := superTriangle(unique)
	pts := make([]Point, 0, len(unique)+3)
	pts = append(pts, s0, s1, s2)
	pts = append(pts, unique...)

	tris := []triInt{{0, 1, 2}}
	n := len(pts)

	for k := 3; k < n; k++ {
		pk := pts[k]
		var bad []triInt
		for _, t := range tris {
			a, b, c := pts[t.i], pts[t.j], pts[t.k]
			if inCircumcircle(a, b, c, pk) {
				bad = append(bad, t)
			}
		}
		if len(bad) == 0 {
			continue
		}

		badSet := make(map[triInt]struct{}, len(bad))
		for _, t := range bad {
			ai, bi, ci := sort3(t.i, t.j, t.k)
			badSet[triInt{ai, bi, ci}] = struct{}{}
		}

		edgeCount := make(map[edgeInt]int)
		for _, t := range bad {
			e1 := edgeIntKey(t.i, t.j)
			e2 := edgeIntKey(t.j, t.k)
			e3 := edgeIntKey(t.k, t.i)
			edgeCount[e1]++
			edgeCount[e2]++
			edgeCount[e3]++
		}

		var newTris []triInt
		for _, t := range tris {
			ai, bi, ci := sort3(t.i, t.j, t.k)
			if _, ok := badSet[triInt{ai, bi, ci}]; !ok {
				newTris = append(newTris, t)
			}
		}
		tris = newTris

		for e, cnt := range edgeCount {
			if cnt != 1 {
				continue
			}
			i, j := e.a, e.b
			ii, jj, kk := ccwOrder(i, j, k, pts)
			tris = append(tris, triInt{ii, jj, kk})
		}
	}

	var out []Triangle
	for _, t := range tris {
		if t.i < 3 || t.j < 3 || t.k < 3 {
			continue
		}
		a, b, c := pts[t.i], pts[t.j], pts[t.k]
		if orient2(a, b, c) < -Smol {
			a, b = b, a
		}
		out = append(out, Triangle{A: a, B: b, C: c})
	}
	return out
}
