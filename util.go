package gaul

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	Pi    = math.Pi
	Tau   = 2 * math.Pi
	Sqrt2 = math.Sqrt2
	Sqrt3 = 1.7320508075688772
	// Smol is a small value used for floating point comparisons
	Smol = 1e-5 // Increased from 1e-6 to handle geometric calculations better
	Phi  = 1.618033988749895
)

// Gcd calculates the greatest common divisor
func Gcd(a int, b int) int {
	if b == 0 {
		return a
	} else {
		return Gcd(b, a%b)
	}
}

// Lerp calculates the linear interpolation between two values
func Lerp(a float64, b float64, i float64) float64 {
	return a + i*(b-a)
}

// Map calculates the linear interpolation from one range to another
func Map(a float64, b float64, c float64, d float64, i float64) float64 {
	p := (i - a) / (b - a)
	return Lerp(c, d, p)
}

// Clamp restricts a value to a given range
func Clamp(a float64, b float64, c float64) float64 {
	if c <= a {
		return a
	}
	if c >= b {
		return b
	}
	return c
}

// NoTinyVals sets values very close to zero to zero
func NoTinyVals(a float64) float64 {
	if math.Abs(a) < Smol {
		return 0
	}
	return a
}

// Linspace creates a slice of n linearly distributed values from i. If b is
// true the range is inclusive of j, otherwise j is the exclusive upper bound.
func Linspace(i float64, j float64, n int, b bool) []float64 {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		// An inclusive single-sample range would otherwise divide by zero.
		return []float64{i}
	}
	N := float64(n)
	if b {
		N -= 1
	}
	d := (j - i) / N
	result := make([]float64, n)
	for k := 0; k < n; k++ {
		result[k] = i + float64(k)*d
	}
	return result
}

// Deg2Rad converts from degrees to radians
func Deg2Rad(f float64) float64 {
	return math.Pi * f / 180
}

// Rad2Deg converts from radians to degrees
func Rad2Deg(f float64) float64 {
	return 180 * f / math.Pi
}

// Shuffle randomly permutes a slice of points using a time-seeded generator.
// Use [Rng.ShufflePoints] for a shuffle that is reproducible from a seed.
func Shuffle(p *[]Point) {
	s := *p
	if len(s) < 2 {
		return
	}
	// Fisher-Yates. The previous implementation did 3n random transpositions,
	// which is both slower and does not produce a uniform permutation.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

// GetTimestampString creates a string based on the current time for use in filenames
func GetTimestampString() string {
	now := time.Now()
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(),
		now.Minute(), now.Second())
}

func Equalf(a, b float64) bool {
	return math.Abs(b-a) <= Smol
}

func (p Point) ToIndexPoint(index int) IndexPoint {
	return IndexPoint{
		Index: index,
		Point: p,
	}
}

func (p Point) ToMetricPoint(index int, metric float64) MetricPoint {
	return MetricPoint{
		Metric: metric,
		Index:  index,
		Point:  p,
	}
}

// An IndexPoint is a wrapper around a point with an extra int identifier, useful when used with trees and heaps
type IndexPoint struct {
	Index int
	Point
}

func (p IndexPoint) ToPoint() Point {
	return p.Point
}

// A MetricPoint is a wrapper around a point with to extra identifiers, useful when used with trees and heaps
type MetricPoint struct {
	Metric float64
	Index  int
	Point
}

func (p MetricPoint) ToIndexPoint() IndexPoint {
	return IndexPoint{
		Index: p.Index,
		Point: p.Point,
	}
}

func (p MetricPoint) ToPoint() Point {
	return p.Point
}

func FloatString(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}

func Smoothstep(t float64) float64 {
	t = Clamp(0, 1, t)
	return t * t * (3 - 2*t)
}
