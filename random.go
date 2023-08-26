package gaul

import (
	"math/rand"
	"time"

	"github.com/ojrac/opensimplex-go"
)

const (
	defaultScale       = 0.001
	defaultOctaves     = 1
	defaultPersistence = 0.9
	defaultLacunarity  = 2.0
)

// Pseudo-random number generator data
type Rng struct {
	seed        int64
	prng        *rand.Rand
	noise       opensimplex.Noise
	octaves     int
	persistence float64
	lacunarity  float64
	xscale      float64
	yscale      float64
	zscale      float64
	xoffset     float64
	yoffset     float64
	zoffset     float64
}

// Returns a PRNG with a system and noise generator
func NewRng(i int64) Rng {
	return Rng{
		seed:        i,
		prng:        rand.New(rand.NewSource(i)),
		noise:       opensimplex.New(i),
		octaves:     defaultOctaves,
		persistence: defaultPersistence,
		lacunarity:  defaultLacunarity,
		xscale:      defaultScale,
		yscale:      defaultScale,
		zscale:      defaultScale,
		xoffset:     0,
		yoffset:     0,
		zoffset:     0,
	}
}

func (r *Rng) Gaussian(mean float64, stdev float64) float64 {
	return rand.NormFloat64()*stdev + mean
}

// The noise scale functions scale the position values passed into the
// noise PRNG. Typically for screen coordinates scale values in the
// range of 0.001 to 0.01 produce visually appealing noise

// Scales the x position in noise calculations
func (r *Rng) SetNoiseScaleX(scale float64) {
	r.xscale = scale
}

// Scales the y position in noise calculations
func (r *Rng) SetNoiseScaleY(scale float64) {
	r.yscale = scale
}

// Scales the z position in noise calculations
func (r *Rng) SetNoiseScaleZ(scale float64) {
	r.zscale = scale
}

// The noise offset functions simple increment/decrement the
// position values before scaling

// Offsets the x position in noise calculations
func (r *Rng) SetNoiseOffsetX(offset float64) {
	r.xoffset = offset
}

// Offsets the y position in noise calculations
func (r *Rng) SetNoiseOffsetY(offset float64) {
	r.yoffset = offset
}

// Offsets the z position in noise calculations
func (r *Rng) SetNoiseOffsetZ(offset float64) {
	r.zoffset = offset
}

// Number of steps when calculating fractal noise
func (r *Rng) SetNoiseOctaves(i int) {
	r.octaves = i
}

// How amplitude scales with octaves
func (r *Rng) SetNoisePersistence(p float64) {
	r.persistence = p
}

// How frequency scales with octaves
func (r *Rng) SetNoiseLacunarity(l float64) {
	r.lacunarity = l
}

// SignedNoise1D generates 1D noise values in the range of [-1, 1]
func (r *Rng) SignedNoise1D(x float64) float64 {
	return r.calcNoise(x, 0, 0)
}

// SignedNoise2D generates 2D noise values in the range of [-1, 1]
func (r *Rng) SignedNoise2D(x float64, y float64) float64 {
	return r.calcNoise(x, y, 0)
}

// SignedNoise3D generates 3D noise values in the range of [-1, 1]
func (r *Rng) SignedNoise3D(x float64, y float64, z float64) float64 {
	return r.calcNoise(x, y, z)
}

// Noise1D 1D noise values in the range of [0, 1]
func (r *Rng) Noise1D(x float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise(x, 0, 0))
}

// Noise2D generates 2D noise values in the range of [0, 1]
func (r *Rng) Noise2D(x float64, y float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise(x, y, 0))
}

// Noise3D generates 3D noise values in the range of [0, 1]
func (r *Rng) Noise3D(x float64, y float64, z float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise(x, y, z))
}

// UniformRandomPoints generates a list of points whose coordinates
// follow a uniform random distribution within a rectangle
func (r *Rng) UniformRandomPoints(num int, rect Rect) []Point {
	points := make([]Point, num)
	for i := 0; i < num; i++ {
		x := rect.X + rand.Float64()*rect.W
		y := rect.Y + rand.Float64()*rect.H
		points[i] = Point{X: x, Y: y}
	}
	return points
}

func (r *Rng) NoisyRandomPoints(num int, threshold float64, rect Rect) []Point {
	var points []Point
	maxtries := 10 * num
	i := 0
	for len(points) < num && i < maxtries {
		x := rect.X + rand.Float64()*rect.W
		y := rect.Y + rand.Float64()*rect.H
		noise := r.Noise2D(x, y)
		if noise >= threshold {
			points = append(points, Point{X: x, Y: y})
		}
		i++
	}
	return points
}

func (r *Rng) calcNoise(x, y, z float64) float64 {
	totalNoise := 0.0
	totalAmp := 0.0
	amp := 1.0
	freq := 1.0
	for i := 0; i < r.octaves; i++ {
		totalNoise += r.noise.Eval3(
			(x+r.xoffset)*r.xscale*freq,
			(y+r.yoffset)*r.yscale*freq,
			(z+r.zoffset)*r.zscale*freq,
		)
		totalAmp += amp
		amp *= r.persistence
		freq *= r.lacunarity
	}
	return totalNoise / totalAmp
}

type LFSRSmall struct {
	state uint16
}

func NewLFSRSmallWithSeed(seed uint16) LFSRSmall {
	return LFSRSmall{state: seed}
}

func NewLFSRSmall() LFSRSmall {
	return LFSRSmall{state: uint16(time.Now().UnixNano())}
}

func (l *LFSRSmall) Next() uint16 {
	b := ((l.state >> 0) ^ (l.state >> 2) ^ (l.state >> 3) ^ (l.state >> 5)) & 1
	l.state = (l.state >> 1) | (b << 15)
	return l.state
}

type LFSRMedium struct {
	state uint32
}

func NewLFSRMediumWithSeed(seed uint32) LFSRMedium {
	return LFSRMedium{state: seed}
}

func NewLFSRMedium() LFSRMedium {
	return LFSRMedium{state: uint32(time.Now().UnixNano())}
}

func (l *LFSRMedium) Next() uint32 {
	l.state ^= l.state << 13
	l.state ^= l.state >> 17
	l.state ^= l.state << 5
	return l.state
}

// LFSRLarge is a 64-bit Xorshift PRNG
// Benchmarks show this is faster than LFSRMedium and LFSRSmall, so you might as well use this one.
// Benchmarks also show this is 6-7x faster than the standard math/rand PRNG
type LFSRLarge struct {
	state uint64
}

func NewLFSRLargeWithSeed(seed uint64) LFSRLarge {
	return LFSRLarge{state: seed}
}

func NewLFSRLarge() LFSRLarge {
	return LFSRLarge{state: uint64(time.Now().UnixNano())}
}

func (l *LFSRLarge) Next() uint64 {
	l.state ^= l.state << 13
	l.state ^= l.state >> 7
	l.state ^= l.state << 17
	return l.state
}

func (l *LFSRLarge) Float64() float64 {
	return float64(l.Next()) / float64(^uint64(0))
}

func (l *LFSRLarge) Uint64n(n uint64) uint64 {
	return l.Next() % n
}

func (l *LFSRLarge) Uint64() uint64 {
	return l.Next()
}
