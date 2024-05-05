package gaul

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolygon_Centroid(t *testing.T) {
	origin := Point{X: 0, Y: 0}
	// square
	points := []Point{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 0, Y: 1},
	}
	var poly Polygon = points
	centroid := poly.Centroid()
	assert.LessOrEqual(t, Distance(centroid, Point{X: 0.5, Y: 0.5}), Smol)
	// square centered at origin
	points = []Point{
		{X: -1, Y: -1},
		{X: 1, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
	}
	poly = points
	centroid = poly.Centroid()
	assert.LessOrEqual(t, Distance(centroid, origin), Smol)
	// regular hexagon
	hexagon, _ := NewRegularPolygon(6, 1, Point{X: 0, Y: 0})
	poly = hexagon.Points()
	centroid = poly.Centroid()
	assert.LessOrEqual(t, Distance(centroid, origin), Smol)
}
