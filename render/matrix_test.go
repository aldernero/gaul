package render

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

const eps = 1e-9

func TestMatrixBasics(t *testing.T) {
	x, y := Identity().Apply(3, 4)
	assert.InDelta(t, 3.0, x, eps)
	assert.InDelta(t, 4.0, y, eps)
	assert.True(t, Identity().IsIdentity())

	x, y = Translation(10, -5).Apply(3, 4)
	assert.InDelta(t, 13.0, x, eps)
	assert.InDelta(t, -1.0, y, eps)

	x, y = Scaling(2, 3).Apply(3, 4)
	assert.InDelta(t, 6.0, x, eps)
	assert.InDelta(t, 12.0, y, eps)

	// Positive rotation is clockwise on screen (y-down): rotating the
	// point (1, 0) by 90 degrees lands on (0, 1), i.e. straight down.
	x, y = Rotation(math.Pi/2).Apply(1, 0)
	assert.InDelta(t, 0.0, x, eps)
	assert.InDelta(t, 1.0, y, eps)

	x, y = Shearing(0.5, 0).Apply(0, 2)
	assert.InDelta(t, 1.0, x, eps)
	assert.InDelta(t, 2.0, y, eps)
}

func TestMatrixMulOrder(t *testing.T) {
	// m.Mul(n) applies n first: translate-then-scale vs scale-then-translate.
	scaleThenTranslate := Scaling(2, 2).Mul(Translation(10, 0))
	x, y := scaleThenTranslate.Apply(1, 1)
	assert.InDelta(t, 22.0, x, eps)
	assert.InDelta(t, 2.0, y, eps)

	translateThenScale := Translation(10, 0).Mul(Scaling(2, 2))
	x, y = translateThenScale.Apply(1, 1)
	assert.InDelta(t, 12.0, x, eps)
	assert.InDelta(t, 2.0, y, eps)
}

func TestMatrixInvert(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for range 100 {
		m := Translation(rng.Float64()*100-50, rng.Float64()*100-50).
			Mul(Rotation(rng.Float64() * 2 * math.Pi)).
			Mul(Scaling(rng.Float64()*3+0.1, rng.Float64()*3+0.1)).
			Mul(Shearing(rng.Float64()-0.5, rng.Float64()-0.5))
		inv, ok := m.Invert()
		assert.True(t, ok)
		x, y := rng.Float64()*100, rng.Float64()*100
		tx, ty := m.Apply(x, y)
		rx, ry := inv.Apply(tx, ty)
		assert.InDelta(t, x, rx, 1e-6)
		assert.InDelta(t, y, ry, 1e-6)
	}

	_, ok := Scaling(0, 1).Invert()
	assert.False(t, ok)
}
