# Generative Art Utility Library (gaul)

Gaul is a generative art library. It was originally part of [Sketchy](https://github.com/aldernero/sketchy), but I
separated the non-gui parts and put them in gaul. The main reason for this is
to make it easy to containerize generative art applications.

## Rendering

Gaul ships its own small 2D renderer in the `render` subpackage, with two
backends behind one drawing API:

- **Raster** — anti-aliased rasterizer (built on `freetype/raster`) writing PNG
- **SVG** — native vector output with true stroked paths and beziers, suitable
  for pen plotting

Coordinates are always pixels: origin top-left, x right, y down.

A `render.Context` fans out to any number of backends, so one drawing pass can
produce both a PNG and an SVG:

```go
ras := render.NewRaster(800, 600)
svg := render.NewSVG(800, 600)
ctx := render.NewContext(ras, svg)
// ... draw ...
ras.SavePNG("out.png")
svg.Save("out.svg")
```

Two drawing styles are supported and can be mixed freely.

Processing-style immediate mode, using the context's transform stack:

```go
ctx.SetFillColor(color.White)
ctx.Push()
ctx.Translate(400, 300)
ctx.Rotate(gaul.Pi / 4)
ctx.DrawRectangle(-50, -50, 100, 100)
ctx.Fill()
ctx.Pop()
```

Primitive-first, where geometry is constructed and transformed independently
(e.g. with `Affine2D`) and only drawn at the last minute:

```go
tri := gaul.Triangle{A: a, B: b, C: c}
xform := gaul.Mult(
    gaul.NewAffine2DWithTranslation(400, 300),
    gaul.NewAffine2DWithRotation(gaul.Pi/4),
)
curve := xform.TransformCurve(tri.ToCurve())
curve.Draw(ctx)
```
