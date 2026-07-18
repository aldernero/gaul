package render

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
)

// SVG is a Renderer that records draw operations as SVG elements and writes
// a complete document on demand. It is designed for pen-plotter workflows:
// strokes are emitted as real stroke attributes (never expanded to filled
// outlines) and beziers stay native Q/C path commands, so downstream tools
// (vpype, axidraw, etc.) receive true stroked curves. Because the Context
// pre-transforms all coordinates, path elements carry flat coordinates and
// no transform attributes.
type SVG struct {
	w, h  float64
	elems bytes.Buffer
	// Precision is the number of decimal places written for coordinates.
	Precision int
}

// NewSVG returns an SVG renderer with a w x h pixel viewport.
func NewSVG(w, h float64) *SVG {
	return &SVG{w: w, h: h, Precision: 3}
}

// Size implements Renderer.
func (s *SVG) Size() (float64, float64) { return s.w, s.h }

// Reset discards all recorded elements.
func (s *SVG) Reset() { s.elems.Reset() }

// WriteTo writes the complete SVG document to w.
func (s *SVG) WriteTo(w io.Writer) (int64, error) {
	var doc bytes.Buffer
	doc.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&doc, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`+"\n",
		s.fmtFloat(s.w), s.fmtFloat(s.h), s.fmtFloat(s.w), s.fmtFloat(s.h))
	doc.Write(s.elems.Bytes())
	doc.WriteString("</svg>\n")
	return doc.WriteTo(w)
}

// Save writes the complete SVG document to a file.
func (s *SVG) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if _, err := s.WriteTo(w); err != nil {
		_ = f.Close()
		return err
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// RenderPath implements Renderer.
func (s *SVG) RenderPath(p *Path, st Style) {
	if p.Empty() || (!st.hasFill() && !st.hasStroke()) {
		return
	}
	s.elems.WriteString(`<path d="`)
	p.Iterate(func(v PathVerb, pts []float64) {
		switch v {
		case VerbMoveTo:
			s.writeCommand('M', pts)
		case VerbLineTo:
			s.writeCommand('L', pts)
		case VerbQuadTo:
			s.writeCommand('Q', pts)
		case VerbCubicTo:
			s.writeCommand('C', pts)
		case VerbClose:
			s.elems.WriteString("Z")
		}
	})
	s.elems.WriteString(`"`)
	s.writeStyle(st)
	s.elems.WriteString("/>\n")
}

func (s *SVG) writeCommand(cmd byte, pts []float64) {
	s.elems.WriteByte(cmd)
	for i, v := range pts {
		if i > 0 {
			s.elems.WriteByte(' ')
		}
		s.elems.WriteString(s.fmtFloat(v))
	}
}

func (s *SVG) writeStyle(st Style) {
	if st.hasFill() {
		hex, opacity := svgColor(st.FillColor)
		s.writeAttr("fill", hex)
		if opacity < 1 {
			s.writeAttr("fill-opacity", s.fmtFloat(opacity))
		}
		if st.FillRule == EvenOdd {
			s.writeAttr("fill-rule", "evenodd")
		}
	} else {
		s.writeAttr("fill", "none")
	}
	if st.hasStroke() {
		hex, opacity := svgColor(st.StrokeColor)
		s.writeAttr("stroke", hex)
		s.writeAttr("stroke-width", s.fmtFloat(st.StrokeWidth))
		if opacity < 1 {
			s.writeAttr("stroke-opacity", s.fmtFloat(opacity))
		}
		switch st.Cap {
		case RoundCap:
			s.writeAttr("stroke-linecap", "round")
		case SquareCap:
			s.writeAttr("stroke-linecap", "square")
		}
		switch st.Join {
		case RoundJoin:
			s.writeAttr("stroke-linejoin", "round")
		case BevelJoin:
			s.writeAttr("stroke-linejoin", "bevel")
		default:
			if st.MiterLimit > 0 {
				s.writeAttr("stroke-miterlimit", s.fmtFloat(st.MiterLimit))
			}
		}
		if pattern := normalizeDashes(st.Dashes); pattern != nil {
			var dashes bytes.Buffer
			for i, d := range st.Dashes {
				if i > 0 {
					dashes.WriteByte(' ')
				}
				dashes.WriteString(s.fmtFloat(d))
			}
			s.writeAttr("stroke-dasharray", dashes.String())
			if st.DashOffset != 0 {
				s.writeAttr("stroke-dashoffset", s.fmtFloat(st.DashOffset))
			}
		}
	}
}

func (s *SVG) writeAttr(name, value string) {
	s.elems.WriteByte(' ')
	s.elems.WriteString(name)
	s.elems.WriteString(`="`)
	s.elems.WriteString(value)
	s.elems.WriteString(`"`)
}

// RenderImage implements Renderer. The image is embedded as a base64 PNG;
// this is the one element type that carries a transform attribute, since a
// raster image's transform cannot be baked into its pixels losslessly.
func (s *SVG) RenderImage(img image.Image, m Matrix) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return
	}
	b := img.Bounds()
	fmt.Fprintf(&s.elems,
		`<image width="%d" height="%d" transform="matrix(%s %s %s %s %s %s)" href="data:image/png;base64,%s"/>`+"\n",
		b.Dx(), b.Dy(),
		s.fmtFloat(m.A), s.fmtFloat(m.B), s.fmtFloat(m.C), s.fmtFloat(m.D), s.fmtFloat(m.E), s.fmtFloat(m.F),
		base64.StdEncoding.EncodeToString(buf.Bytes()))
}

// RenderText implements Renderer. Text is written as a <text> element
// referencing the font by family name; the viewer substitutes its own font
// file, so metrics can differ from raster output (see Font).
func (s *SVG) RenderText(text string, x, y float64, f *Font, sizePx float64, c color.Color) {
	if f == nil || text == "" || !visible(c) {
		return
	}
	hex, opacity := svgColor(c)
	s.elems.WriteString(`<text`)
	s.writeAttr("x", s.fmtFloat(x))
	s.writeAttr("y", s.fmtFloat(y))
	s.writeAttr("font-family", f.Name)
	s.writeAttr("font-size", s.fmtFloat(sizePx)+"px")
	s.writeAttr("fill", hex)
	if opacity < 1 {
		s.writeAttr("fill-opacity", s.fmtFloat(opacity))
	}
	s.elems.WriteString(">")
	if err := xml.EscapeText(&s.elems, []byte(text)); err != nil {
		s.elems.WriteString("</text>\n")
		return
	}
	s.elems.WriteString("</text>\n")
}

func (s *SVG) fmtFloat(v float64) string {
	prec := s.Precision
	if prec <= 0 {
		prec = 3
	}
	scale := math.Pow(10, float64(prec))
	v = math.Round(v*scale) / scale
	if v == 0 {
		v = 0 // normalize -0
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// svgColor returns the #rrggbb hex form of c and its opacity in [0, 1].
func svgColor(c color.Color) (hex string, opacity float64) {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return fmt.Sprintf("#%02x%02x%02x", n.R, n.G, n.B), float64(n.A) / 255
}
