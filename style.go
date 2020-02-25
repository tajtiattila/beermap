package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tajtiattila/beermap/icon"
)

type Style struct {
	Name   string
	Ignore bool // ignore this style
	Cond   Cond // condition to use this style

	Shape icon.Drawable
	Color color.Color // shape fill
}

type Styler struct {
	r *icon.Renderer

	styles []Style

	niceLabel bool
}

func NewStylerPath(fn string) (*Styler, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewStyler(f, func(fontname string) ([]byte, error) {
		return ioutil.ReadFile(filepath.Join(filepath.Dir(fn), fontname))
	})
}

func NewStyler(r io.Reader, readFont func(fn string) ([]byte, error)) (*Styler, error) {
	var j struct {
		Font      string  `json:"font"`
		Styles    []Style `json:"styles"`
		NiceLabel bool    `json:"niceLabel"`
	}
	if err := json.NewDecoder(r).Decode(&j); err != nil {
		return nil, err
	}
	font, err := readFont(j.Font)
	if err != nil {
		return nil, err
	}
	iconr, err := icon.NewRendererFont(font)
	if err != nil {
		return nil, errors.Wrap(err, "can't init icon renderer")
	}
	return &Styler{
		r:         iconr,
		styles:    j.Styles,
		niceLabel: j.NiceLabel,
	}, nil
}

func (st *Styler) Visible(p Pub) bool {
	for _, s := range st.styles {
		if !s.Ignore && s.Cond.Accept(p) && s.Shape == nil {
			return false
		}
	}
	return true
}

func (st *Styler) PubIcon(p Pub) image.Image {
	label := p.Label
	if st.niceLabel {
		if n, err := strconv.Atoi(label); err == nil {
			label = fmt.Sprint(n)
		}
	}
	for _, s := range st.styles {
		if !s.Ignore && s.Cond.Accept(p) {
			return st.r.Render(s.Shape, icon.SimpleColors(s.Color), label)
		}
	}
	return st.r.Render(icon.Square, icon.SimpleColors(color.Black), label)
}

func (s *Style) UnmarshalJSON(p []byte) error {
	var j jStyle
	if err := json.Unmarshal(p, &j); err != nil {
		return err
	}

	s.Name = j.Name
	s.Ignore = j.Ignore

	var err error
	s.Cond, err = decodeCond(j.Cond)
	if err != nil {
		return err
	}

	switch j.Shape {
	case "circle":
		s.Shape = icon.Circle
	case "square":
		s.Shape = icon.Square
	case "none":
		// ignore color
		return nil
	case "":
		return errors.New("missing shape")
	default:
		return errors.Errorf("unknown shape %q", j.Shape)
	}

	s.Color, err = decodeColor(j.Color)
	if err != nil {
		return err
	}

	return nil
}

func decodeOptionalColor(s string, def color.Color) (color.Color, error) {
	if s == "" {
		return def, nil
	}
	return decodeColor(s)
}

func decodeColor(s string) (color.Color, error) {
	if s == "" {
		return nil, errors.New("empty color spec")
	}

	if s[0] != '#' || (len(s) != 4 && len(s) != 7) {
		return nil, errors.Errorf("invalid color spec %q", s)
	}

	v, err := hexDigits(s[1:])
	if err != nil {
		return nil, errors.Wrapf(err, "invalid color spec %q", s)
	}

	var r, g, b uint8
	if len(v) == 3 {
		// eg #6f2
		r = v[0] * 0x11
		g = v[1] * 0x11
		b = v[2] * 0x11
	} else {
		// eg #66ff22
		r = v[0]<<4 + v[1]
		g = v[2]<<4 + v[3]
		b = v[4]<<4 + v[5]
	}
	return color.NRGBA{r, g, b, 0xff}, nil
}

func hexDigits(s string) ([]byte, error) {
	var v []byte
	for _, r := range s {
		switch {
		case '0' <= r && r <= '9':
			v = append(v, byte(r-'0'))
		case 'a' <= r && r <= 'f':
			v = append(v, byte(r-'a'+10))
		case 'A' <= r && r <= 'F':
			v = append(v, byte(r-'A'+10))
		default:
			return nil, errors.Errorf("invalid hex digit %q", r)
		}
	}
	return v, nil
}

func jsstring(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

type jStyle struct {
	Name string `json:"name"`

	Ignore bool `json:"ignore"`

	Cond interface{} `json:"cond"`

	Shape string `json:"shape"`
	Color string `json:"color"`
}
