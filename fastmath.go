package gaul

import "math"

const (
	LUTSize   = 1 << 16
	LUTMask   = LUTSize - 1
	LUTFactor = LUTSize / Tau
)

// TrigLUT is a lookup table for sine and cosine.
// In my tests, it is about 9x faster than math.Sin and math.Cos.
// The margin of error is < 0.005%.
// This LUT is not suitable for tangents, however.
type TrigLUT struct {
	sinTable []float64
}

func NewTrigLUT() *TrigLUT {
	lut := TrigLUT{}
	lut.sinTable = make([]float64, LUTSize)
	for i, a := range Linspace(0, Tau, LUTSize, false) {
		lut.sinTable[i] = math.Sin(a)
	}
	return &lut
}

func (lut *TrigLUT) Sin(x float64) float64 {
	// pos is the (fractional) table index for x. Flooring rather than
	// truncating keeps the index and the interpolation fraction consistent
	// for negative x, and masking wraps the index into the table.
	pos := x * LUTFactor
	base := math.Floor(pos)
	i := int(int64(base) & LUTMask)
	j := (i + 1) & LUTMask
	return Lerp(lut.sinTable[i], lut.sinTable[j], pos-base)
}

func (lut *TrigLUT) Cos(x float64) float64 {
	return lut.Sin(x + Tau/4)
}
