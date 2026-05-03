// Voronoi diagram via Fortune's sweep line algorithm.
//
// Core implementation derived from the MIT-licensed package
// https://github.com/pzsz/voronoi (Przemyslaw Szczepaniak), itself a port of
// Raymond Hill's JavaScript implementation of Steven Fortune's algorithm.
// Adapted for package gaul (Point/Curve API, VoronoiWithRect / VoronoiWithCurve entrypoints).

package gaul

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var noVoronoiVertex = Point{X: math.Inf(1), Y: math.Inf(1)}

// Red-Black tree code (based on C version of "rbtree" by Franck Bui-Huu
// https://github.com/fbuihuu/libtree/blob/master/rb.c
type rbTree struct {
	root *rbNode
}

type rbNodeValue interface {
	bindToNode(node *rbNode)
	getNode() *rbNode
}

type rbNode struct {
	value    rbNodeValue
	left     *rbNode
	right    *rbNode
	parent   *rbNode
	previous *rbNode
	next     *rbNode
	red      bool
}

func (t *rbTree) insertSuccessor(node *rbNode, vsuccessor rbNodeValue) {
	successor := &rbNode{value: vsuccessor}
	vsuccessor.bindToNode(successor)

	var parent *rbNode
	if node != nil {
		// >>> rhill 2011-05-27: Performance: cache previous/next nodes
		successor.previous = node
		successor.next = node.next
		if node.next != nil {
			node.next.previous = successor
		}
		node.next = successor
		// <<<
		if node.right != nil {
			// in-place expansion of node.rbRight.getFirst()
			node = node.right
			for ; node.left != nil; node = node.left {
			}
			node.left = successor
		} else {
			node.right = successor
		}
		parent = node

		// rhill 2011-06-07: if node is null, successor must be inserted
		// to the left-most part of the tree
	} else if t.root != nil {
		node = t.getFirst(t.root)
		// >>> Performance: cache previous/next nodes
		successor.previous = nil
		successor.next = node
		node.previous = successor
		// <<<
		node.left = successor
		parent = node
	} else {
		// >>> Performance: cache previous/next nodes
		successor.previous = nil
		successor.next = nil
		// <<<
		t.root = successor
		parent = nil
	}
	successor.left = nil
	successor.right = nil
	successor.parent = parent
	successor.red = true
	// Fixup the modified tree by recoloring nodes and performing
	// rotations (2 at most) hence the red-black tree properties are
	// preserved.
	var grandpa, uncle *rbNode
	node = successor
	for parent != nil && parent.red {
		grandpa = parent.parent
		if parent == grandpa.left {
			uncle = grandpa.right
			if uncle != nil && uncle.red {
				parent.red = false
				uncle.red = false
				grandpa.red = true
				node = grandpa
			} else {
				if node == parent.right {
					t.rotateLeft(parent)
					node = parent
					parent = node.parent
				}
				parent.red = false
				grandpa.red = true
				t.rotateRight(grandpa)
			}
		} else {
			uncle = grandpa.left
			if uncle != nil && uncle.red {
				parent.red = false
				uncle.red = false
				grandpa.red = true
				node = grandpa
			} else {
				if node == parent.left {
					t.rotateRight(parent)
					node = parent
					parent = node.parent
				}
				parent.red = false
				grandpa.red = true
				t.rotateLeft(grandpa)
			}
		}
		parent = node.parent
	}
	t.root.red = false
}

func (t *rbTree) removeNode(node *rbNode) {
	// >>> rhill 2011-05-27: Performance: cache previous/next nodes
	if node.next != nil {
		node.next.previous = node.previous
	}
	if node.previous != nil {
		node.previous.next = node.next
	}
	node.next = nil
	node.previous = nil
	// <<<
	var parent = node.parent
	var left = node.left
	var right = node.right
	var next *rbNode
	if left == nil {
		next = right
	} else if right == nil {
		next = left
	} else {
		next = t.getFirst(right)
	}
	if parent != nil {
		if parent.left == node {
			parent.left = next
		} else {
			parent.right = next
		}
	} else {
		t.root = next
	}
	// enforce red-black rules
	isRed := false
	if left != nil && right != nil {
		isRed = next.red
		next.red = node.red
		next.left = left
		left.parent = next
		if next != right {
			parent = next.parent
			next.parent = node.parent
			node = next.right
			parent.left = node
			next.right = right
			right.parent = next
		} else {
			next.parent = parent
			parent = next
			node = next.right
		}
	} else {
		isRed = node.red
		node = next
	}
	// 'node' is now the sole successor's child and 'parent' its
	// new parent (since the successor can have been moved)
	if node != nil {
		node.parent = parent
	}
	// the 'easy' cases
	if isRed {
		return
	}
	if node != nil && node.red {
		node.red = false
		return
	}
	// the other cases
	var sibling *rbNode
	for {
		if node == t.root {
			break
		}
		if node == parent.left {
			sibling = parent.right
			if sibling.red {
				sibling.red = false
				parent.red = true
				t.rotateLeft(parent)
				sibling = parent.right
			}
			if (sibling.left != nil && sibling.left.red) || (sibling.right != nil && sibling.right.red) {
				if sibling.right == nil || !sibling.right.red {
					sibling.left.red = false
					sibling.red = true
					t.rotateRight(sibling)
					sibling = parent.right
				}
				sibling.red = parent.red
				parent.red = false
				sibling.right.red = false
				t.rotateLeft(parent)
				node = t.root
				break
			}
		} else {
			sibling = parent.left
			if sibling.red {
				sibling.red = false
				parent.red = true
				t.rotateRight(parent)
				sibling = parent.left
			}
			if (sibling.left != nil && sibling.left.red) || (sibling.right != nil && sibling.right.red) {
				if sibling.left == nil || !sibling.left.red {
					sibling.right.red = false
					sibling.red = true
					t.rotateLeft(sibling)
					sibling = parent.left
				}
				sibling.red = parent.red
				parent.red = false
				sibling.left.red = false
				t.rotateRight(parent)
				node = t.root
				break
			}
		}
		sibling.red = true
		node = parent
		parent = parent.parent
		if node.red {
			break
		}
	}
	if node != nil {
		node.red = false
	}
}

func (t *rbTree) rotateLeft(node *rbNode) {
	var p = node
	var q = node.right // can't be null
	var parent = p.parent
	if parent != nil {
		if parent.left == p {
			parent.left = q
		} else {
			parent.right = q
		}
	} else {
		t.root = q
	}
	q.parent = parent
	p.parent = q
	p.right = q.left
	if p.right != nil {
		p.right.parent = p
	}
	q.left = p
}

func (t *rbTree) rotateRight(node *rbNode) {
	var p = node
	var q = node.left // can't be null
	var parent = p.parent
	if parent != nil {
		if parent.left == p {
			parent.left = q
		} else {
			parent.right = q
		}
	} else {
		t.root = q
	}
	q.parent = parent
	p.parent = q
	p.left = q.right
	if p.left != nil {
		p.left.parent = p
	}
	q.right = p
}

func (t *rbTree) getFirst(node *rbNode) *rbNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (t *rbTree) getLast(node *rbNode) *rbNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

type EdgeVertex struct {
	Point
	Edges []*Edge
}

// Edge structure
type Edge struct {
	// Cell on the left
	LeftCell *Cell
	// Cell on the right
	RightCell *Cell
	// Start Point
	Va EdgeVertex
	// End Point
	Vb EdgeVertex
}

func (e *Edge) GetOtherCell(cell *Cell) *Cell {
	if cell == e.LeftCell {
		return e.RightCell
	} else if cell == e.RightCell {
		return e.LeftCell
	}
	return nil
}

func (e *Edge) GetOtherEdgeVertex(v Point) EdgeVertex {
	if v == e.Va.Point {
		return e.Vb
	} else if v == e.Vb.Point {
		return e.Va
	}
	return EdgeVertex{Point: noVoronoiVertex, Edges: nil}
}

func newEdge(LeftCell, RightCell *Cell) *Edge {
	return &Edge{
		LeftCell:  LeftCell,
		RightCell: RightCell,
		Va:        EdgeVertex{Point: noVoronoiVertex, Edges: nil},
		Vb:        EdgeVertex{Point: noVoronoiVertex, Edges: nil},
	}
}

// Halfedge (directed edge)
type Halfedge struct {
	Cell  *Cell
	Edge  *Edge
	Angle float64
}

// Sort interface for halfedges
type Halfedges []*Halfedge

func (s Halfedges) Len() int      { return len(s) }
func (s Halfedges) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// For sorting by angle
type halfedgesByAngle struct{ Halfedges }

func (s halfedgesByAngle) Less(i, j int) bool { return s.Halfedges[i].Angle > s.Halfedges[j].Angle }

func newHalfedge(edge *Edge, LeftCell, RightCell *Cell) *Halfedge {
	ret := &Halfedge{
		Cell: LeftCell,
		Edge: edge,
	}

	// 'angle' is a value to be used for properly sorting the
	// halfsegments counterclockwise. By convention, we will
	// use the angle of the line defined by the 'site to the left'
	// to the 'site to the right'.
	// However, border edges have no 'site to the right': thus we
	// use the angle of line perpendicular to the halfsegment (the
	// edge should have both end points defined in such case.)
	if RightCell != nil {
		ret.Angle = math.Atan2(RightCell.Site.Y-LeftCell.Site.Y, RightCell.Site.X-LeftCell.Site.X)
	} else {
		va := edge.Va
		vb := edge.Vb
		// rhill 2011-05-31: used to call GetStartpoint()/GetEndpoint(),
		// but for performance purpose, these are expanded in place here.
		if edge.LeftCell == LeftCell {
			ret.Angle = math.Atan2(vb.X-va.X, va.Y-vb.Y)
		} else {
			ret.Angle = math.Atan2(va.X-vb.X, vb.Y-va.Y)
		}
	}
	return ret
}

func (h *Halfedge) GetStartpoint() Point {
	if h.Edge.LeftCell == h.Cell {
		return h.Edge.Va.Point
	}
	return h.Edge.Vb.Point

}

func (h *Halfedge) GetEndpoint() Point {
	if h.Edge.LeftCell == h.Cell {
		return h.Edge.Vb.Point
	}
	return h.Edge.Va.Point
}

// Cell of voronoi diagram
type Cell struct {
	// Site of the cell
	Site Point
	// Array of halfedges sorted counterclockwise
	Halfedges []*Halfedge
}

func newCell(site Point) *Cell {
	return &Cell{Site: site}
}

func (t *Cell) prepare() int {
	halfedges := t.Halfedges
	iHalfedge := len(halfedges) - 1

	// get rid of unused halfedges
	// rhill 2011-05-27: Keep it simple, no point here in trying
	// to be fancy: dangling edges are a typically a minority.
	for ; iHalfedge >= 0; iHalfedge-- {
		edge := halfedges[iHalfedge].Edge

		if edge.Vb.Point == noVoronoiVertex || edge.Va.Point == noVoronoiVertex {
			halfedges[iHalfedge] = halfedges[len(halfedges)-1]
			halfedges = halfedges[:len(halfedges)-1]
		}
	}

	sort.Sort(halfedgesByAngle{halfedges})
	t.Halfedges = halfedges
	return len(halfedges)
}

type Voronoi struct {
	cells []*Cell
	edges []*Edge

	cellsMap map[Point]*Cell

	beachline        rbTree
	circleEvents     rbTree
	firstCircleEvent *circleEvent
}

type fortuneDiagram struct {
	Cells []*Cell
	Edges []*Edge
	//	EdgesVertices map[Point]EdgeVertex
}

func (s *Voronoi) getCell(site Point) *Cell {
	ret := s.cellsMap[site]
	if ret == nil {
		panic(fmt.Sprintf("Couldn't find cell for site %v", site))
	}
	return ret
}

func (s *Voronoi) createEdge(LeftCell, RightCell *Cell, va, vb Point) *Edge {
	edge := newEdge(LeftCell, RightCell)
	s.edges = append(s.edges, edge)
	if va != noVoronoiVertex {
		s.setEdgeStartpoint(edge, LeftCell, RightCell, va)
	}

	if vb != noVoronoiVertex {
		s.setEdgeEndpoint(edge, LeftCell, RightCell, vb)
	}

	lCell := LeftCell
	rCell := RightCell

	lCell.Halfedges = append(lCell.Halfedges, newHalfedge(edge, LeftCell, RightCell))
	rCell.Halfedges = append(rCell.Halfedges, newHalfedge(edge, RightCell, LeftCell))
	return edge
}

func (s *Voronoi) createBorderEdge(LeftCell *Cell, va, vb Point) *Edge {
	edge := newEdge(LeftCell, nil)
	edge.Va.Point = va
	edge.Vb.Point = vb

	s.edges = append(s.edges, edge)
	return edge
}

func (s *Voronoi) setEdgeStartpoint(edge *Edge, LeftCell, RightCell *Cell, vertex Point) {
	if edge.Va.Point == noVoronoiVertex && edge.Vb.Point == noVoronoiVertex {
		edge.Va.Point = vertex
		edge.LeftCell = LeftCell
		edge.RightCell = RightCell
	} else if edge.LeftCell == RightCell {
		edge.Vb.Point = vertex
	} else {
		edge.Va.Point = vertex
	}
}

func (s *Voronoi) setEdgeEndpoint(edge *Edge, LeftCell, RightCell *Cell, vertex Point) {
	s.setEdgeStartpoint(edge, RightCell, LeftCell, vertex)
}

type Beachsection struct {
	node        *rbNode
	site        Point
	circleEvent *circleEvent
	edge        *Edge
}

// rbNodeValue intergface
func (s *Beachsection) bindToNode(node *rbNode) {
	s.node = node
}

// rbNodeValue intergface
func (s *Beachsection) getNode() *rbNode {
	return s.node
}

// Calculate the left break point of a particular beach section,
// given a particular sweep line
func leftBreakPoint(arc *Beachsection, directrix float64) float64 {
	site := arc.site
	rfocx := site.X
	rfocy := site.Y
	pby2 := rfocy - directrix
	// parabola in degenerate case where focus is on directrix
	if pby2 == 0 {
		return rfocx
	}

	lArc := arc.getNode().previous
	if lArc == nil {
		return math.Inf(-1)
	}
	site = lArc.value.(*Beachsection).site
	lfocx := site.X
	lfocy := site.Y
	plby2 := lfocy - directrix
	// parabola in degenerate case where focus is on directrix
	if plby2 == 0 {
		return lfocx
	}
	hl := lfocx - rfocx
	aby2 := 1/pby2 - 1/plby2
	b := hl / plby2
	if aby2 != 0 {
		return (-b+math.Sqrt(b*b-2*aby2*(hl*hl/(-2*plby2)-lfocy+plby2/2+rfocy-pby2/2)))/aby2 + rfocx
	}
	// both parabolas have same distance to directrix, thus break point is midway
	return (rfocx + lfocx) / 2
}

// calculate the right break point of a particular beach section,
// given a particular directrix
func rightBreakPoint(arc *Beachsection, directrix float64) float64 {
	rArc := arc.getNode().next
	if rArc != nil {
		return leftBreakPoint(rArc.value.(*Beachsection), directrix)
	}
	site := arc.site
	if site.Y == directrix {
		return site.X
	}
	return math.Inf(1)
}

func (s *Voronoi) detachBeachsection(arc *Beachsection) {
	s.detachCircleEvent(arc)
	s.beachline.removeNode(arc.node)
}

type BeachsectionPtrs []*Beachsection

func (s *BeachsectionPtrs) appendLeft(b *Beachsection) {
	*s = append(*s, b)
	for id := len(*s) - 1; id > 0; id-- {
		(*s)[id] = (*s)[id-1]
	}
	(*s)[0] = b
}

func (s *BeachsectionPtrs) appendRight(b *Beachsection) {
	*s = append(*s, b)
}

func (s *Voronoi) removeBeachsection(beachsection *Beachsection) {
	circle := beachsection.circleEvent
	x := circle.x
	y := circle.ycenter
	vertex := Point{x, y}
	previous := beachsection.node.previous
	next := beachsection.node.next
	disappearingTransitions := BeachsectionPtrs{beachsection}
	abs_fn := math.Abs

	// remove collapsed beachsection from beachline
	s.detachBeachsection(beachsection)

	// there could be more than one empty arc at the deletion point, this
	// happens when more than two edges are linked by the same vertex,
	// so we will collect all those edges by looking up both sides of
	// the deletion point.
	// by the way, there is *always* a predecessor/successor to any collapsed
	// beach section, it's just impossible to have a collapsing first/last
	// beach sections on the beachline, since they obviously are unconstrained
	// on their left/right side.

	// look left
	lArc := previous.value.(*Beachsection)
	for lArc.circleEvent != nil &&
		abs_fn(x-lArc.circleEvent.x) < 1e-9 &&
		abs_fn(y-lArc.circleEvent.ycenter) < 1e-9 {

		previous = lArc.node.previous
		disappearingTransitions.appendLeft(lArc)
		s.detachBeachsection(lArc) // mark for reuse
		lArc = previous.value.(*Beachsection)
	}
	// even though it is not disappearing, I will also add the beach section
	// immediately to the left of the left-most collapsed beach section, for
	// convenience, since we need to refer to it later as this beach section
	// is the 'left' site of an edge for which a start point is set.
	disappearingTransitions.appendLeft(lArc)
	s.detachCircleEvent(lArc)

	// look right
	var rArc = next.value.(*Beachsection)
	for rArc.circleEvent != nil &&
		abs_fn(x-rArc.circleEvent.x) < 1e-9 &&
		abs_fn(y-rArc.circleEvent.ycenter) < 1e-9 {
		next = rArc.node.next
		disappearingTransitions.appendRight(rArc)
		s.detachBeachsection(rArc) // mark for reuse
		rArc = next.value.(*Beachsection)
	}
	// we also have to add the beach section immediately to the right of the
	// right-most collapsed beach section, since there is also a disappearing
	// transition representing an edge's start point on its left.
	disappearingTransitions.appendRight(rArc)
	s.detachCircleEvent(rArc)

	// walk through all the disappearing transitions between beach sections and
	// set the start point of their (implied) edge.
	nArcs := len(disappearingTransitions)

	for iArc := 1; iArc < nArcs; iArc++ {
		rArc = disappearingTransitions[iArc]
		lArc = disappearingTransitions[iArc-1]

		lSite := s.getCell(lArc.site)
		rSite := s.getCell(rArc.site)

		s.setEdgeStartpoint(rArc.edge, lSite, rSite, vertex)
	}

	// create a new edge as we have now a new transition between
	// two beach sections which were previously not adjacent.
	// since this edge appears as a new vertex is defined, the vertex
	// actually define an end point of the edge (relative to the site
	// on the left)
	lArc = disappearingTransitions[0]
	rArc = disappearingTransitions[nArcs-1]
	lSite := s.getCell(lArc.site)
	rSite := s.getCell(rArc.site)

	rArc.edge = s.createEdge(lSite, rSite, noVoronoiVertex, vertex)

	// create circle events if any for beach sections left in the beachline
	// adjacent to collapsed sections
	s.attachCircleEvent(lArc)
	s.attachCircleEvent(rArc)
}

func (s *Voronoi) addBeachsection(site Point) {
	x := site.X
	directrix := site.Y

	// find the left and right beach sections which will surround the newly
	// created beach section.
	// rhill 2011-06-01: This loop is one of the most often executed,
	// hence we expand in-place the comparison-against-epsilon calls.
	var lNode, rNode *rbNode
	var dxl, dxr float64
	node := s.beachline.root

	for node != nil {
		nodeBeachline := node.value.(*Beachsection)
		dxl = leftBreakPoint(nodeBeachline, directrix) - x
		// x lessThanWithEpsilon xl => falls somewhere before the left edge of the beachsection
		if dxl > 1e-9 {
			// this case should never happen
			// if (!node.rbLeft) {
			//    rNode = node.rbLeft;
			//    break;
			//    }
			node = node.left
		} else {
			dxr = x - rightBreakPoint(nodeBeachline, directrix)
			// x greaterThanWithEpsilon xr => falls somewhere after the right edge of the beachsection
			if dxr > 1e-9 {
				if node.right == nil {
					lNode = node
					break
				}
				node = node.right
			} else {
				// x equalWithEpsilon xl => falls exactly on the left edge of the beachsection
				if dxl > -1e-9 {
					lNode = node.previous
					rNode = node
				} else if dxr > -1e-9 {
					// x equalWithEpsilon xr => falls exactly on the right edge of the beachsection
					lNode = node
					rNode = node.next
					// falls exactly somewhere in the middle of the beachsection
				} else {
					lNode = node
					rNode = node
				}
				break
			}
		}
	}

	var lArc, rArc *Beachsection

	if lNode != nil {
		lArc = lNode.value.(*Beachsection)
	}
	if rNode != nil {
		rArc = rNode.value.(*Beachsection)
	}

	// at this point, keep in mind that lArc and/or rArc could be
	// undefined or null.

	// create a new beach section object for the site and add it to RB-tree
	newArc := &Beachsection{site: site}
	if lArc == nil {
		s.beachline.insertSuccessor(nil, newArc)
	} else {
		s.beachline.insertSuccessor(lArc.node, newArc)
	}

	// cases:
	//

	// [null,null]
	// least likely case: new beach section is the first beach section on the
	// beachline.
	// This case means:
	//   no new transition appears
	//   no collapsing beach section
	//   new beachsection become root of the RB-tree
	if lArc == nil && rArc == nil {
		return
	}

	// [lArc,rArc] where lArc == rArc
	// most likely case: new beach section split an existing beach
	// section.
	// This case means:
	//   one new transition appears
	//   the left and right beach section might be collapsing as a result
	//   two new nodes added to the RB-tree
	if lArc == rArc {
		// invalidate circle event of split beach section
		s.detachCircleEvent(lArc)

		// split the beach section into two separate beach sections
		rArc = &Beachsection{site: lArc.site}
		s.beachline.insertSuccessor(newArc.node, rArc)

		// since we have a new transition between two beach sections,
		// a new edge is born
		lCell := s.getCell(lArc.site)
		newCell := s.getCell(newArc.site)
		newArc.edge = s.createEdge(lCell, newCell, noVoronoiVertex, noVoronoiVertex)
		rArc.edge = newArc.edge

		// check whether the left and right beach sections are collapsing
		// and if so create circle events, to be notified when the point of
		// collapse is reached.
		s.attachCircleEvent(lArc)
		s.attachCircleEvent(rArc)
		return
	}

	// [lArc,null]
	// even less likely case: new beach section is the *last* beach section
	// on the beachline -- this can happen *only* if *all* the previous beach
	// sections currently on the beachline share the same y value as
	// the new beach section.
	// This case means:
	//   one new transition appears
	//   no collapsing beach section as a result
	//   new beach section become right-most node of the RB-tree
	if lArc != nil && rArc == nil {
		lCell := s.getCell(lArc.site)
		newCell := s.getCell(newArc.site)
		newArc.edge = s.createEdge(lCell, newCell, noVoronoiVertex, noVoronoiVertex)
		return
	}

	// [null,rArc]
	// impossible case: because sites are strictly processed from top to bottom,
	// and left to right, which guarantees that there will always be a beach section
	// on the left -- except of course when there are no beach section at all on
	// the beach line, which case was handled above.
	// rhill 2011-06-02: No point testing in non-debug version
	//if (!lArc && rArc) {
	//    throw "Voronoi.addBeachsection(): What is this I don't even";
	//    }

	// [lArc,rArc] where lArc != rArc
	// somewhat less likely case: new beach section falls *exactly* in between two
	// existing beach sections
	// This case means:
	//   one transition disappears
	//   two new transitions appear
	//   the left and right beach section might be collapsing as a result
	//   only one new node added to the RB-tree
	if lArc != rArc {
		// invalidate circle events of left and right sites
		s.detachCircleEvent(lArc)
		s.detachCircleEvent(rArc)

		// an existing transition disappears, meaning a vertex is defined at
		// the disappearance point.
		// since the disappearance is caused by the new beachsection, the
		// vertex is at the center of the circumscribed circle of the left,
		// new and right beachsections.
		// http://mathforum.org/library/drmath/view/55002.html
		// Except that I bring the origin at A to simplify
		// calculation
		LeftSite := lArc.site
		ax := LeftSite.X
		ay := LeftSite.Y
		bx := site.X - ax
		by := site.Y - ay
		RightSite := rArc.site
		cx := RightSite.X - ax
		cy := RightSite.Y - ay
		d := 2 * (bx*cy - by*cx)
		hb := bx*bx + by*by
		hc := cx*cx + cy*cy
		vertex := Point{(cy*hb-by*hc)/d + ax, (bx*hc-cx*hb)/d + ay}

		lCell := s.getCell(LeftSite)
		cell := s.getCell(site)
		rCell := s.getCell(RightSite)

		// one transition disappear
		s.setEdgeStartpoint(rArc.edge, lCell, rCell, vertex)

		// two new transitions appear at the new vertex location
		newArc.edge = s.createEdge(lCell, cell, noVoronoiVertex, vertex)
		rArc.edge = s.createEdge(cell, rCell, noVoronoiVertex, vertex)

		// check whether the left and right beach sections are collapsing
		// and if so create circle events, to handle the point of collapse.
		s.attachCircleEvent(lArc)
		s.attachCircleEvent(rArc)
		return
	}
}

type circleEvent struct {
	node    *rbNode
	site    Point
	arc     *Beachsection
	x       float64
	y       float64
	ycenter float64
}

func (s *circleEvent) bindToNode(node *rbNode) {
	s.node = node
}

func (s *circleEvent) getNode() *rbNode {
	return s.node
}

func (s *Voronoi) attachCircleEvent(arc *Beachsection) {
	lArc := arc.node.previous
	rArc := arc.node.next
	if lArc == nil || rArc == nil {
		return // does that ever happen?
	}
	LeftSite := lArc.value.(*Beachsection).site
	cSite := arc.site
	RightSite := rArc.value.(*Beachsection).site

	// If site of left beachsection is same as site of
	// right beachsection, there can't be convergence
	if LeftSite == RightSite {
		return
	}

	// Find the circumscribed circle for the three sites associated
	// with the beachsection triplet.
	// rhill 2011-05-26: It is more efficient to calculate in-place
	// rather than getting the resulting circumscribed circle from an
	// object returned by calling Voronoi.circumcircle()
	// http://mathforum.org/library/drmath/view/55002.html
	// Except that I bring the origin at cSite to simplify calculations.
	// The bottom-most part of the circumcircle is our Fortune 'circle
	// event', and its center is a vertex potentially part of the final
	// Voronoi diagram.
	bx := cSite.X
	by := cSite.Y
	ax := LeftSite.X - bx
	ay := LeftSite.Y - by
	cx := RightSite.X - bx
	cy := RightSite.Y - by

	// If points l->c->r are clockwise, then center beach section does not
	// collapse, hence it can't end up as a vertex (we reuse 'd' here, which
	// sign is reverse of the orientation, hence we reverse the test.
	// http://en.wikipedia.org/wiki/Curve_orientation#Orientation_of_a_simple_polygon
	// rhill 2011-05-21: Nasty finite precision error which caused circumcircle() to
	// return infinites: 1e-12 seems to fix the problem.
	d := 2 * (ax*cy - ay*cx)
	if d >= -2e-12 {
		return
	}

	ha := ax*ax + ay*ay
	hc := cx*cx + cy*cy
	x := (cy*ha - ay*hc) / d
	y := (ax*hc - cx*ha) / d
	ycenter := y + by

	// Important: ybottom should always be under or at sweep, so no need
	// to waste CPU cycles by checking

	// recycle circle event object if possible
	circleEventInst := &circleEvent{
		arc:     arc,
		site:    cSite,
		x:       x + bx,
		y:       ycenter + math.Sqrt(x*x+y*y),
		ycenter: ycenter,
	}

	arc.circleEvent = circleEventInst

	// find insertion point in RB-tree: circle events are ordered from
	// smallest to largest
	var predecessor *rbNode = nil
	node := s.circleEvents.root
	for node != nil {
		nodeValue := node.value.(*circleEvent)
		if circleEventInst.y < nodeValue.y || (circleEventInst.y == nodeValue.y && circleEventInst.x <= nodeValue.x) {
			if node.left != nil {
				node = node.left
			} else {
				predecessor = node.previous
				break
			}
		} else {
			if node.right != nil {
				node = node.right
			} else {
				predecessor = node
				break
			}
		}
	}
	s.circleEvents.insertSuccessor(predecessor, circleEventInst)
	if predecessor == nil {
		s.firstCircleEvent = circleEventInst
	}
}

func (s *Voronoi) detachCircleEvent(arc *Beachsection) {
	circle := arc.circleEvent
	if circle != nil {
		if circle.node.previous == nil {
			if circle.node.next != nil {
				s.firstCircleEvent = circle.node.next.value.(*circleEvent)
			} else {
				s.firstCircleEvent = nil
			}
		}
		s.circleEvents.removeNode(circle.node) // remove from RB-tree
		arc.circleEvent = nil
	}
}

// Bounding Box
type BBox struct {
	Xl, Xr, Yt, Yb float64
}

// Create new Bounding Box
func NewBBox(xl, xr, yt, yb float64) BBox {
	return BBox{xl, xr, yt, yb}
}

// connect dangling edges (not if a cursory test tells us
// it is not going to be visible.
// return value:
//
//	false: the dangling endpoint couldn't be connected
//	true: the dangling endpoint could be connected
func connectEdge(edge *Edge, bbox BBox) bool {
	// skip if end point already connected
	vb := edge.Vb.Point
	if vb != noVoronoiVertex {
		return true
	}

	// make local copy for performance purpose
	va := edge.Va.Point
	xl := bbox.Xl
	xr := bbox.Xr
	yt := bbox.Yt
	yb := bbox.Yb
	LeftSite := edge.LeftCell.Site
	RightSite := edge.RightCell.Site
	lx := LeftSite.X
	ly := LeftSite.Y
	rx := RightSite.X
	ry := RightSite.Y
	fx := (lx + rx) / 2
	fy := (ly + ry) / 2

	var fm, fb float64

	// get the line equation of the bisector if line is not vertical
	if !equalWithEpsilon(ry, ly) {
		fm = (lx - rx) / (ry - ly)
		fb = fy - fm*fx
	}

	// remember, direction of line (relative to left site):
	// upward: left.X < right.X
	// downward: left.X > right.X
	// horizontal: left.X == right.X
	// upward: left.X < right.X
	// rightward: left.Y < right.Y
	// leftward: left.Y > right.Y
	// vertical: left.Y == right.Y

	// depending on the direction, find the best side of the
	// bounding box to use to determine a reasonable start point

	// special case: vertical line
	if equalWithEpsilon(ry, ly) {
		// doesn't intersect with viewport
		if fx < xl || fx >= xr {
			return false
		}
		// downward
		if lx > rx {
			if va == noVoronoiVertex {
				va = Point{fx, yt}
			} else if va.Y >= yb {
				return false
			}
			vb = Point{fx, yb}
			// upward
		} else {
			if va == noVoronoiVertex {
				va = Point{fx, yb}
			} else if va.Y < yt {
				return false
			}
			vb = Point{fx, yt}
		}
		// closer to vertical than horizontal, connect start point to the
		// top or bottom side of the bounding box
	} else if fm < -1 || fm > 1 {
		// downward
		if lx > rx {
			if va == noVoronoiVertex {
				va = Point{(yt - fb) / fm, yt}
			} else if va.Y >= yb {
				return false
			}
			vb = Point{(yb - fb) / fm, yb}
			// upward
		} else {
			if va == noVoronoiVertex {
				va = Point{(yb - fb) / fm, yb}
			} else if va.Y < yt {
				return false
			}
			vb = Point{(yt - fb) / fm, yt}
		}
		// closer to horizontal than vertical, connect start point to the
		// left or right side of the bounding box
	} else {
		// rightward
		if ly < ry {
			if va == noVoronoiVertex {
				va = Point{xl, fm*xl + fb}
			} else if va.X >= xr {
				return false
			}
			vb = Point{xr, fm*xr + fb}
			// leftward
		} else {
			if va == noVoronoiVertex {
				va = Point{xr, fm*xr + fb}
			} else if va.X < xl {
				return false
			}
			vb = Point{xl, fm*xl + fb}
		}
	}
	edge.Va.Point = va
	edge.Vb.Point = vb
	return true
}

// line-clipping code taken from:
//
//	Liang-Barsky function by Daniel White
//	http://www.skytopia.com/project/articles/compsci/clipping.html
//
// Thanks!
// A bit modified to minimize code paths
func clipEdge(edge *Edge, bbox BBox) bool {
	ax := edge.Va.X
	ay := edge.Va.Y
	bx := edge.Vb.X
	by := edge.Vb.Y
	t0 := float64(0)
	t1 := float64(1)
	dx := bx - ax
	dy := by - ay

	// left
	q := ax - bbox.Xl
	if dx == 0 && q < 0 {
		return false
	}
	r := -q / dx
	if dx < 0 {
		if r < t0 {
			return false
		} else if r < t1 {
			t1 = r
		}
	} else if dx > 0 {
		if r > t1 {
			return false
		} else if r > t0 {
			t0 = r
		}
	}
	// right
	q = bbox.Xr - ax
	if dx == 0 && q < 0 {
		return false
	}
	r = q / dx
	if dx < 0 {
		if r > t1 {
			return false
		} else if r > t0 {
			t0 = r
		}
	} else if dx > 0 {
		if r < t0 {
			return false
		} else if r < t1 {
			t1 = r
		}
	}

	// top
	q = ay - bbox.Yt
	if dy == 0 && q < 0 {
		return false
	}
	r = -q / dy
	if dy < 0 {
		if r < t0 {
			return false
		} else if r < t1 {
			t1 = r
		}
	} else if dy > 0 {
		if r > t1 {
			return false
		} else if r > t0 {
			t0 = r
		}
	}
	// bottom
	q = bbox.Yb - ay
	if dy == 0 && q < 0 {
		return false
	}
	r = q / dy
	if dy < 0 {
		if r > t1 {
			return false
		} else if r > t0 {
			t0 = r
		}
	} else if dy > 0 {
		if r < t0 {
			return false
		} else if r < t1 {
			t1 = r
		}
	}

	// if we reach this point, Voronoi edge is within bbox

	// if t0 > 0, va needs to change
	// rhill 2011-06-03: we need to create a new vertex rather
	// than modifying the existing one, since the existing
	// one is likely shared with at least another edge
	if t0 > 0 {
		edge.Va.Point = Point{ax + t0*dx, ay + t0*dy}
	}

	// if t1 < 1, vb needs to change
	// rhill 2011-06-03: we need to create a new vertex rather
	// than modifying the existing one, since the existing
	// one is likely shared with at least another edge
	if t1 < 1 {
		edge.Vb.Point = Point{ax + t1*dx, ay + t1*dy}
	}

	return true
}

func equalWithEpsilon(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func lessThanWithEpsilon(a, b float64) bool {
	return b-a > 1e-9
}

func greaterThanWithEpsilon(a, b float64) bool {
	return a-b > 1e-9
}

// Connect/cut edges at bounding box
func (s *Voronoi) clipEdges(bbox BBox) {
	// connect all dangling edges to bounding box
	// or get rid of them if it can't be done
	abs_fn := math.Abs

	// iterate backward so we can splice safely
	for i := len(s.edges) - 1; i >= 0; i-- {
		edge := s.edges[i]
		// edge is removed if:
		//   it is wholly outside the bounding box
		//   it is actually a point rather than a line
		if !connectEdge(edge, bbox) || !clipEdge(edge, bbox) || (abs_fn(edge.Va.X-edge.Vb.X) < 1e-9 && abs_fn(edge.Va.Y-edge.Vb.Y) < 1e-9) {
			edge.Va.Point = noVoronoiVertex
			edge.Vb.Point = noVoronoiVertex
			s.edges[i] = s.edges[len(s.edges)-1]
			s.edges = s.edges[0 : len(s.edges)-1]
		}
	}
}

func (s *Voronoi) closeCells(bbox BBox) {
	// prune, order halfedges, then add missing ones
	// required to close cells
	xl := bbox.Xl
	xr := bbox.Xr
	yt := bbox.Yt
	yb := bbox.Yb
	cells := s.cells
	abs_fn := math.Abs

	for _, cell := range cells {
		// trim non fully-defined halfedges and sort them counterclockwise
		if cell.prepare() == 0 {
			continue
		}

		// close open cells
		// step 1: find first 'unclosed' point, if any.
		// an 'unclosed' point will be the end point of a halfedge which
		// does not match the start point of the following halfedge
		halfedges := cell.Halfedges
		nHalfedges := len(halfedges)

		// special case: only one site, in which case, the viewport is the cell
		// ...
		// all other cases
		iLeft := 0
		for iLeft < nHalfedges {
			iRight := (iLeft + 1) % nHalfedges
			endpoint := halfedges[iLeft].GetEndpoint()
			startpoint := halfedges[iRight].GetStartpoint()
			// if end point is not equal to start point, we need to add the missing
			// halfedge(s) to close the cell
			if abs_fn(endpoint.X-startpoint.X) >= 1e-9 || abs_fn(endpoint.Y-startpoint.Y) >= 1e-9 {
				// if we reach this point, cell needs to be closed by walking
				// counterclockwise along the bounding box until it connects
				// to next halfedge in the list
				va := endpoint
				vb := endpoint
				// walk downward along left side
				if equalWithEpsilon(endpoint.X, xl) && lessThanWithEpsilon(endpoint.Y, yb) {
					if equalWithEpsilon(startpoint.X, xl) {
						vb = Point{xl, startpoint.Y}
					} else {
						vb = Point{xl, yb}
					}

					// walk rightward along bottom side
				} else if equalWithEpsilon(endpoint.Y, yb) && lessThanWithEpsilon(endpoint.X, xr) {
					if equalWithEpsilon(startpoint.Y, yb) {
						vb = Point{startpoint.X, yb}
					} else {
						vb = Point{xr, yb}
					}
					// walk upward along right side
				} else if equalWithEpsilon(endpoint.X, xr) && greaterThanWithEpsilon(endpoint.Y, yt) {
					if equalWithEpsilon(startpoint.X, xr) {
						vb = Point{xr, startpoint.Y}
					} else {
						vb = Point{xr, yt}
					}
					// walk leftward along top side
				} else if equalWithEpsilon(endpoint.Y, yt) && greaterThanWithEpsilon(endpoint.X, xl) {
					if equalWithEpsilon(startpoint.Y, yt) {
						vb = Point{startpoint.X, yt}
					} else {
						vb = Point{xl, yt}
					}
				} else {
					//			break
				}

				// Create new border edge. Slide it into iLeft+1 position
				edge := s.createBorderEdge(cell, va, vb)
				cell.Halfedges = append(cell.Halfedges, nil)
				halfedges = cell.Halfedges
				nHalfedges = len(halfedges)

				copy(halfedges[iLeft+2:len(halfedges)], halfedges[iLeft+1:len(halfedges)-1])
				halfedges[iLeft+1] = newHalfedge(edge, cell, nil)

			}
			iLeft++
		}
	}
}

func (s *Voronoi) gatherVertexEdges() {
	vertexEdgeMap := make(map[Point][]*Edge)

	for _, edge := range s.edges {
		vertexEdgeMap[edge.Va.Point] = append(
			vertexEdgeMap[edge.Va.Point], edge)
		vertexEdgeMap[edge.Vb.Point] = append(
			vertexEdgeMap[edge.Vb.Point], edge)
	}

	for vertex, edgeSlice := range vertexEdgeMap {
		for _, edge := range edgeSlice {
			if vertex == edge.Va.Point {
				edge.Va.Edges = edgeSlice
			}
			if vertex == edge.Vb.Point {
				edge.Vb.Edges = edgeSlice
			}
		}
	}
}

// Compute voronoi diagram. If closeCells == true, edges from bounding box will be
// included in diagram.

// fortuneSites sorts by Y then X for deterministic sweep order.
type fortuneSites []Point

func (s fortuneSites) Len() int      { return len(s) }
func (s fortuneSites) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s fortuneSites) Less(i, j int) bool {
	if s[i].Y != s[j].Y {
		return s[i].Y < s[j].Y
	}
	return s[i].X < s[j].X
}

func computeFortuneDiagram(sites []Point, bbox BBox, closeCells bool) *fortuneDiagram {
	s := &Voronoi{
		cellsMap: make(map[Point]*Cell),
	}

	// Initialize site event queue
	fs := fortuneSites(append([]Point(nil), sites...))
	sort.Sort(fs)
	sites = []Point(fs)

	pop := func() *Point {
		if len(sites) == 0 {
			return nil
		}

		site := sites[0]
		sites = sites[1:]
		return &site
	}

	site := pop()

	// process queue
	xsitex := math.SmallestNonzeroFloat64
	xsitey := math.SmallestNonzeroFloat64
	var circle *circleEvent

	// main loop
	for {
		// we need to figure whether we handle a site or circle event
		// for this we find out if there is a site event and it is
		// 'earlier' than the circle event
		circle = s.firstCircleEvent

		// add beach section
		if site != nil && (circle == nil || site.Y < circle.y || (site.Y == circle.y && site.X < circle.x)) {
			// only if site is not a duplicate
			if site.X != xsitex || site.Y != xsitey {
				// first create cell for new site
				nCell := newCell(*site)
				s.cells = append(s.cells, nCell)
				s.cellsMap[*site] = nCell
				// then create a beachsection for that site
				s.addBeachsection(*site)
				// remember last site coords to detect duplicate
				xsitey = site.Y
				xsitex = site.X
			}
			site = pop()
			// remove beach section
		} else if circle != nil {
			s.removeBeachsection(circle.arc)
			// all done, quit
		} else {
			break
		}
	}

	// wrapping-up:
	//   connect dangling edges to bounding box
	//   cut edges as per bounding box
	//   discard edges completely outside bounding box
	//   discard edges which are point-like
	s.clipEdges(bbox)

	//   add missing edges in order to close opened cells
	if closeCells {
		s.closeCells(bbox)
	} else {
		for _, cell := range s.cells {
			cell.prepare()
		}
	}

	s.gatherVertexEdges()

	result := &fortuneDiagram{
		Edges: s.edges,
		Cells: s.cells,
	}
	return result
}

// pointKey is an exact float64 pair for map keys (duplicate site coordinates).
type pointKey struct {
	x, y float64
}

func uniqueSitesFortune(sites []Point) []Point {
	seen := make(map[pointKey]struct{}, len(sites))
	out := make([]Point, 0, len(sites))
	for _, p := range sites {
		k := pointKey{p.X, p.Y}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, p)
	}
	return out
}

// rectToBBox maps [gaul.Rect] to the Fortune clipper's bbox. The rhill/pzsz
// convention names Yt/Yb by value (not English "top"): Yt is the smaller Y,
// Yb is the larger Y (see their NewBBox(xl, xr, yt, yb) tests).
func rectToBBox(r Rect) BBox {
	return BBox{
		Xl: r.X,
		Xr: r.X + r.W,
		Yt: r.Y,
		Yb: r.Y + r.H,
	}
}

func voronoiCellPolygon(c *Cell) []Point {
	n := len(c.Halfedges)
	if n == 0 {
		return nil
	}
	pts := make([]Point, 0, n)
	for i := 0; i < n; i++ {
		p := c.Halfedges[i].GetStartpoint()
		if len(pts) == 0 || !voronoiPointEqual(pts[len(pts)-1], p) {
			pts = append(pts, p)
		}
	}
	if len(pts) >= 2 && voronoiPointEqual(pts[0], pts[len(pts)-1]) {
		pts = pts[:len(pts)-1]
	}
	return voronoiEnsurePolygonCCW(pts)
}

// voronoiPolygonSignedArea2 is twice the signed area (shoelace sum) of a simple polygon.
// Positive means counterclockwise when +Y is up (standard Cartesian).
func voronoiPolygonSignedArea2(pts []Point) float64 {
	var s float64
	n := len(pts)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		s += pts[i].X*pts[j].Y - pts[j].X*pts[i].Y
	}
	return s
}

// voronoiEnsurePolygonCCW reverses pts in place if the winding is clockwise, so that
// the polygon is counterclockwise. Degenerate or near-zero-area polygons are unchanged.
func voronoiEnsurePolygonCCW(pts []Point) []Point {
	if len(pts) < 3 {
		return pts
	}
	if voronoiPolygonSignedArea2(pts) >= -Smol {
		return pts
	}
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
	return pts
}

func voronoiPointEqual(a, b Point) bool {
	return Equalf(a.X, b.X) && Equalf(a.Y, b.Y)
}

// voronoiPointInOrOnConvexCCW reports whether p lies inside or on the boundary of a
// simple convex polygon with counterclockwise winding.
func voronoiPointInOrOnConvexCCW(p Point, poly []Point) bool {
	n := len(poly)
	if n < 3 {
		return false
	}
	for i := 0; i < n; i++ {
		a := poly[i]
		b := poly[(i+1)%n]
		if orient2(a, b, p) < -Smol {
			return false
		}
	}
	return true
}

// voronoiIsConvexCCW returns true if poly has CCW winding and every vertex is a left
// turn (no interior angle greater than 180°).
func voronoiIsConvexCCW(poly []Point) bool {
	n := len(poly)
	if n < 3 {
		return false
	}
	for i := 0; i < n; i++ {
		if orient2(poly[i], poly[(i+1)%n], poly[(i+2)%n]) < -Smol {
			return false
		}
	}
	return voronoiPolygonSignedArea2(poly) > Smol
}

// voronoiIntersectSegmentLine returns the intersection of the closed segment s–e with
// the infinite line through la–lb. ok is false when the segment is parallel to the line.
func voronoiIntersectSegmentLine(s, e, la, lb Point) (Point, bool) {
	den := orient2(la, lb, e) - orient2(la, lb, s)
	if math.Abs(den) < 1e-14 {
		return Point{}, false
	}
	t := -orient2(la, lb, s) / den
	if t < -Smol || t > 1+Smol {
		return Point{}, false
	}
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return Point{
		X: s.X + t*(e.X-s.X),
		Y: s.Y + t*(e.Y-s.Y),
	}, true
}

func voronoiInsideClipHalfPlane(clipA, clipB, p Point) bool {
	return orient2(clipA, clipB, p) >= -Smol
}

// voronoiSutherlandHodgman clips subject (closed polygon vertices, no repeated first
// point) to the convex clip polygon clip (same convention). clip must be CCW.
func voronoiSutherlandHodgman(subject []Point, clip []Point) []Point {
	if len(subject) < 1 || len(clip) < 3 {
		return nil
	}
	out := append([]Point(nil), subject...)
	nClip := len(clip)
	for i := 0; i < nClip; i++ {
		ca := clip[i]
		cb := clip[(i+1)%nClip]
		if len(out) == 0 {
			return nil
		}
		var next []Point
		prev := out[len(out)-1]
		prevIn := voronoiInsideClipHalfPlane(ca, cb, prev)
		for j := 0; j < len(out); j++ {
			curr := out[j]
			currIn := voronoiInsideClipHalfPlane(ca, cb, curr)
			if currIn {
				if !prevIn {
					if ip, ok := voronoiIntersectSegmentLine(prev, curr, ca, cb); ok {
						next = append(next, ip)
					}
				}
				next = append(next, curr)
			} else if prevIn {
				if ip, ok := voronoiIntersectSegmentLine(prev, curr, ca, cb); ok {
					next = append(next, ip)
				}
			}
			prev, prevIn = curr, currIn
		}
		out = voronoiDedupeConsecutivePolygonVerts(next)
	}
	return out
}

func voronoiDedupeConsecutivePolygonVerts(pts []Point) []Point {
	if len(pts) == 0 {
		return nil
	}
	out := make([]Point, 0, len(pts))
	for _, p := range pts {
		if len(out) > 0 && voronoiPointEqual(out[len(out)-1], p) {
			continue
		}
		out = append(out, p)
	}
	n := len(out)
	if n >= 2 && voronoiPointEqual(out[0], out[n-1]) {
		out = out[:n-1]
	}
	return out
}

// VoronoiWithRect computes the Euclidean Voronoi diagram for sites clipped to bounds
// using Fortune's sweep line algorithm (O(n log n)). Each returned Curve is closed and
// corresponds to the Voronoi cell of sites[i] intersected with bounds (a polygon, not
// a triangle mesh). Vertices are ordered counterclockwise for positive signed area in
// a Y-up coordinate system. Duplicate input coordinates yield identical curves. Sites
// must lie strictly inside or on bounds (see [Rect.ContainsPoint]).
func VoronoiWithRect(bounds Rect, sites []Point) ([]Curve, error) {
	if bounds.W <= 0 || bounds.H <= 0 {
		return nil, errors.New("gaul VoronoiWithRect: bounds width and height must be positive")
	}
	if len(sites) == 0 {
		return nil, nil
	}
	for i, p := range sites {
		if !bounds.ContainsPoint(p) {
			return nil, fmt.Errorf("gaul VoronoiWithRect: site %d is not inside bounds", i)
		}
	}
	unique := uniqueSitesFortune(sites)
	bbox := rectToBBox(bounds)
	d := computeFortuneDiagram(unique, bbox, true)

	bySite := make(map[pointKey]Curve, len(d.Cells))
	for _, c := range d.Cells {
		pts := voronoiCellPolygon(c)
		var curve Curve
		curve.Closed = true
		if len(pts) == 0 {
			curve.Points = nil
		} else {
			curve.Points = append(curve.Points, pts...)
		}
		k := pointKey{c.Site.X, c.Site.Y}
		bySite[k] = curve
	}

	// One site: the sweep leaves no finite bisectors; the clipped cell is the full bounds.
	if len(unique) == 1 {
		full := bounds.ToCurve()
		voronoiEnsurePolygonCCW(full.Points)
		bySite[pointKey{unique[0].X, unique[0].Y}] = full
	}

	out := make([]Curve, len(sites))
	for i, p := range sites {
		out[i] = bySite[pointKey{p.X, p.Y}]
	}
	return out, nil
}

// VoronoiWithCurve computes the Euclidean Voronoi diagram for sites clipped to a convex
// polygonal boundary using Fortune's algorithm on the boundary's axis-aligned bounding
// box, followed by Sutherland–Hodgman clipping of each cell to boundary. Each returned
// [Curve] is closed and corresponds to the cell of sites[i] intersected with boundary.
// Vertices are counterclockwise for positive signed area in a Y-up coordinate system.
//
// boundary must be [Curve.Closed] with at least three vertices, define a strictly
// positive area, and form a simple convex polygon (for example a [Rect] from
// [Rect.ToCurve], a triangle, or another Voronoi cell). Non-convex boundaries are not
// supported because clipping is O(segments) per cell via convex half-plane cuts.
// Duplicate site coordinates yield identical curves. Sites must lie inside or on the
// boundary polygon.
func VoronoiWithCurve(boundary Curve, sites []Point) ([]Curve, error) {
	if !boundary.Closed {
		return nil, errors.New("gaul VoronoiWithCurve: boundary curve must be closed")
	}
	if len(boundary.Points) < 3 {
		return nil, errors.New("gaul VoronoiWithCurve: boundary must have at least three points")
	}
	br := boundary.Boundary()
	if br.W <= 0 || br.H <= 0 {
		return nil, errors.New("gaul VoronoiWithCurve: boundary has empty axis-aligned extent")
	}
	clipPoly := append([]Point(nil), boundary.Points...)
	voronoiEnsurePolygonCCW(clipPoly)
	if !voronoiIsConvexCCW(clipPoly) {
		return nil, errors.New("gaul VoronoiWithCurve: boundary must be a simple convex polygon")
	}
	if len(sites) == 0 {
		return nil, nil
	}
	for i, p := range sites {
		if !voronoiPointInOrOnConvexCCW(p, clipPoly) {
			return nil, fmt.Errorf("gaul VoronoiWithCurve: site %d is not inside boundary", i)
		}
	}
	unique := uniqueSitesFortune(sites)
	bbox := rectToBBox(br)
	d := computeFortuneDiagram(unique, bbox, true)

	bySite := make(map[pointKey]Curve, len(d.Cells))
	for _, c := range d.Cells {
		pts := voronoiCellPolygon(c)
		clipped := voronoiSutherlandHodgman(pts, clipPoly)
		var curve Curve
		curve.Closed = true
		if len(clipped) < 3 {
			curve.Points = nil
		} else {
			curve.Points = append(curve.Points, clipped...)
			voronoiEnsurePolygonCCW(curve.Points)
		}
		k := pointKey{c.Site.X, c.Site.Y}
		bySite[k] = curve
	}
	if len(unique) == 1 {
		full := Curve{
			Closed: true,
			Points: append([]Point(nil), clipPoly...),
		}
		voronoiEnsurePolygonCCW(full.Points)
		bySite[pointKey{unique[0].X, unique[0].Y}] = full
	}

	out := make([]Curve, len(sites))
	for i, p := range sites {
		out[i] = bySite[pointKey{p.X, p.Y}]
	}
	return out, nil
}
