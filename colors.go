package gaul

import (
	"image/color"
	"math"
	"regexp"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"golang.org/x/image/colornames"
)

type ColorType int

const (
	BackgroundColorType = iota
	OutlineColorType
	TextColorType
	FillColorType
)

type ColorConfig struct {
	Background color.Color
	Outline    color.Color
	Text       color.Color
	Fill       color.Color
	Gradient   SimpleGradient
}

func (cc *ColorConfig) Set(hexString string, colorType ColorType, defaultString string) {
	colorHexString := defaultString
	if hexString != "" {
		colorHexString = hexString
	}
	c := StringToColor(colorHexString)
	switch colorType {
	case BackgroundColorType:
		cc.Background = c
	case OutlineColorType:
		cc.Outline = c
	case TextColorType:
		cc.Text = c
	case FillColorType:
		cc.Fill = c
	}
}

func StringToColor(colorString string) color.Color {
	if colorString == "" {
		return color.Transparent
	}
	re := regexp.MustCompile("#[0-9a-f]{6}")
	name := strings.ToLower(colorString)
	if re.MatchString(name) {
		c, err := colorful.Hex(name)
		if err != nil {
			panic(err)
		}
		return c
	}
	return NamedColor(name)
}

func NamedColor(name string) color.Color {
	val, ok := colornames.Map[strings.ToLower(name)]
	if !ok {
		panic("invalid color name")
	}
	return val
}

type SimpleGradient struct {
	StartColor color.Color
	EndColor   color.Color
}

func NewSimpleGradient(c1, c2 string) SimpleGradient {
	gradient := SimpleGradient{
		StartColor: StringToColor(c1),
		EndColor:   StringToColor(c2),
	}
	return gradient
}

func (sg *SimpleGradient) Color(percentage float64) color.Color {
	val := Clamp(0, 1, percentage)
	c1, _ := colorful.MakeColor(sg.StartColor)
	c2, _ := colorful.MakeColor(sg.EndColor)
	return c1.BlendHcl(c2, val)
}

type Gradient struct {
	stops []color.Color
}

func (g *Gradient) AddStop(colorString string) {
	g.stops = append(g.stops, StringToColor(colorString))
}

func (g *Gradient) NumStops() int {
	return len(g.stops)
}

func NewGradientFromNamed(names []string) Gradient {
	var grad Gradient
	for _, i := range names {
		grad.AddStop(i)
	}
	return grad
}

func (g *Gradient) Color(percentage float64) color.Color {
	val := Clamp(0, 1, percentage)
	n := g.NumStops()
	if n == 1 {
		return g.stops[0]
	}
	if val == 0 {
		return g.stops[0]
	}
	if val == 1 {
		return g.stops[n-1]
	}
	i := Map(0, 1, 0, float64(n)-1, percentage)
	lerp := i - math.Floor(i)
	index := int(math.Floor(i))
	c1, _ := colorful.MakeColor(g.stops[index])
	c2, _ := colorful.MakeColor(g.stops[index+1])
	return c1.BlendHcl(c2, lerp)
}

func (g *Gradient) LinearPalette(num int) []color.Color {
	palette := make([]color.Color, num)
	ls := Linspace(0, 1, num, true)
	for i, p := range ls {
		palette[i] = g.Color(p)
	}
	return palette
}

func (g *Gradient) LinearPaletteStrings(num int) []string {
	palette := g.LinearPalette(num)
	paletteStrings := make([]string, num)
	for i, c := range palette {
		d, ok := colorful.MakeColor(c)
		if !ok {
			panic("invalid color")
		}
		paletteStrings[i] = d.Hex()
	}
	return paletteStrings
}

type SinePalette struct {
	A, B, C, D Vec3
	Alpha      float64
}

func NewSinePalette(c, d Vec3) SinePalette {
	return SinePalette{
		A:     Vec3{0.5, 0.5, 0.5}, // typically don't need to change
		B:     Vec3{0.5, 0.5, 0.5}, // typically don't need to change
		C:     c,
		D:     d,
		Alpha: 1.0,
	}
}

func (sp *SinePalette) ColorAt(t float64) color.Color {
	r := sp.A.X + sp.B.X*math.Cos(2*math.Pi*(sp.C.X*t+sp.D.X))
	g := sp.A.Y + sp.B.Y*math.Cos(2*math.Pi*(sp.C.Y*t+sp.D.Y))
	b := sp.A.Z + sp.B.Z*math.Cos(2*math.Pi*(sp.C.Z*t+sp.D.Z))
	var result color.RGBA64
	result.R = uint16(Clamp(0, 1, r) * 65535)
	result.G = uint16(Clamp(0, 1, g) * 65535)
	result.B = uint16(Clamp(0, 1, b) * 65535)
	result.A = uint16(Clamp(0, 1, sp.Alpha) * 65535)
	return result
}

func (sp *SinePalette) Palette(num int) []color.Color {
	palette := make([]color.Color, num)
	ls := Linspace(0, 1, num, true)
	for i, p := range ls {
		palette[i] = sp.ColorAt(p)
	}
	return palette
}

func (sp *SinePalette) ToGridPng() (string, error) {
	w := 1024.0
	h := 1024.0
	c := canvas.New(w, h)
	ctx := canvas.NewContext(c)
	ctx.SetStrokeWidth(0.1)
	minPower := 0.0
	maxPower := math.Floor(math.Log2(h)) - 1
	y := 0.0
	dy := h / float64(maxPower)
	power := minPower
	for y <= h {
		n := math.Pow(2, power)
		x := 0.0
		dx := w / n
		for x <= w {
			k := sp.ColorAt((x + 0.5*dx) / w)
			ctx.SetFillColor(k)
			ctx.SetFillColor(k)
			ctx.DrawPath(x, y, canvas.Rectangle(dx, dy))
			ctx.FillStroke()
			x += dx
		}
		power += 1.0
		y += dy
	}
	fname := "gaul-sinepalette-grid_" + GetTimestampString() + ".png"
	if err := renderers.Write(fname, c); err != nil {
		return fname, err
	}
	return fname, nil
}

func (sp *SinePalette) ToPng() (string, error) {
	w := 285.0
	h := 55.0
	c := canvas.New(w, h)
	ctx := canvas.NewContext(c)
	fontFamily := canvas.NewFontFamily("DejaVu Sans")
	if err := fontFamily.LoadLocalFont("DejaVuSans", canvas.FontRegular); err != nil {
		return "", err
	}
	fontFace := fontFamily.Face(14.0, color.White, canvas.FontRegular, canvas.FontNormal)
	ctx.SetFillColor(canvas.Black)
	ctx.SetStrokeColor(canvas.Transparent)
	ctx.DrawPath(0, 0, canvas.Rectangle(ctx.Width(), ctx.Height()))
	ctx.Close()
	t := Linspace(0, 1, 1000, true)
	dw := 0.8 * w / 1000
	dh := 0.2 * h
	aText := "A = " + sp.A.ToString(3)
	aTextBox := canvas.NewTextBox(fontFace, aText, 0.8*w, 10, canvas.Left, canvas.Center, 0, 0)
	bText := "B = " + sp.B.ToString(3)
	bTextBox := canvas.NewTextBox(fontFace, bText, 0.8*w, 10, canvas.Right, canvas.Center, 0, 0)
	cText := "C = " + sp.C.ToString(3)
	cTextBox := canvas.NewTextBox(fontFace, cText, 0.8*w, 10, canvas.Left, canvas.Center, 0, 0)
	dText := "D = " + sp.D.ToString(3)
	dTextBox := canvas.NewTextBox(fontFace, dText, 0.8*w, 10, canvas.Right, canvas.Center, 0, 0)
	ctx.SetFillColor(color.White)
	ctx.SetStrokeColor(color.White)
	ctx.DrawText(0.1*w, h+dh+8, aTextBox)
	ctx.DrawText(0.1*w, h+dh+8, bTextBox)
	ctx.DrawText(0.1*w, h-dh, cTextBox)
	ctx.DrawText(0.1*w, h-dh, dTextBox)
	ctx.SetStrokeWidth(2 * dw)
	for i, j := range t {
		lineColor1 := sp.ColorAt(j)
		ctx.SetStrokeColor(lineColor1)
		x := 0.1*w + float64(i)*dw
		ctx.MoveTo(x, h-dh)
		ctx.LineTo(x, h+dh)
		ctx.Stroke()
	}
	fname := "gaul-sinepalette_" + GetTimestampString() + ".png"
	if err := renderers.Write(fname, c); err != nil {
		return "", err
	}
	return fname, nil
}
