package render

import (
	"bufio"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"

	ftraster "github.com/golang/freetype/raster"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
)

// Raster is a Renderer that rasterizes into an *image.RGBA using
// github.com/golang/freetype/raster, with anti-aliased fills and strokes.
type Raster struct {
	img     *image.RGBA
	rz      *ftraster.Rasterizer
	painter *ftraster.RGBAPainter
}

// NewRaster returns a Raster with a new w x h image.
func NewRaster(w, h int) *Raster {
	return NewRasterFromImage(image.NewRGBA(image.Rect(0, 0, w, h)))
}

// NewRasterFromImage returns a Raster that draws into the caller-owned
// image without copying, so pixels land directly in img. This suits
// render loops that reuse one buffer every frame. Device coordinate (0, 0)
// maps to the image's Bounds().Min corner.
func NewRasterFromImage(img *image.RGBA) *Raster {
	b := img.Bounds()
	rz := ftraster.NewRasterizer(b.Dx(), b.Dy())
	rz.Dx, rz.Dy = b.Min.X, b.Min.Y
	return &Raster{
		img:     img,
		rz:      rz,
		painter: ftraster.NewRGBAPainter(img),
	}
}

// Image returns the underlying image.
func (r *Raster) Image() *image.RGBA { return r.img }

// Bounds returns the underlying image bounds.
func (r *Raster) Bounds() image.Rectangle { return r.img.Bounds() }

// At returns the pixel at (x, y) in image coordinates.
func (r *Raster) At(x, y int) color.Color { return r.img.At(x, y) }

// SetPixel sets the pixel at (x, y) in image coordinates.
func (r *Raster) SetPixel(x, y int, c color.Color) { r.img.Set(x, y, c) }

// SavePNG writes the image to a PNG file.
func (r *Raster) SavePNG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if err := png.Encode(w, r.img); err != nil {
		_ = f.Close()
		return err
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// EncodePNG writes the image as PNG to w.
func (r *Raster) EncodePNG(w io.Writer) error {
	return png.Encode(w, r.img)
}

// Size implements Renderer.
func (r *Raster) Size() (float64, float64) {
	b := r.img.Bounds()
	return float64(b.Dx()), float64(b.Dy())
}

// RenderPath implements Renderer.
func (r *Raster) RenderPath(p *Path, s Style) {
	fill := s.hasFill()
	stroke := s.hasStroke()
	if !fill && !stroke {
		return
	}
	polys, closed := p.Flatten()
	if len(polys) == 0 {
		return
	}
	if fill {
		var rp ftraster.Path
		for _, poly := range polys {
			rp.Start(fixPoint(poly[0], poly[1]))
			for i := 2; i < len(poly); i += 2 {
				rp.Add1(fixPoint(poly[i], poly[i+1]))
			}
			// Filling implicitly closes every subpath.
			rp.Add1(fixPoint(poly[0], poly[1]))
		}
		r.rz.Clear()
		r.rz.UseNonZeroWinding = s.FillRule == NonZero
		r.rz.AddPath(rp)
		r.painter.SetColor(s.FillColor)
		r.rz.Rasterize(r.painter)
	}
	if stroke {
		var strokePolys [][]float64
		if len(s.Dashes) > 0 {
			strokePolys = dashPolylines(polys, closed, s.Dashes, s.DashOffset)
		} else {
			strokePolys = closePolylines(polys, closed)
		}
		if len(strokePolys) == 0 {
			return
		}
		var rp ftraster.Path
		for _, poly := range strokePolys {
			rp.Start(fixPoint(poly[0], poly[1]))
			for i := 2; i < len(poly); i += 2 {
				rp.Add1(fixPoint(poly[i], poly[i+1]))
			}
		}
		r.rz.Clear()
		// Stroke outlines self-overlap at joins; non-zero winding keeps the
		// overlap from punching holes.
		r.rz.UseNonZeroWinding = true
		r.rz.AddStroke(rp, fix(s.StrokeWidth), capper(s.Cap), joiner(s.Join))
		r.painter.SetColor(s.StrokeColor)
		r.rz.Rasterize(r.painter)
	}
}

// RenderImage implements Renderer.
func (r *Raster) RenderImage(img image.Image, m Matrix) {
	b := img.Bounds()
	if dx, dy, ok := m.isIntTranslation(); ok {
		dst := image.Rect(dx, dy, dx+b.Dx(), dy+b.Dy()).Add(r.img.Bounds().Min)
		draw.Draw(r.img, dst, img, b.Min, draw.Over)
		return
	}
	// The matrix maps the image's top-left corner to the device position, so
	// shift the source bounds origin to (0, 0) first and device (0, 0) to
	// the destination's Bounds().Min.
	rmin := r.img.Bounds().Min
	dm := Translation(float64(rmin.X), float64(rmin.Y)).Mul(m).Mul(Translation(float64(-b.Min.X), float64(-b.Min.Y)))
	xdraw.BiLinear.Transform(r.img, f64.Aff3{dm.A, dm.C, dm.E, dm.B, dm.D, dm.F}, img, b, xdraw.Over, nil)
}

// RenderText implements Renderer. Glyphs render upright at the given anchor;
// the Context's transform applies to the anchor point only, not glyph shapes.
func (r *Raster) RenderText(text string, x, y float64, f *Font, sizePx float64, c color.Color) {
	if f == nil || text == "" || !visible(c) {
		return
	}
	face, err := f.face(sizePx)
	if err != nil {
		return
	}
	d := font.Drawer{
		Dst:  r.img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.Point26_6{X: fix(x), Y: fix(y)},
	}
	d.DrawString(text)
}

func fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(math.Round(x * 64))
}

func fixPoint(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{X: fix(x), Y: fix(y)}
}

func capper(c LineCap) ftraster.Capper {
	switch c {
	case RoundCap:
		return ftraster.RoundCapper
	case SquareCap:
		return ftraster.SquareCapper
	default:
		return ftraster.ButtCapper
	}
}

func joiner(j LineJoin) ftraster.Joiner {
	// freetype/raster has no miter joiner; MiterJoin falls back to bevel.
	if j == RoundJoin {
		return ftraster.RoundJoiner
	}
	return ftraster.BevelJoiner
}
