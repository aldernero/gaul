package render

// dashPolylines splits flattened polylines into the "on" runs of a dash
// pattern, returning open polylines. Each run is stroked separately so caps
// apply to every dash. Closed polylines are dashed as loops (the pattern
// runs around the perimeter without a seam reset). Follows SVG semantics:
// an odd-length pattern is repeated to make it even, and offset shifts the
// start position within the pattern.
func dashPolylines(polys [][]float64, closed []bool, dashes []float64, offset float64) [][]float64 {
	pattern := normalizeDashes(dashes)
	if pattern == nil {
		return closePolylines(polys, closed)
	}
	total := 0.0
	for _, d := range pattern {
		total += d
	}
	var out [][]float64
	for i, poly := range polys {
		pts := poly
		if closed[i] {
			pts = append(append([]float64(nil), poly...), poly[0], poly[1])
		}
		out = appendDashedPolyline(out, pts, pattern, total, offset)
	}
	return out
}

// closePolylines returns the polylines with closed ones given an explicit
// closing point, leaving open ones untouched.
func closePolylines(polys [][]float64, closed []bool) [][]float64 {
	out := make([][]float64, len(polys))
	for i, poly := range polys {
		if closed[i] {
			out[i] = append(append([]float64(nil), poly...), poly[0], poly[1])
		} else {
			out[i] = poly
		}
	}
	return out
}

// normalizeDashes validates and normalizes a dash pattern. It returns nil if
// the pattern is empty or degenerate (non-positive total), meaning solid.
func normalizeDashes(dashes []float64) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	total := 0.0
	for _, d := range dashes {
		if d < 0 {
			return nil
		}
		total += d
	}
	if total <= 0 {
		return nil
	}
	pattern := append([]float64(nil), dashes...)
	if len(pattern)%2 == 1 {
		pattern = append(pattern, pattern...)
	}
	return pattern
}

// appendDashedPolyline walks one polyline, appending its "on" runs to out.
func appendDashedPolyline(out [][]float64, pts []float64, pattern []float64, patternTotal, offset float64) [][]float64 {
	// Resolve the starting position within the pattern.
	phase := offset
	for phase < 0 {
		phase += patternTotal
	}
	for phase >= patternTotal {
		phase -= patternTotal
	}
	idx := 0
	for phase >= pattern[idx] {
		phase -= pattern[idx]
		idx = (idx + 1) % len(pattern)
	}
	remain := pattern[idx] - phase // distance left in the current element
	on := idx%2 == 0

	var run []float64
	flush := func() {
		if len(run) >= 4 {
			out = append(out, run)
		}
		run = nil
	}
	if on {
		run = append(run, pts[0], pts[1])
	}
	x0, y0 := pts[0], pts[1]
	for i := 2; i < len(pts); i += 2 {
		x1, y1 := pts[i], pts[i+1]
		segLen := dist(x0, y0, x1, y1)
		for segLen > 0 {
			if remain >= segLen {
				// Element outlasts the segment.
				remain -= segLen
				segLen = 0
				if on {
					run = append(run, x1, y1)
				}
				x0, y0 = x1, y1
				continue
			}
			// Element boundary falls inside the segment.
			t := remain / segLen
			mx := x0 + (x1-x0)*t
			my := y0 + (y1-y0)*t
			if on {
				run = append(run, mx, my)
				flush()
			} else {
				run = append(run, mx, my)
			}
			segLen -= remain
			x0, y0 = mx, my
			idx = (idx + 1) % len(pattern)
			remain = pattern[idx]
			on = !on
			if !on {
				run = nil
			}
		}
	}
	flush()
	return out
}
