package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
)

type IconRenderer struct {
	font *truetype.Font
}

func NewIconRenderer(res string) (*IconRenderer, error) {
	raw, err := ioutil.ReadFile(filepath.Join(res, "Roboto-Medium.ttf"))
	if err != nil {
		return nil, errors.Wrap(err, "can't load font")
	}

	font, err := freetype.ParseFont(raw)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse font")
	}

	return &IconRenderer{
		font: font,
	}, nil
}

type CircleIcon struct {
	Outline color.Color
	Fill    color.Color
	Shadow  color.Color

	Text  color.Color
	Label string
}

func (ci CircleIcon) Render(r *IconRenderer) image.Image {
	const dim = 56
	const center = dim / 2
	const radius = center - 2 // outer
	im := image.NewRGBA(image.Rect(0, 0, dim, dim))

	tmp := image.NewRGBA(image.Rect(0, 0, dim, dim))
	drawCircle(tmp, ci.Shadow, center, center+2, radius)
	draw.Draw(im, im.Bounds(), tmp, tmp.Bounds().Min, draw.Over)

	clearRGBA(tmp)
	drawCircle(tmp, ci.Outline, center, center, radius)
	draw.Draw(im, im.Bounds(), tmp, tmp.Bounds().Min, draw.Over)

	clearRGBA(tmp)
	drawCircle(tmp, ci.Fill, center, center, radius-2)
	draw.Draw(im, im.Bounds(), tmp, tmp.Bounds().Min, draw.Over)

	if ci.Label == "" {
		return im
	}

	var fs int
	if len(ci.Label) <= 2 {
		fs = 30
	} else {
		fs = 24
	}

	c := freetype.NewContext()
	c.SetFont(r.font)
	c.SetDst(tmp)
	c.SetClip(tmp.Bounds())
	c.SetSrc(image.NewUniform(ci.Text))
	c.SetFontSize(float64(fs))
	tdim, err := c.DrawString(ci.Label, freetype.Pt(0, 0))
	if err != nil {
		log.Println(err)
		return im
	}

	dy := (fs * 17 / 24) / 2
	textp := freetype.Pt(center, center+dy)
	textp.X -= tdim.X / 2

	c.SetDst(im)
	c.DrawString(ci.Label, textp)

	return im
}

func (ci CircleIcon) RenderPNG(r *IconRenderer) []byte {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, ci.Render(r))
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func clearRGBA(im *image.RGBA) {
	for i := range im.Pix {
		im.Pix[i] = 0
	}
}

func drawCircle(im draw.Image, c color.Color, cx, cy, r int) {
	cn := color.NRGBAModel.Convert(c).(color.NRGBA)
	alpha := float64(cn.A)
	cr := image.Rect(cx-r, cy-r, cx+r, cy+r)
	cr = im.Bounds().Intersect(cr)
	for y := cr.Min.Y; y < cr.Max.Y; y++ {
		fy0 := float64(y) + 0.5 - float64(cy)
		//fy1 := float64(y + 1 - cy)
		for x := cr.Min.X; x < cr.Max.X; x++ {
			fx0 := float64(x) + 0.5 - float64(cx)
			//fx1 := float64(x + 1 - cx)

			rx := dst(fy0, fx0)
			/*
				r1 := dst(fy1, fx0)
				r2 := dst(fy0, fx1)
				r3 := dst(fy1, fx1)
				rx := (r0 + r1 + r2 + r3) / 4
			*/

			v := rx - float64(r) + 0.5
			switch {
			case v < 0:
				im.Set(x, y, c)
			case v < 1:
				cn.A = uint8(alpha * (1 - v))
				im.Set(x, y, cn)
			}
		}
	}
}

func dst(a, b float64) float64 {
	return math.Sqrt(a*a + b*b)
}
