package gaul

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"regexp"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/aldernero/gaul/render"
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

// hexColorRe matches a full 6-digit hex color. It is anchored so that strings
// merely containing a hex code are not mistaken for one, and compiled once
// rather than on every parse.
var hexColorRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// ErrInvalidColor is returned by ParseColor and LookupNamedColor for a string
// that is neither a 6-digit hex color nor a known SVG color name.
var ErrInvalidColor = errors.New("invalid color")

// ParseColor converts a "#rrggbb" hex string or an SVG color name into a
// color. An empty string yields a transparent color. It returns an error
// wrapping ErrInvalidColor if the string is neither.
func ParseColor(colorString string) (color.Color, error) {
	if colorString == "" {
		return color.Transparent, nil
	}
	name := strings.ToLower(colorString)
	if hexColorRe.MatchString(name) {
		c, err := colorful.Hex(name)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrInvalidColor, colorString, err)
		}
		return c, nil
	}
	return LookupNamedColor(name)
}

// LookupNamedColor returns the SVG color with the given name, case
// insensitively. It returns an error wrapping ErrInvalidColor if the name is
// unknown.
func LookupNamedColor(name string) (color.Color, error) {
	val, ok := colornames.Map[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("%w: unknown color name %q", ErrInvalidColor, name)
	}
	return val, nil
}

// StringToColor converts a "#rrggbb" hex string or an SVG color name into a
// color, panicking if the string is neither. Use [ParseColor] to handle
// invalid input.
func StringToColor(colorString string) color.Color {
	c, err := ParseColor(colorString)
	if err != nil {
		panic(err)
	}
	return c
}

// NamedColor returns the SVG color with the given name, panicking if the name
// is unknown. Use [LookupNamedColor] to handle invalid input.
func NamedColor(name string) color.Color {
	c, err := LookupNamedColor(name)
	if err != nil {
		panic(err)
	}
	return c
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
	// positions holds a normalized position in [0,1] for each stop. When nil,
	// stops are treated as evenly spaced (the original behavior). When set, it
	// must be the same length as stops and sorted in ascending order.
	positions []float64
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
	if n == 0 {
		return color.Transparent
	}
	if n == 1 {
		return g.stops[0]
	}
	if g.positions != nil {
		return g.colorAtPositions(val)
	}
	if val == 0 {
		return g.stops[0]
	}
	if val == 1 {
		return g.stops[n-1]
	}
	i := Map(0, 1, 0, float64(n)-1, val)
	lerp := i - math.Floor(i)
	index := int(math.Floor(i))
	c1, _ := colorful.MakeColor(g.stops[index])
	c2, _ := colorful.MakeColor(g.stops[index+1])
	return c1.BlendHcl(c2, lerp)
}

// colorAtPositions interpolates using the explicit stop positions. val is
// assumed already clamped to [0,1].
func (g *Gradient) colorAtPositions(val float64) color.Color {
	n := len(g.stops)
	// Clamp below the first / above the last stop.
	if val <= g.positions[0] {
		return g.stops[0]
	}
	if val >= g.positions[n-1] {
		return g.stops[n-1]
	}
	// Find the bracketing pair [index, index+1] with positions surrounding val.
	index := 0
	for index < n-1 && g.positions[index+1] < val {
		index++
	}
	p1, p2 := g.positions[index], g.positions[index+1]
	var lerp float64
	if span := p2 - p1; span > 0 {
		lerp = (val - p1) / span
	}
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
		paletteStrings[i] = hexOf(c)
	}
	return paletteStrings
}

// hexOf renders a color as "#rrggbb". colorful.MakeColor cannot convert a
// fully transparent color (it divides by alpha), so that case is handled
// separately rather than panicking.
func hexOf(c color.Color) string {
	if d, ok := colorful.MakeColor(c); ok {
		return d.Hex()
	}
	return "#000000"
}

type ColorSpace int

const (
	ColorSpaceRGB ColorSpace = iota
	ColorSpaceHSV
)

type SinePalette struct {
	A, B, C, D Vec3
	Alpha      float64
	Space      ColorSpace
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

func hsvToRGB(h, s, v float64) (r, g, b float64) {
	h = math.Mod(h, 1.0)
	if h < 0 {
		h += 1.0
	}
	if s == 0 {
		return v, v, v
	}
	h6 := h * 6.0
	i := int(h6)
	f := h6 - float64(i)
	p := v * (1 - s)
	q := v * (1 - s*f)
	u := v * (1 - s*(1-f))
	switch i % 6 {
	case 0:
		return v, u, p
	case 1:
		return q, v, p
	case 2:
		return p, v, u
	case 3:
		return p, q, v
	case 4:
		return u, p, v
	default:
		return v, p, q
	}
}

func (sp *SinePalette) ColorAt(t float64) color.Color {
	x := sp.A.X + sp.B.X*math.Cos(2*math.Pi*(sp.C.X*t+sp.D.X))
	y := sp.A.Y + sp.B.Y*math.Cos(2*math.Pi*(sp.C.Y*t+sp.D.Y))
	z := sp.A.Z + sp.B.Z*math.Cos(2*math.Pi*(sp.C.Z*t+sp.D.Z))
	var result color.RGBA64
	result.A = uint16(Clamp(0, 1, sp.Alpha) * 65535)
	switch sp.Space {
	case ColorSpaceHSV:
		r, g, b := hsvToRGB(x, Clamp(0, 1, y), Clamp(0, 1, z))
		result.R = uint16(Clamp(0, 1, r) * 65535)
		result.G = uint16(Clamp(0, 1, g) * 65535)
		result.B = uint16(Clamp(0, 1, b) * 65535)
	default:
		result.R = uint16(Clamp(0, 1, x) * 65535)
		result.G = uint16(Clamp(0, 1, y) * 65535)
		result.B = uint16(Clamp(0, 1, z) * 65535)
	}
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
	r := render.NewRaster(int(w), int(h))
	ctx := render.NewContext(r)
	ctx.SetStrokeWidth(0.1)
	// Each row halves the cell width: row r has 2^r cells. Looping over
	// integer row and column counts keeps the final row and column on the
	// canvas, which a floating-point "for y <= h" accumulation does not.
	rows := int(math.Floor(math.Log2(h))) - 1
	dy := h / float64(rows)
	for row := 0; row < rows; row++ {
		cols := 1 << uint(row)
		dx := w / float64(cols)
		y := float64(row) * dy
		for col := 0; col < cols; col++ {
			x := float64(col) * dx
			k := sp.ColorAt((x + 0.5*dx) / w)
			ctx.SetFillColor(k)
			ctx.SetStrokeColor(k)
			// Coarsest row at the bottom, like the original y-up layout.
			ctx.DrawRectangle(x, h-y-dy, dx, dy)
			ctx.FillStroke()
		}
	}
	fname := "gaul-sinepalette-grid_" + GetTimestampString() + ".png"
	if err := r.SavePNG(fname); err != nil {
		return fname, err
	}
	return fname, nil
}

func (sp *SinePalette) ColorAtStop(i, n int) color.Color {
	if n <= 1 {
		return sp.ColorAt(0)
	}
	return sp.ColorAt(float64(i) / float64(n-1))
}

func (sp *SinePalette) ToPng() (string, error) {
	w := 570.0
	h := 110.0
	r := render.NewRaster(int(w), int(h))
	ctx := render.NewContext(r)
	ctx.Clear(color.Black)
	font, err := render.LoadFont(goregular.TTF, "Go Regular")
	if err != nil {
		return "", err
	}
	ctx.SetFont(font, 14)
	// Layout: A and B labels above the palette strip (left- and
	// right-aligned), C and D labels below.
	left := 0.1 * w
	right := 0.9 * w
	topBaseline := 20.0
	bottomBaseline := h - 8.0
	stripTop := 30.0
	stripBottom := h - 30.0
	ctx.SetFillColor(color.White)
	aText := "A = " + sp.A.ToString(3)
	bText := "B = " + sp.B.ToString(3)
	cText := "C = " + sp.C.ToString(3)
	dText := "D = " + sp.D.ToString(3)
	ctx.DrawString(aText, left, topBaseline)
	ctx.DrawString(bText, right-ctx.MeasureString(bText), topBaseline)
	ctx.DrawString(cText, left, bottomBaseline)
	ctx.DrawString(dText, right-ctx.MeasureString(dText), bottomBaseline)
	t := Linspace(0, 1, 1000, true)
	dw := 0.8 * w / 1000
	ctx.SetStrokeWidth(2 * dw)
	for i, j := range t {
		ctx.SetStrokeColor(sp.ColorAt(j))
		x := left + float64(i)*dw
		ctx.MoveTo(x, stripTop)
		ctx.LineTo(x, stripBottom)
		ctx.Stroke()
	}
	fname := "gaul-sinepalette_" + GetTimestampString() + ".png"
	if err := r.SavePNG(fname); err != nil {
		return "", err
	}
	return fname, nil
}

// ColorAt satisfies the Palette interface for SimpleGradient.
func (sg *SimpleGradient) ColorAt(t float64) color.Color { return sg.Color(t) }

// ColorAtStop satisfies the Palette interface for SimpleGradient.
func (sg *SimpleGradient) ColorAtStop(i, n int) color.Color {
	if n <= 1 {
		return sg.ColorAt(0)
	}
	return sg.ColorAt(float64(i) / float64(n-1))
}

// ColorAt satisfies the Palette interface for Gradient.
func (g *Gradient) ColorAt(t float64) color.Color { return g.Color(t) }

// ColorAtStop satisfies the Palette interface for Gradient.
func (g *Gradient) ColorAtStop(i, n int) color.Color {
	if n <= 1 {
		return g.ColorAt(0)
	}
	return g.ColorAt(float64(i) / float64(n-1))
}

// NewGradientFromColors creates a Gradient directly from color.Color values.
// Used when reconstructing a gradient from the palette database.
func NewGradientFromColors(colors []color.Color) Gradient {
	return Gradient{stops: colors}
}

// NewGradientFromColorStops creates a Gradient with explicit normalized stop
// positions in [0,1]. positions must be the same length as colors and sorted in
// ascending order. If positions is nil the stops are treated as evenly spaced,
// matching NewGradientFromColors.
func NewGradientFromColorStops(colors []color.Color, positions []float64) Gradient {
	if positions != nil && len(positions) != len(colors) {
		panic("gaul: positions length must match colors length")
	}
	return Gradient{stops: colors, positions: positions}
}

// Palette is the common interface for all gaul palette types.
// ColorAt accepts a normalized position t ∈ [0,1].
// ColorAtStop returns the color at discrete stop i of n (t = i/(n-1)).
type Palette interface {
	ColorAt(t float64) color.Color
	ColorAtStop(i, n int) color.Color
}

var (
	_ Palette = (*SimpleGradient)(nil)
	_ Palette = (*Gradient)(nil)
	_ Palette = (*SinePalette)(nil)
)
