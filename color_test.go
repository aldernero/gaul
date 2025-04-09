package gaul

import (
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
