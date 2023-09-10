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

// Rng is a random number generator with a system PRNG and simplex noise
type Rng struct {
	seed        int64
	Prng        *rand.Rand
	Noise       opensimplex.Noise
	octaves     int
	persistence float64
	lacunarity  float64
	xscale      float64
	yscale      float64
	zscale      float64
	wscale      float64
	xoffset     float64
	yoffset     float64
	zoffset     float64
	woffset     float64
}

// NewRng returns a PRNG with a system and Noise generator
func NewRng(i int64) Rng {
	return Rng{
		seed:        i,
		Prng:        rand.New(rand.NewSource(i)),
		Noise:       opensimplex.New(i),
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

func (r *Rng) SetSeed(seed int64) {
	r.seed = seed
	r.Prng = rand.New(rand.NewSource(seed))
	r.Noise = opensimplex.NewNormalized(seed)
}

func (r *Rng) Gaussian(mean float64, stdev float64) float64 {
	return rand.NormFloat64()*stdev + mean
}

// The Noise scale functions scale the position values passed into the
// Noise PRNG. Typically for screen coordinates scale values in the
// range of 0.001 to 0.01 produce visually appealing Noise

// SetNoiseScaleX scales the x position in Noise calculations
func (r *Rng) SetNoiseScaleX(scale float64) {
	r.xscale = scale
}

// SetNoiseScaleY scales the y position in Noise calculations
func (r *Rng) SetNoiseScaleY(scale float64) {
	r.yscale = scale
}

// SetNoiseScaleZ scales the z position in Noise calculations
func (r *Rng) SetNoiseScaleZ(scale float64) {
	r.zscale = scale
}

// The Noise offset functions simple increment/decrement the
// position values before scaling

// SetNoiseOffsetX offsets the x position in Noise calculations
func (r *Rng) SetNoiseOffsetX(offset float64) {
	r.xoffset = offset
}

// SetNoiseOffsetY offsets the y position in Noise calculations
func (r *Rng) SetNoiseOffsetY(offset float64) {
	r.yoffset = offset
}

// SetNoiseOffsetZ offsets the z position in Noise calculations
func (r *Rng) SetNoiseOffsetZ(offset float64) {
	r.zoffset = offset
}

// SetNoiseOctaves sets the number of steps when calculating fractal Noise
func (r *Rng) SetNoiseOctaves(i int) {
	r.octaves = i
}

// SetNoisePersistence sets how amplitude scales with octaves
func (r *Rng) SetNoisePersistence(p float64) {
	r.persistence = p
}

// SetNoiseLacunarity sets how frequency scales with octaves
func (r *Rng) SetNoiseLacunarity(l float64) {
	r.lacunarity = l
}

// SignedNoise1D generates 1D Noise values in the range of [-1, 1]
func (r *Rng) SignedNoise1D(x float64) float64 {
	return r.calcNoise1(x)
}

// SignedNoise2D generates 2D Noise values in the range of [-1, 1]
func (r *Rng) SignedNoise2D(x float64, y float64) float64 {
	return r.calcNoise2(x, y)
}

// SignedNoise3D generates 3D Noise values in the range of [-1, 1]
func (r *Rng) SignedNoise3D(x float64, y float64, z float64) float64 {
	return r.calcNoise3(x, y, z)
}

// Noise1D 1D Noise values in the range of [0, 1]
func (r *Rng) Noise1D(x float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise1(x))
}

// Noise2D generates 2D Noise values in the range of [0, 1]
func (r *Rng) Noise2D(x float64, y float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise2(x, y))
}

// Noise3D generates 3D Noise values in the range of [0, 1]
func (r *Rng) Noise3D(x float64, y float64, z float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise3(x, y, z))
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

func (r *Rng) calcNoise1(x float64) float64 {
	return r.calcNoise(1, x, 0, 0, 0)
}

func (r *Rng) calcNoise2(x, y float64) float64 {
	return r.calcNoise(2, x, y, 0, 0)
}

func (r *Rng) calcNoise3(x, y, z float64) float64 {
	return r.calcNoise(3, x, y, z, 0)
}

func (r *Rng) calcNoise4(x, y, z, w float64) float64 {
	return r.calcNoise(4, x, y, z, w)
}

func (r *Rng) calcNoise(dim int, x, y, z, w float64) float64 {
	totalNoise := 0.0
	totalAmp := 0.0
	amp := 1.0
	freq := 1.0
	for i := 0; i < r.octaves; i++ {
		switch dim {
		case 1:
			totalNoise += r.Noise.Eval2(
				(x+r.xoffset)*r.xscale*freq,
				0,
			)
		case 2:
			totalNoise += r.Noise.Eval2(
				(x+r.xoffset)*r.xscale*freq,
				(y+r.yoffset)*r.yscale*freq,
			)
		case 3:
			totalNoise += r.Noise.Eval3(
				(x+r.xoffset)*r.xscale*freq,
				(y+r.yoffset)*r.yscale*freq,
				(z+r.zoffset)*r.zscale*freq,
			)
		case 4:
			totalNoise += r.Noise.Eval4(
				(x+r.xoffset)*r.xscale*freq,
				(y+r.yoffset)*r.yscale*freq,
				(z+r.zoffset)*r.zscale*freq,
				(w+r.zoffset)*r.wscale*freq,
			)
		}
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
