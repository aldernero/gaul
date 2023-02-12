package gaul

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestTrigLUT_Sin(t *testing.T) {
	lut := NewTrigLUT()
	for _, a := range Linspace(0, Tau, 10000, false) {
		slow := math.Sin(a)
		fast := lut.Sin(a)
		diff := math.Abs(slow - fast)
		assert.Lessf(t, diff, 0.0001, "a=%f, diff=%f", a, diff)
	}
}

func TestTrigLUT_Cos(t *testing.T) {
	lut := NewTrigLUT()
	for _, a := range Linspace(0, Tau, 10000, false) {
		slow := math.Cos(a)
		fast := lut.Cos(a)
		diff := math.Abs(slow - fast)
		assert.Lessf(t, diff, 0.0001, "a=%f, diff=%f", a, diff)
	}
}

func BenchmarkTrigLUT_FastSinCos(b *testing.B) {
	lut := NewTrigLUT()
	a := Linspace(0, Tau, 1000, false)
	for i := 0; i < b.N; i++ {
		for _, x := range a {
			lut.Sin(x)
			lut.Cos(x)
		}
	}
}

func BenchmarkTrigLUT_SlowSinCos(b *testing.B) {
	a := Linspace(0, Tau, 1000, false)
	for i := 0; i < b.N; i++ {
		for _, x := range a {
			math.Sin(x)
			math.Cos(x)
		}
	}
}
