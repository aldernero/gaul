// Package render provides a small 2D renderer for gaul with two backends: a
// rasterizer (writing PNG) built on github.com/golang/freetype/raster, and an
// SVG writer suitable for pen plotting.
//
// Coordinates are always pixels, with the origin at the top-left corner, x
// increasing to the right and y increasing down. There are no physical units
// and no y-axis flips anywhere.
//
// Two drawing styles are supported, and can be freely mixed:
//
// Processing-style immediate mode, where the Context carries a transform
// stack and drawing state:
//
//	r := render.NewRaster(800, 600)
//	ctx := render.NewContext(r)
//	ctx.SetFillColor(color.White)
//	ctx.Push()
//	ctx.Translate(400, 300)
//	ctx.Rotate(gaul.Pi / 4)
//	ctx.DrawRectangle(-50, -50, 100, 100)
//	ctx.Fill()
//	ctx.Pop()
//	r.SavePNG("out.png")
//
// Primitive-first, where geometry is constructed and transformed
// independently (e.g. with gaul's Affine2D) and the Context is only used at
// draw time:
//
//	c := gaul.Circle{Center: gaul.Point{X: 400, Y: 300}, Radius: 100}
//	c.Draw(ctx)
//
// A Context fans out to any number of Renderers, so a single drawing pass can
// produce both a PNG and an SVG:
//
//	r := render.NewRaster(800, 600)
//	s := render.NewSVG(800, 600)
//	ctx := render.NewContext(r, s)
//	// ... draw ...
//	r.SavePNG("out.png")
//	s.Save("out.svg")
package render

import (
	"image"
	"image/color"
)

// Renderer consumes fully-transformed, device-space draw operations. All
// coordinates handed to a Renderer are in pixels, origin top-left, y down;
// the Context applies its transform stack before calling these methods, so
// implementations hold no transform state of their own.
type Renderer interface {
	// Size returns the drawing surface dimensions in pixels.
	Size() (w, h float64)
	// RenderPath draws p, already in device coordinates, with style s. The
	// renderer must not retain or mutate p after returning.
	RenderPath(p *Path, s Style)
	// RenderImage draws img transformed by m (device coordinates).
	RenderImage(img image.Image, m Matrix)
	// RenderText draws a single line of text with its baseline-left anchor
	// at (x, y) in device coordinates.
	RenderText(text string, x, y float64, f *Font, sizePx float64, c color.Color)
}
