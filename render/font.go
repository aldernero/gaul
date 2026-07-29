package render

import (
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// Font is a parsed TTF or OTF font usable with Context.SetFont. Text support
// is intentionally basic: single strings drawn at a baseline position, no
// layout or wrapping.
//
// The two backends render text differently by design: the raster backend
// draws real glyphs from the font file, while the SVG backend emits <text>
// elements referencing the font by name, so the viewer substitutes its own
// font file and metrics may differ. For pen-plotter output, avoid text or
// convert it to paths downstream.
type Font struct {
	otf *opentype.Font

	faces map[float64]font.Face
	// Name is the font family name written to SVG output.
	Name string

	mu sync.Mutex
}

// LoadFont parses TTF/OTF font data. The name is used as the SVG
// font-family. Example using the Go font shipped with x/image:
//
//	f, err := render.LoadFont(goregular.TTF, "Go Regular")
func LoadFont(data []byte, name string) (*Font, error) {
	otf, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return &Font{Name: name, otf: otf, faces: make(map[float64]font.Face)}, nil
}

// LoadFontFile parses a TTF/OTF font file.
func LoadFontFile(path, name string) (*Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFont(data, name)
}

// face returns a cached font.Face for the given pixel size. DPI is fixed at
// 72 so that point size equals pixel size.
func (f *Font) face(sizePx float64) (font.Face, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if face, ok := f.faces[sizePx]; ok {
		return face, nil
	}
	face, err := opentype.NewFace(f.otf, &opentype.FaceOptions{
		Size:    sizePx,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	f.faces[sizePx] = face
	return face, nil
}

// Measure returns the advance width of s in pixels at the given size.
func (f *Font) Measure(s string, sizePx float64) float64 {
	face, err := f.face(sizePx)
	if err != nil {
		return 0
	}
	return float64(font.MeasureString(face, s)) / 64
}

// Metrics returns the ascent and descent in pixels at the given size.
func (f *Font) Metrics(sizePx float64) (ascent, descent float64) {
	face, err := f.face(sizePx)
	if err != nil {
		return 0, 0
	}
	m := face.Metrics()
	return float64(m.Ascent) / 64, float64(m.Descent) / 64
}
