package gaul

import (
	"math"
	"testing"
)

func TestTriangle_ContainsPoint(t *testing.T) {
	// Create a simple triangle with vertices at (0,0), (1,0), and (0,1)
	triangle := Triangle{
		A: Point{X: 0, Y: 0},
		B: Point{X: 1, Y: 0},
		C: Point{X: 0, Y: 1},
	}

	tests := []struct {
		name     string
		point    Point
		expected bool
	}{
		{
			name:     "point inside triangle",
			point:    Point{X: 0.2, Y: 0.2},
			expected: true,
		},
		{
			name:     "point outside triangle",
			point:    Point{X: 1.5, Y: 1.5},
			expected: false,
		},
		{
			name:     "point on vertex",
			point:    Point{X: 0, Y: 0},
			expected: true,
		},
		{
			name:     "point on edge",
			point:    Point{X: 0.5, Y: 0},
			expected: true,
		},
		{
			name:     "point outside but aligned with triangle",
			point:    Point{X: 0.5, Y: 1.5},
			expected: false,
		},
		{
			name:     "point at centroid",
			point:    Point{X: 1.0 / 3.0, Y: 1.0 / 3.0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := triangle.ContainsPoint(tt.point)
			if got != tt.expected {
				t.Errorf("Triangle.ContainsPoint() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test with a different triangle orientation
func TestTriangle_ContainsPoint_Rotated(t *testing.T) {
	// Create a triangle with vertices at (1,1), (2,1), and (1,2)
	triangle := Triangle{
		A: Point{X: 1, Y: 1},
		B: Point{X: 2, Y: 1},
		C: Point{X: 1, Y: 2},
	}

	tests := []struct {
		name     string
		point    Point
		expected bool
	}{
		{
			name:     "point inside translated triangle",
			point:    Point{X: 1.2, Y: 1.2},
			expected: true,
		},
		{
			name:     "point outside translated triangle",
			point:    Point{X: 2.5, Y: 2.5},
			expected: false,
		},
		{
			name:     "point on translated triangle vertex",
			point:    Point{X: 1, Y: 1},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := triangle.ContainsPoint(tt.point)
			if got != tt.expected {
				t.Errorf("Triangle.ContainsPoint() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTriangle_IsIsosceles(t *testing.T) {
	tests := []struct {
		name     string
		triangle Triangle
		expected bool
	}{
		{
			name: "isosceles triangle (AB = AC)",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 1},
			},
			expected: true,
		},
		{
			name: "isosceles triangle (AB = BC)",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.5},
			},
			expected: true,
		},
		{
			name: "isosceles triangle (AC = BC)",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.866}, // sqrt(3)/2
			},
			expected: true,
		},
		{
			name: "scalene triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.3, Y: 0.7}, // Actually scalene
			},
			expected: false,
		},
		{
			name: "equilateral triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.866}, // sqrt(3)/2
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.triangle.IsIsosceles()
			if got != tt.expected {
				t.Errorf("Triangle.IsIsosceles() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTriangle_IsEquilateral(t *testing.T) {
	tests := []struct {
		name     string
		triangle Triangle
		expected bool
	}{
		{
			name: "equilateral triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.866025404}, // sqrt(3)/2 more precise
			},
			expected: true,
		},
		{
			name: "isosceles triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 1},
			},
			expected: false,
		},
		{
			name: "scalene triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.3, Y: 0.7},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.triangle.IsEquilateral()
			if got != tt.expected {
				t.Errorf("Triangle.IsEquilateral() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTriangle_IsRight(t *testing.T) {
	tests := []struct {
		name     string
		triangle Triangle
		expected bool
	}{
		{
			name: "right triangle (90° at A)",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			expected: true,
		},
		{
			name: "right triangle (90° at B)",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 1, Y: 1},
			},
			expected: true,
		},
		{
			name: "right triangle (90° at C)",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 1},
				C: Point{X: 0, Y: 1},
			},
			expected: true,
		},
		{
			name: "acute triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.3, Y: 0.4}, // Actually acute
			},
			expected: false,
		},
		{
			name: "obtuse triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.2},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.triangle.IsRight()
			if got != tt.expected {
				t.Errorf("Triangle.IsRight() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTriangle_IsAcute(t *testing.T) {
	tests := []struct {
		name     string
		triangle Triangle
		expected bool
	}{
		{
			name: "acute triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.6}, // Creates a triangle with all angles < 90°
			},
			expected: true,
		},
		{
			name: "right triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			expected: false,
		},
		{
			name: "obtuse triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.2},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.triangle.IsAcute()
			if got != tt.expected {
				t.Errorf("Triangle.IsAcute() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTriangle_IsObtuse(t *testing.T) {
	tests := []struct {
		name     string
		triangle Triangle
		expected bool
	}{
		{
			name: "obtuse triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.2},
			},
			expected: true,
		},
		{
			name: "right triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			expected: false,
		},
		{
			name: "acute triangle",
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0.5, Y: 0.5},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.triangle.IsObtuse()
			if got != tt.expected {
				t.Errorf("Triangle.IsObtuse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLineSDF(t *testing.T) {
	// Test cases for Line.SDF
	tests := []struct {
		line     Line
		point    Point
		expected float64
	}{
		{
			line:     Line{P: Point{X: 0, Y: 0}, Q: Point{X: 1, Y: 0}},
			point:    Point{X: 0.5, Y: 1},
			expected: 1.0,
		},
		{
			line:     Line{P: Point{X: 0, Y: 0}, Q: Point{X: 1, Y: 0}},
			point:    Point{X: 0.5, Y: 0},
			expected: 0.0,
		},
		{
			line:     Line{P: Point{X: 0, Y: 0}, Q: Point{X: 1, Y: 1}},
			point:    Point{X: 0, Y: 1},
			expected: math.Sqrt(2) / 2,
		},
		{
			line:     Line{P: Point{X: 0, Y: 0}, Q: Point{X: 1, Y: 0}},
			point:    Point{X: -1, Y: 0},
			expected: 1.0,
		},
	}

	for i, test := range tests {
		result := test.line.SDF(test.point)
		if !Equalf(result, test.expected) {
			t.Errorf("Test %d: Expected %f, got %f", i, test.expected, result)
		}
	}
}

func TestTriangleSDF(t *testing.T) {
	// Test cases for Triangle.SDF
	tests := []struct {
		triangle Triangle
		point    Point
		expected float64
	}{
		{
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			point:    Point{X: 0.25, Y: 0.25},
			expected: -0.25, // Inside triangle, negative distance
		},
		{
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			point:    Point{X: 0.5, Y: 0.5},
			expected: 0.0, // On edge of triangle
		},
		{
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			point:    Point{X: 0, Y: 0},
			expected: 0.0, // On vertex
		},
		{
			triangle: Triangle{
				A: Point{X: 0, Y: 0},
				B: Point{X: 1, Y: 0},
				C: Point{X: 0, Y: 1},
			},
			point:    Point{X: 0.5, Y: 0},
			expected: 0.0, // On edge
		},
	}

	for i, test := range tests {
		result := test.triangle.SDF(test.point)
		t.Logf("Test %d: Point %v, Inside: %v, Result: %f, Expected: %f",
			i, test.point, test.triangle.ContainsPoint(test.point), result, test.expected)
		if !Equalf(result, test.expected) {
			t.Errorf("Test %d: Expected %f, got %f", i, test.expected, result)
		}
	}
}
