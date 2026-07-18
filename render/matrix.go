package render

import "math"

// Matrix is a 2D affine transformation:
//
//	| A C E |
//	| B D F |
//	| 0 0 1 |
//
// mapping (x, y) to (A*x + C*y + E, B*x + D*y + F). The field order matches
// the SVG matrix(a b c d e f) parameter order. Because the coordinate system
// is y-down, positive rotation angles rotate clockwise on screen.
type Matrix struct {
	A, B, C, D, E, F float64
}

// Identity returns the identity transform.
func Identity() Matrix {
	return Matrix{A: 1, D: 1}
}

// Translation returns a transform that translates by (tx, ty).
func Translation(tx, ty float64) Matrix {
	return Matrix{A: 1, D: 1, E: tx, F: ty}
}

// Scaling returns a transform that scales by (sx, sy) about the origin.
func Scaling(sx, sy float64) Matrix {
	return Matrix{A: sx, D: sy}
}

// Rotation returns a transform that rotates by angle radians about the
// origin. Positive angles rotate clockwise on screen (y-down).
func Rotation(angle float64) Matrix {
	s, c := math.Sincos(angle)
	return Matrix{A: c, B: s, C: -s, D: c}
}

// Shearing returns a transform that shears by sx horizontally and sy
// vertically.
func Shearing(sx, sy float64) Matrix {
	return Matrix{A: 1, B: sy, C: sx, D: 1}
}

// Mul returns m * n, the transform that applies n first and then m.
func (m Matrix) Mul(n Matrix) Matrix {
	return Matrix{
		A: m.A*n.A + m.C*n.B,
		B: m.B*n.A + m.D*n.B,
		C: m.A*n.C + m.C*n.D,
		D: m.B*n.C + m.D*n.D,
		E: m.A*n.E + m.C*n.F + m.E,
		F: m.B*n.E + m.D*n.F + m.F,
	}
}

// Apply transforms the point (x, y).
func (m Matrix) Apply(x, y float64) (float64, float64) {
	return m.A*x + m.C*y + m.E, m.B*x + m.D*y + m.F
}

// Invert returns the inverse transform. The second return value is false if
// m is singular, in which case the identity is returned.
func (m Matrix) Invert() (Matrix, bool) {
	det := m.A*m.D - m.B*m.C
	if det == 0 {
		return Identity(), false
	}
	inv := 1 / det
	return Matrix{
		A: m.D * inv,
		B: -m.B * inv,
		C: -m.C * inv,
		D: m.A * inv,
		E: (m.C*m.F - m.D*m.E) * inv,
		F: (m.B*m.E - m.A*m.F) * inv,
	}, true
}

// IsIdentity reports whether m is exactly the identity transform.
func (m Matrix) IsIdentity() bool {
	return m == Matrix{A: 1, D: 1}
}

// isIntTranslation reports whether m is a pure translation by whole pixels,
// returning the integer offsets. Used for fast-path image blitting.
func (m Matrix) isIntTranslation() (dx, dy int, ok bool) {
	if m.A != 1 || m.B != 0 || m.C != 0 || m.D != 1 {
		return 0, 0, false
	}
	if m.E != math.Trunc(m.E) || m.F != math.Trunc(m.F) {
		return 0, 0, false
	}
	return int(m.E), int(m.F), true
}
