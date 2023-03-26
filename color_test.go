package gaul

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSimpleGradientFromNamed(t *testing.T) {
	sg := NewSimpleGradientFromNamed("red", "blue")
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
