package gaul

import (
	"image/color"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleGradient(t *testing.T) {
	sg := NewSimpleGradient("red", "blue")
	// red
	c := sg.Color(0.0)
	r, g, b, _ := c.RGBA()
	assert.Equal(t, uint8(255), uint8(r))
	assert.Equal(t, uint8(0), uint8(g))
	assert.Equal(t, uint8(0), uint8(b))
	// blue
	c = sg.Color(1.0)
	r, g, b, _ = c.RGBA()
	assert.Equal(t, uint8(0), uint8(r))
	assert.Equal(t, uint8(0), uint8(g))
	assert.Equal(t, uint8(255), uint8(b))
}

func TestNewGradientFromNamed(t *testing.T) {
	grad := NewGradientFromNamed([]string{"red", "blue"})
	// red
	c := grad.Color(0.0)
	r, g, b, _ := c.RGBA()
	assert.Equal(t, uint8(255), uint8(r))
	assert.Equal(t, uint8(0), uint8(g))
	assert.Equal(t, uint8(0), uint8(b))
	// blue
	c = grad.Color(1.0)
	r, g, b, _ = c.RGBA()
	assert.Equal(t, uint8(0), uint8(r))
	assert.Equal(t, uint8(0), uint8(g))
	assert.Equal(t, uint8(255), uint8(b))
}

func TestGradient_LinearPaletteStrings(t *testing.T) {
	grad := NewGradientFromNamed([]string{"red", "blue"})
	palette := grad.LinearPaletteStrings(2)
	assert.Equal(t, []string{"#ff0000", "#0000ff"}, palette)
	palette = grad.LinearPaletteStrings(3)
	assert.Equal(t, []string{"#ff0000", "#fb0080", "#0000ff"}, palette)
	grad = NewGradientFromNamed([]string{"red", "green", "blue"})
	palette = grad.LinearPaletteStrings(3)
	assert.Equal(t, []string{"#ff0000", "#008000", "#0000ff"}, palette)
}

func TestSinePalette_ToGridPng(t *testing.T) {
	a := Vec3{rand.Float64(), rand.Float64(), rand.Float64()}
	b := Vec3{rand.Float64(), rand.Float64(), rand.Float64()}
	c := Vec3{rand.Float64(), rand.Float64(), rand.Float64()}
	d := Vec3{rand.Float64(), rand.Float64(), rand.Float64()}
	pal := NewSinePalette(c, d)
	pal.A = a
	pal.B = b
	fname, err := pal.ToGridPng()
	assert.Nil(t, err)
	t.Log(fname)
}

func TestGradientPositions(t *testing.T) {
	red := StringToColor("red")
	green := StringToColor("green")
	blue := StringToColor("blue")
	// Green is placed close to the start (0.2) rather than the midpoint.
	g := NewGradientFromColorStops(
		[]color.Color{red, green, blue},
		[]float64{0, 0.2, 1},
	)

	// Exact stop hits.
	r, gr, b, _ := g.Color(0).RGBA()
	assert.Equal(t, uint8(255), uint8(r))
	assert.Equal(t, uint8(0), uint8(gr))
	assert.Equal(t, uint8(0), uint8(b))

	r, gr, b, _ = g.Color(0.2).RGBA()
	assert.Equal(t, uint8(0), uint8(r))
	assert.Equal(t, uint8(128), uint8(gr)) // "green" is 0,128,0
	assert.Equal(t, uint8(0), uint8(b))

	r, gr, b, _ = g.Color(1).RGBA()
	assert.Equal(t, uint8(0), uint8(r))
	assert.Equal(t, uint8(0), uint8(gr))
	assert.Equal(t, uint8(255), uint8(b))

	// A nil positions slice must reproduce the even-spacing behavior.
	even := NewGradientFromColorStops([]color.Color{red, green, blue}, nil)
	plain := NewGradientFromColors([]color.Color{red, green, blue})
	c1 := even.Color(0.5)
	c2 := plain.Color(0.5)
	assert.Equal(t, c1, c2)
}
