package gaul

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aldernero/gaul/render"
)

func unitSquare() Rect {
	return Rect{X: 0, Y: 0, W: 1, H: 1}
}

func TestVoronoiWithRect_empty(t *testing.T) {
	curves, err := VoronoiWithRect(unitSquare(), nil)
	require.NoError(t, err)
	assert.Nil(t, curves)
}

func TestVoronoiWithRect_badBounds(t *testing.T) {
	_, err := VoronoiWithRect(Rect{X: 0, Y: 0, W: -1, H: 1}, []Point{{0.5, 0.5}})
	require.Error(t, err)
	_, err = VoronoiWithRect(Rect{X: 0, Y: 0, W: 1, H: 0}, []Point{{0.5, 0.5}})
	require.Error(t, err)
}

func TestVoronoiWithRect_siteOutside(t *testing.T) {
	_, err := VoronoiWithRect(unitSquare(), []Point{{1.5, 0.5}})
	require.Error(t, err)
}

func TestVoronoiWithRect_singleSite(t *testing.T) {
	b := unitSquare()
	curves, err := VoronoiWithRect(b, []Point{{0.5, 0.5}})
	require.NoError(t, err)
	require.Len(t, curves, 1)
	assert.True(t, curves[0].Closed)
	assert.InDelta(t, b.W*b.H, curves[0].Area(), 1e-6)
	tc := b.ToCurve()
	assert.InDelta(t, b.W*b.H, tc.Area(), 1e-6)
}

func TestVoronoiWithRect_twoSites_areaSum(t *testing.T) {
	b := unitSquare()
	curves, err := VoronoiWithRect(b, []Point{{0.25, 0.5}, {0.75, 0.5}})
	require.NoError(t, err)
	require.Len(t, curves, 2)
	sum := curves[0].Area() + curves[1].Area()
	assert.InDelta(t, b.W*b.H, sum, 1e-5)
}

// Voronoi cells are closed polygons (not triangles). Each should be wound CCW for +Y up.
func TestVoronoiWithRect_cellPolygonsAreCCW(t *testing.T) {
	b := unitSquare()
	cases := [][]Point{
		{{0.25, 0.5}, {0.75, 0.5}},
		{{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}},
	}
	for _, sites := range cases {
		curves, err := VoronoiWithRect(b, sites)
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

func TestVoronoiWithRect_verticesInsideBounds(t *testing.T) {
	b := Rect{X: 2, Y: 3, W: 10, H: 7}
	sites := []Point{
		{3, 4}, {8, 5}, {7, 8}, {4, 7},
	}
	curves, err := VoronoiWithRect(b, sites)
	require.NoError(t, err)
	require.Len(t, curves, len(sites))
	for _, c := range curves {
		for _, p := range c.Points {
			assert.True(t, b.ContainsPoint(p), "vertex %v outside bounds", p)
		}
	}
}

func TestVoronoiWithRect_grid_areaSum(t *testing.T) {
	b := unitSquare()
	var sites []Point
	for x := 0.25; x < 1; x += 0.5 {
		for y := 0.25; y < 1; y += 0.5 {
			sites = append(sites, Point{X: x, Y: y})
		}
	}
	curves, err := VoronoiWithRect(b, sites)
	require.NoError(t, err)
	var sum float64
	for _, c := range curves {
		sum += c.Area()
	}
	assert.InDelta(t, 1.0, sum, 1e-4)
}

func TestVoronoiWithRect_duplicates(t *testing.T) {
	b := unitSquare()
	p := Point{X: 0.5, Y: 0.5}
	curves, err := VoronoiWithRect(b, []Point{p, p})
	require.NoError(t, err)
	require.Len(t, curves, 2)
	assert.InDelta(t, curves[0].Area(), curves[1].Area(), 1e-9)
	require.Equal(t, len(curves[0].Points), len(curves[1].Points))
	for i := range curves[0].Points {
		assert.True(t, voronoiPointEqual(curves[0].Points[i], curves[1].Points[i]))
	}
}

func TestVoronoiWithRect_twoSites_nearestSite(t *testing.T) {
	b := unitSquare()
	a := Point{X: 0.2, Y: 0.5}
	c := Point{X: 0.8, Y: 0.5}
	curves, err := VoronoiWithRect(b, []Point{a, c})
	require.NoError(t, err)
	probe := Point{X: 0.4, Y: 0.5}
	assert.True(t, pointInPolygonOrOnEdge(probe, curves[0]))
	assert.False(t, pointInPolygonOrOnEdge(probe, curves[1]))
	probe2 := Point{X: 0.6, Y: 0.5}
	assert.False(t, pointInPolygonOrOnEdge(probe2, curves[0]))
	assert.True(t, pointInPolygonOrOnEdge(probe2, curves[1]))
}

func triangleCCW(a, b, c Point) Curve {
	return Curve{
		Closed: true,
		Points: []Point{a, b, c},
	}
}

func TestVoronoiWithCurve_empty(t *testing.T) {
	curves, err := VoronoiWithCurve(unitSquare().ToCurve(), nil)
	require.NoError(t, err)
	assert.Nil(t, curves)
}

func TestVoronoiWithCurve_notClosed(t *testing.T) {
	c := Curve{Closed: false, Points: []Point{{0, 0}, {1, 0}, {0, 1}}}
	_, err := VoronoiWithCurve(c, []Point{{0.1, 0.1}})
	require.Error(t, err)
}

func TestVoronoiWithCurve_siteOutsideTriangle(t *testing.T) {
	boundary := triangleCCW(Point{X: 0, Y: 0}, Point{X: 3, Y: 0}, Point{X: 1.5, Y: 2.6})
	_, err := VoronoiWithCurve(boundary, []Point{{5, 5}})
	require.Error(t, err)
}

func TestVoronoiWithCurve_nonConvexRejected(t *testing.T) {
	boundary := Curve{
		Closed: true,
		Points: []Point{
			{0, 0}, {4, 0}, {2, 2}, {4, 4}, {0, 4},
		},
	}
	_, err := VoronoiWithCurve(boundary, []Point{{2, 3.5}})
	require.Error(t, err)
}

func TestVoronoiWithCurve_agreesWithRectOnSquare(t *testing.T) {
	b := unitSquare()
	sites := []Point{{0.25, 0.5}, {0.75, 0.5}}
	rectCurves, err := VoronoiWithRect(b, sites)
	require.NoError(t, err)
	curveBoundary := b.ToCurve()
	curCurves, err := VoronoiWithCurve(curveBoundary, sites)
	require.NoError(t, err)
	require.Len(t, rectCurves, len(curCurves))
	for i := range rectCurves {
		assert.InDelta(t, rectCurves[i].Area(), curCurves[i].Area(), 1e-5, "cell %d area mismatch", i)
	}
}

func TestVoronoiWithCurve_triangleTwoSitesAreaSum(t *testing.T) {
	boundary := triangleCCW(Point{X: 0, Y: 0}, Point{X: 4, Y: 0}, Point{X: 2, Y: 3})
	sites := []Point{{1, 0.3}, {3, 0.3}}
	curves, err := VoronoiWithCurve(boundary, sites)
	require.NoError(t, err)
	require.Len(t, curves, 2)
	want := boundary.Area()
	sum := curves[0].Area() + curves[1].Area()
	assert.InDelta(t, want, sum, 1e-4)
}

func TestVoronoiWithCurve_nestedInParentCell(t *testing.T) {
	outer := unitSquare()
	parentSites := []Point{{0.15, 0.5}, {0.85, 0.5}}
	parentCells, err := VoronoiWithRect(outer, parentSites)
	require.NoError(t, err)
	boundary := parentCells[0]
	require.True(t, boundary.Closed)
	require.GreaterOrEqual(t, len(boundary.Points), 3)
	innerSites := []Point{{0.06, 0.5}, {0.12, 0.5}}
	curves, err := VoronoiWithCurve(boundary, innerSites)
	require.NoError(t, err)
	require.Len(t, curves, 2)
	var sum float64
	for _, c := range curves {
		sum += c.Area()
	}
	assert.InDelta(t, boundary.Area(), sum, 1e-3)
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

// voronoiRasterSize is the raster size in pixels for headless draw tests.
const voronoiRasterSize = 200

// voronoiDrawStyle configures how Voronoi cells are drawn on a render context.
type voronoiDrawStyle int

const (
	voronoiDrawFill voronoiDrawStyle = iota
	voronoiDrawStroke
	voronoiDrawFillAndStroke
)

// renderVoronoi builds a Voronoi diagram, draws every cell through a render
// context onto an in-memory raster, and checks the result is non-trivial.
// This replaces the old tdewolff/canvas tests that stressed its Settle /
// path-boolean engine, which has no analog in the in-repo renderer.
func renderVoronoi(t *testing.T, b Rect, sites []Point, style voronoiDrawStyle) {
	t.Helper()
	curves, err := VoronoiWithRect(b, sites)
	require.NoError(t, err)

	ras := render.NewRaster(voronoiRasterSize, voronoiRasterSize)
	ctx := render.NewContext(ras)
	// Map the diagram bounds onto the raster.
	ctx.Scale(voronoiRasterSize/b.W, voronoiRasterSize/b.H)
	ctx.Translate(-b.X, -b.Y)
	ctx.SetFillColor(color.RGBA{200, 200, 220, 255})
	ctx.SetStrokeColor(nil)
	switch style {
	case voronoiDrawStroke:
		ctx.SetFillColor(nil)
		ctx.SetStrokeColor(color.RGBA{20, 20, 40, 255})
		ctx.SetStrokeWidth(1)
	case voronoiDrawFillAndStroke:
		ctx.SetStrokeColor(color.RGBA{20, 20, 40, 255})
		ctx.SetStrokeWidth(1)
	}

	for i := range curves {
		curves[i].Draw(ctx)
	}

	painted := 0
	img := ras.Image()
	for y := range voronoiRasterSize {
		for x := range voronoiRasterSize {
			if img.RGBAAt(x, y).A > 0 {
				painted++
			}
		}
	}
	assert.Greater(t, painted, voronoiRasterSize, "expected non-trivial pixel coverage")
}

// TestVoronoiRender_fill exercises fill-only rendering (default Curve.Draw).
func TestVoronoiRender_fill(t *testing.T) {
	cases := []struct {
		name  string
		sites []Point
	}{
		{"two_sites", []Point{{0.25, 0.5}, {0.75, 0.5}}},
		{"duplicate_sites", []Point{{0.5, 0.5}, {0.5, 0.5}}},
		{"collinear", []Point{{0.1, 0.5}, {0.5, 0.5}, {0.9, 0.5}}},
		{"triangle", []Point{{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}}},
		{"nearly_coincident", []Point{{0.5, 0.5}, {0.5 + 1e-9, 0.5 + 1e-9}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			renderVoronoi(t, unitSquare(), tc.sites, voronoiDrawFill)
		})
	}
}

func TestVoronoiRender_fill_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoi(t, unitSquare(), sites, voronoiDrawFill)
}

// TestVoronoiRender_stroke draws cell outlines with stroke enabled.
func TestVoronoiRender_stroke(t *testing.T) {
	cases := []struct {
		name  string
		sites []Point
	}{
		{"two_sites", []Point{{0.25, 0.5}, {0.75, 0.5}}},
		{"duplicate_sites", []Point{{0.5, 0.5}, {0.5, 0.5}}},
		{"collinear", []Point{{0.1, 0.5}, {0.5, 0.5}, {0.9, 0.5}}},
		{"nearly_coincident", []Point{{0.5, 0.5}, {0.5 + 1e-9, 0.5 + 1e-9}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			renderVoronoi(t, unitSquare(), tc.sites, voronoiDrawStroke)
		})
	}
}

func TestVoronoiRender_stroke_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoi(t, unitSquare(), sites, voronoiDrawStroke)
}

func TestVoronoiRender_fillAndStroke(t *testing.T) {
	renderVoronoi(t, unitSquare(), []Point{{0.2, 0.3}, {0.8, 0.3}, {0.5, 0.85}}, voronoiDrawFillAndStroke)
}

func TestVoronoiRender_stroke_randomSites(t *testing.T) {
	b := Rect{X: 0, Y: 0, W: 10, H: 10}
	rng := rand.New(rand.NewSource(7))
	sites := make([]Point, 80)
	for i := range sites {
		sites[i] = Point{
			X: b.X + rng.Float64()*b.W,
			Y: b.Y + rng.Float64()*b.H,
		}
	}
	renderVoronoi(t, b, sites, voronoiDrawStroke)
}

// TestVoronoiRender_svg draws a diagram to the SVG backend and checks the
// document is produced (path elements with real stroke attributes).
func TestVoronoiRender_svg(t *testing.T) {
	curves, err := VoronoiWithRect(unitSquare(), []Point{{0.25, 0.5}, {0.75, 0.5}})
	require.NoError(t, err)

	svg := render.NewSVG(voronoiRasterSize, voronoiRasterSize)
	ctx := render.NewContext(svg)
	ctx.Scale(voronoiRasterSize, voronoiRasterSize)
	ctx.SetFillColor(nil)
	ctx.SetStrokeColor(color.RGBA{20, 20, 40, 255})
	ctx.SetStrokeWidth(1)
	for i := range curves {
		curves[i].Draw(ctx)
	}
	var buf strings.Builder
	_, err = svg.WriteTo(&buf)
	require.NoError(t, err)
	doc := buf.String()
	assert.Contains(t, doc, "<path")
	assert.Contains(t, doc, "stroke-width")
}

// BenchmarkVoronoiWithRect measures [VoronoiWithRect] with a fixed rectangle and a
// deterministic pseudo-random site set per sub-benchmark size. Sites are
// generated once; the benchmark loop only runs the diagram construction.
//
// Example: go test -bench=BenchmarkVoronoiWithRect -run=^$ -benchmem -count=5
func BenchmarkVoronoiWithRect(b *testing.B) {
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
				_, err = VoronoiWithRect(bounds, sites)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
