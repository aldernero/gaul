package gaul

// PointHeap is a binary min/max heap of MetricPoints ordered by their Metric
// field, which is typically an inter-point distance.
type PointHeap struct {
	points    []MetricPoint
	isMinHeap bool
}

// NewMaxPointHeap returns an empty heap whose root is the largest metric.
func NewMaxPointHeap() *PointHeap {
	return &PointHeap{isMinHeap: false}
}

// NewMinPointHeap returns an empty heap whose root is the smallest metric.
func NewMinPointHeap() *PointHeap {
	return &PointHeap{isMinHeap: true}
}

// Len returns the number of points in the heap.
func (m *PointHeap) Len() int {
	return len(m.points)
}

// Push adds a point to the heap.
func (m *PointHeap) Push(p MetricPoint) {
	m.points = append(m.points, p)
	m.siftUp(len(m.points) - 1)
}

// Peek returns the root of the heap without removing it. It panics if the
// heap is empty.
func (m *PointHeap) Peek() MetricPoint {
	if len(m.points) == 0 {
		panic("can't peek empty heap")
	}
	return m.points[0]
}

// Pop removes and returns the root of the heap. It panics if the heap is
// empty.
func (m *PointHeap) Pop() MetricPoint {
	n := len(m.points)
	if n == 0 {
		panic("can't pop empty heap")
	}
	root := m.points[0]
	// Move the last element to the root and sift it down. Re-slicing off the
	// front instead would renumber every node and destroy the heap ordering.
	last := n - 1
	m.points[0] = m.points[last]
	m.points = m.points[:last]
	m.siftDown(0)
	return root
}

// Report empties the heap, returning the points in heap order (ascending
// metric for a min-heap, descending for a max-heap).
func (m *PointHeap) Report() []MetricPoint {
	n := len(m.points)
	result := make([]MetricPoint, n)
	for i := 0; i < n; i++ {
		result[i] = m.Pop()
	}
	return result
}

// ReportReversed empties the heap, returning the points in the reverse of
// heap order.
func (m *PointHeap) ReportReversed() []MetricPoint {
	n := len(m.points)
	result := make([]MetricPoint, n)
	for i := 0; i < n; i++ {
		result[n-i-1] = m.Pop()
	}
	return result
}

// higher reports whether the point at index i belongs above the point at
// index j in the heap.
func (m *PointHeap) higher(i, j int) bool {
	if m.isMinHeap {
		return m.points[i].Metric < m.points[j].Metric
	}
	return m.points[i].Metric > m.points[j].Metric
}

func (m *PointHeap) siftUp(i int) {
	for i > 0 {
		parent := m.parent(i)
		if !m.higher(i, parent) {
			return
		}
		m.swap(i, parent)
		i = parent
	}
}

func (m *PointHeap) siftDown(i int) {
	n := len(m.points)
	for {
		best := i
		if l := m.left(i); l < n && m.higher(l, best) {
			best = l
		}
		if r := m.right(i); r < n && m.higher(r, best) {
			best = r
		}
		if best == i {
			return
		}
		m.swap(i, best)
		i = best
	}
}

func (m *PointHeap) parent(i int) int {
	return (i - 1) / 2
}

func (m *PointHeap) left(i int) int {
	return 2*i + 1
}

func (m *PointHeap) right(i int) int {
	return 2*i + 2
}

func (m *PointHeap) swap(i, j int) {
	m.points[i], m.points[j] = m.points[j], m.points[i]
}
