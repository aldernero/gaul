package render

import (
	"image"
	"image/color"
)

type opKind uint8

const (
	opPath opKind = iota
	opImage
	opText
)

type recordedOp struct {
	kind  opKind
	path  *Path
	style Style
	img   image.Image
	m     Matrix
	text  string
	x, y  float64
	font  *Font
	size  float64
	col   color.Color
}

// Recorder is a Renderer that records draw operations so they can be
// replayed later into other renderers. It preserves a frame exactly as it
// was drawn: attach a Recorder alongside a live Raster during drawing, then
// Replay the recording into a fresh Raster (at any scale) or an SVG to
// export the same frame without re-running the drawing code.
//
// Recorded paths are deep-copied; images and fonts are retained by
// reference and must not be mutated before Replay.
type Recorder struct {
	w, h float64
	ops  []recordedOp
}

// NewRecorder returns a Recorder with the given logical surface size.
func NewRecorder(w, h float64) *Recorder {
	return &Recorder{w: w, h: h}
}

// Size implements Renderer.
func (r *Recorder) Size() (float64, float64) { return r.w, r.h }

// Empty reports whether nothing has been recorded.
func (r *Recorder) Empty() bool { return len(r.ops) == 0 }

// Reset discards all recorded operations, keeping the backing array. The
// cleared entries are zeroed so the recording does not pin paths, images, and
// fonts alive in the retained array.
func (r *Recorder) Reset() {
	clear(r.ops)
	r.ops = r.ops[:0]
}

// RenderPath implements Renderer.
func (r *Recorder) RenderPath(p *Path, s Style) {
	s.Dashes = append([]float64(nil), s.Dashes...)
	r.ops = append(r.ops, recordedOp{kind: opPath, path: p.Copy(), style: s})
}

// RenderImage implements Renderer.
func (r *Recorder) RenderImage(img image.Image, m Matrix) {
	r.ops = append(r.ops, recordedOp{kind: opImage, img: img, m: m})
}

// RenderText implements Renderer.
func (r *Recorder) RenderText(text string, x, y float64, f *Font, sizePx float64, c color.Color) {
	r.ops = append(r.ops, recordedOp{kind: opText, text: text, x: x, y: y, font: f, size: sizePx, col: c})
}

// Replay sends every recorded operation, in order, to each destination
// renderer. The recording is left intact and can be replayed again.
func (r *Recorder) Replay(dst ...Renderer) {
	for _, d := range dst {
		for i := range r.ops {
			op := &r.ops[i]
			switch op.kind {
			case opPath:
				d.RenderPath(op.path, op.style)
			case opImage:
				d.RenderImage(op.img, op.m)
			case opText:
				d.RenderText(op.text, op.x, op.y, op.font, op.size, op.col)
			}
		}
	}
}
