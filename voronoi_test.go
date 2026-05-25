package gaul

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
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

// voronoiCanvasMM is the canvas size in millimeters for headless draw tests.
const voronoiCanvasMM = 100.0

// curveToCanvasPath builds a canvas path from a gaul curve (line segments only).
func curveToCanvasPath(c Curve) *canvas.Path {
	p := &canvas.Path{}
	n := len(c.Points)
	if n == 0 {
		return p
	}
	p.MoveTo(c.Points[0].X, c.Points[0].Y)
	for i := 1; i < n; i++ {
		p.LineTo(c.Points[i].X, c.Points[i].Y)
	}
	if c.Closed {
		p.Close()
	}
	return p
}

// voronoiDrawStyle configures how Voronoi cells are drawn on a canvas context.
type voronoiDrawStyle int

const (
	voronoiDrawFill voronoiDrawStyle = iota
	voronoiDrawStroke
	voronoiDrawFillAndStroke
)

// renderVoronoiOnCanvas builds a Voronoi diagram, draws every cell on an in-memory
// canvas, and rasterizes it. No window is opened: canvas.New only allocates layers;
// intersection code runs during rasterizer.RenderPath when strokes are enabled, or
// when Path.Settle is called explicitly (see settleEach).
func renderVoronoiOnCanvas(t *testing.T, b Rect, sites []Point, style voronoiDrawStyle, settleEach bool) {
	t.Helper()
	curves, err := VoronoiWithRect(b, sites)
	require.NoError(t, err)

	c := canvas.New(voronoiCanvasMM, voronoiCanvasMM)
	ctx := canvas.NewContext(c)
	ctx.SetFillColor(color.RGBA{200, 200, 220, 255})
	switch style {
	case voronoiDrawStroke:
		ctx.SetFillColor(canvas.Transparent)
		ctx.SetStrokeColor(color.RGBA{20, 20, 40, 255})
		ctx.SetStrokeWidth(0.2)
	case voronoiDrawFillAndStroke:
		ctx.SetStrokeColor(color.RGBA{20, 20, 40, 255})
		ctx.SetStrokeWidth(0.15)
	}

	for _, cell := range curves {
		if settleEach {
			settled := curveToCanvasPath(cell).Settle(canvas.NonZero)
			ctx.DrawPath(0, 0, settled)
			ctx.Fill()
			continue
		}
		switch style {
		case voronoiDrawStroke:
			n := len(cell.Points)
			if n < 2 {
				continue
			}
			ctx.MoveTo(cell.Points[0].X, cell.Points[0].Y)
			for i := 1; i < n; i++ {
				ctx.LineTo(cell.Points[i].X, cell.Points[i].Y)
			}
			if cell.Closed {
				ctx.Close()
			}
			ctx.Stroke()
		default:
			cell.Draw(ctx)
		}
	}

	rasterizeCanvas(c)
}

// TestVoronoiCanvasRender_fill exercises fill-only rendering (default Curve.Draw).
// Bentley–Ottmann settling is not used for fill-only paths in the rasterizer.
func TestVoronoiCanvasRender_fill(t *testing.T) {
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
			renderVoronoiOnCanvas(t, unitSquare(), tc.sites, voronoiDrawFill, false)
		})
	}
}

func TestVoronoiCanvasRender_fill_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoiOnCanvas(t, unitSquare(), sites, voronoiDrawFill, false)
}

// TestVoronoiCanvasRender_stroke draws cell outlines with stroke enabled. Path.Stroke
// calls Settle unless canvas.FastStroke is true — this is the main intersection stress path.
func TestVoronoiCanvasRender_stroke(t *testing.T) {
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
			renderVoronoiOnCanvas(t, unitSquare(), tc.sites, voronoiDrawStroke, false)
		})
	}
}

func TestVoronoiCanvasRender_stroke_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoiOnCanvas(t, unitSquare(), sites, voronoiDrawStroke, false)
}

func TestVoronoiCanvasRender_fillAndStroke(t *testing.T) {
	renderVoronoiOnCanvas(t, unitSquare(), []Point{{0.2, 0.3}, {0.8, 0.3}, {0.5, 0.85}}, voronoiDrawFillAndStroke, false)
}

func TestVoronoiCanvasRender_stroke_randomSites(t *testing.T) {
	b := Rect{X: 0, Y: 0, W: 10, H: 10}
	rng := rand.New(rand.NewSource(7))
	sites := make([]Point, 80)
	for i := range sites {
		sites[i] = Point{
			X: b.X + rng.Float64()*b.W,
			Y: b.Y + rng.Float64()*b.H,
		}
	}
	renderVoronoiOnCanvas(t, b, sites, voronoiDrawStroke, false)
}

// TestVoronoiCanvasSettle calls Path.Settle on each cell directly (always runs Bentley–Ottmann).
func TestVoronoiCanvasSettle(t *testing.T) {
	cases := []struct {
		name  string
		sites []Point
	}{
		{"two_sites", []Point{{0.25, 0.5}, {0.75, 0.5}}},
		{"duplicate_sites", []Point{{0.5, 0.5}, {0.5, 0.5}}},
		{"collinear", []Point{{0.1, 0.5}, {0.5, 0.5}, {0.9, 0.5}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			renderVoronoiOnCanvas(t, unitSquare(), tc.sites, voronoiDrawFill, true)
		})
	}
}

func TestVoronoiCanvasSettle_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.1; x < 1.0; x += 0.2 {
		for y := 0.1; y < 1.0; y += 0.2 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoiOnCanvas(t, unitSquare(), sites, voronoiDrawFill, true)
}

// voronoiMergedPath joins all cell outlines into one canvas path (many subpaths, shared edges
// drawn twice). This maximizes overlap for Bentley–Ottmann compared to settling each cell alone.
func voronoiMergedPath(t *testing.T, b Rect, sites []Point) *canvas.Path {
	t.Helper()
	curves, err := VoronoiWithRect(b, sites)
	require.NoError(t, err)
	paths := make(canvas.Paths, len(curves))
	for i, cell := range curves {
		paths[i] = curveToCanvasPath(cell)
	}
	return paths.Merge()
}

// rasterizeCanvas renders an in-memory canvas headlessly (no window).
func rasterizeCanvas(c *canvas.Canvas) {
	ras := rasterizer.New(c.W, c.H, canvas.DPMM(2), canvas.DefaultColorSpace)
	c.RenderTo(ras)
	ras.Close()
}

// renderVoronoiMergedSettle merges every cell into one path and runs a single Settle pass.
func renderVoronoiMergedSettle(t *testing.T, b Rect, sites []Point) {
	t.Helper()
	merged := voronoiMergedPath(t, b, sites)
	settled := merged.Settle(canvas.NonZero)

	c := canvas.New(voronoiCanvasMM, voronoiCanvasMM)
	ctx := canvas.NewContext(c)
	ctx.SetFillColor(color.RGBA{180, 200, 220, 255})
	ctx.DrawPath(0, 0, settled)
	rasterizeCanvas(c)
}

// renderVoronoiMergedStroke strokes the merged path in one DrawPath call so shared cell edges
// are processed together through Path.Stroke → Settle (when FastStroke is false).
func renderVoronoiMergedStroke(t *testing.T, b Rect, sites []Point) {
	t.Helper()
	merged := voronoiMergedPath(t, b, sites)

	c := canvas.New(voronoiCanvasMM, voronoiCanvasMM)
	ctx := canvas.NewContext(c)
	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(color.RGBA{20, 20, 40, 255})
	ctx.SetStrokeWidth(0.2)
	ctx.DrawPath(0, 0, merged)
	rasterizeCanvas(c)
}

// TestVoronoiCanvasMergedSettle runs one Settle over all cells merged (heavy overlap).
func TestVoronoiCanvasMergedSettle(t *testing.T) {
	cases := []struct {
		name  string
		sites []Point
	}{
		{"two_sites", []Point{{0.25, 0.5}, {0.75, 0.5}}},
		{"duplicate_sites", []Point{{0.5, 0.5}, {0.5, 0.5}}},
		{"collinear", []Point{{0.1, 0.5}, {0.5, 0.5}, {0.9, 0.5}}},
		{"triangle", []Point{{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			renderVoronoiMergedSettle(t, unitSquare(), tc.sites)
		})
	}
}

func TestVoronoiCanvasMergedSettle_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoiMergedSettle(t, unitSquare(), sites)
}

// TestVoronoiCanvasMergedStroke strokes the merged path (shared edges in one stroke op).
func TestVoronoiCanvasMergedStroke(t *testing.T) {
	renderVoronoiMergedStroke(t, unitSquare(), []Point{{0.25, 0.5}, {0.75, 0.5}})
	renderVoronoiMergedStroke(t, unitSquare(), []Point{{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}})
}

// TestVoronoiCanvasMergedStroke_denseGrid strokes all cells as one path (shared edges in one
// Stroke → Settle pass). Regression for canvas bentleyOttmann polygon walking at tight junctions.
func TestVoronoiCanvasMergedStroke_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	renderVoronoiMergedStroke(t, unitSquare(), sites)
}

// TestVoronoiCanvasMergedStroke_denseGrid_fastStroke verifies canvas.FastStroke bypasses the
// Settle panic in TestVoronoiCanvasMergedStroke_denseGrid for the same site layout.
func TestVoronoiCanvasMergedStroke_denseGrid_fastStroke(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	orig := canvas.FastStroke
	canvas.FastStroke = true
	t.Cleanup(func() { canvas.FastStroke = orig })
	renderVoronoiMergedStroke(t, unitSquare(), sites)
}

// TestVoronoiCanvasStroke_fastStroke compares stroke rendering with canvas.FastStroke on and off.
// When true, Path.Stroke skips Settle (canvas bypass for intersection).
func TestVoronoiCanvasStroke_fastStroke(t *testing.T) {
	sites := []Point{{0.25, 0.5}, {0.75, 0.5}, {0.5, 0.85}}
	for _, fast := range []bool{false, true} {
		name := "settle_on_stroke"
		if fast {
			name = "fast_stroke_no_settle"
		}
		t.Run(name, func(t *testing.T) {
			orig := canvas.FastStroke
			canvas.FastStroke = fast
			t.Cleanup(func() { canvas.FastStroke = orig })
			renderVoronoiOnCanvas(t, unitSquare(), sites, voronoiDrawStroke, false)
		})
	}
}

func TestVoronoiCanvasStroke_fastStroke_denseGrid(t *testing.T) {
	var sites []Point
	for x := 0.05; x < 1.0; x += 0.1 {
		for y := 0.05; y < 1.0; y += 0.1 {
			sites = append(sites, Point{x, y})
		}
	}
	for _, fast := range []bool{false, true} {
		name := "settle_on_stroke"
		if fast {
			name = "fast_stroke_no_settle"
		}
		t.Run(name, func(t *testing.T) {
			orig := canvas.FastStroke
			canvas.FastStroke = fast
			t.Cleanup(func() { canvas.FastStroke = orig })
			renderVoronoiOnCanvas(t, unitSquare(), sites, voronoiDrawStroke, false)
		})
	}
}

func TestVoronoiCanvasMergedStroke_fastStroke(t *testing.T) {
	sites := []Point{{0.1, 0.1}, {0.9, 0.1}, {0.5, 0.9}}
	for _, fast := range []bool{false, true} {
		name := "settle_on_stroke"
		if fast {
			name = "fast_stroke_no_settle"
		}
		t.Run(name, func(t *testing.T) {
			orig := canvas.FastStroke
			canvas.FastStroke = fast
			t.Cleanup(func() { canvas.FastStroke = orig })
			renderVoronoiMergedStroke(t, unitSquare(), sites)
		})
	}
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
