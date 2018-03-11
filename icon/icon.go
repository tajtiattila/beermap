package icon

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
)

type Renderer struct {
	font *truetype.Font

	Dim     int // icon width and height
	Padding int // padding, also shadow size
	Stroke  int // outer stroke width
}

func NewRendererFontPath(fontpath string) (*Renderer, error) {
	fontdata, err := ioutil.ReadFile(fontpath)
	if err != nil {
		return nil, err
	}

	return NewRendererFont(fontdata)
}

func NewRendererFont(fontdata []byte) (*Renderer, error) {
	font, err := freetype.ParseFont(fontdata)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse font")
	}

	return &Renderer{
		font: font,

		Dim:     56,
		Padding: 2,
		Stroke:  2,
	}, nil
}

func (r *Renderer) Render(d Drawable, c Colors, label string) image.Image {
	return d.Render(r, c, label)
}

func (r *Renderer) RenderPNG(d Drawable, c Colors, label string) []byte {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, d.Render(r, c, label))
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (r *Renderer) centerLabel(im draw.Image, textColor color.Color, label string) {
	if label == "" {
		return
	}

	var fs int
	if len(label) <= 2 {
		fs = 30
	} else {
		fs = 24
	}
	fs = r.Dim * fs / 56

	c := freetype.NewContext()
	c.SetFont(r.font)
	c.SetSrc(image.NewUniform(textColor))
	c.SetFontSize(float64(fs))
	tdim, err := c.DrawString(label, freetype.Pt(0, 0))
	if err != nil {
		log.Println(err)
		return
	}

	center := r.Dim / 2

	dy := (fs * 17 / 24) / 2
	textp := freetype.Pt(center, center+dy)
	textp.X -= tdim.X / 2

	c.SetDst(im)
	c.SetClip(im.Bounds())
	c.DrawString(label, textp)
}

type Drawable interface {
	Render(r *Renderer, colors Colors, label string) image.Image
}

type Colors struct {
	Outline color.Color
	Fill    color.Color
	Shadow  color.Color
	Text    color.Color
}

func SimpleColors(fill color.Color) Colors {
	return Colors{
		Outline: color.White,
		Fill:    fill,
		Shadow:  color.NRGBA{0, 0, 0, 73},
		Text:    color.White,
	}
}

type circleIcon struct{}

var Circle Drawable = circleIcon{}

func (circleIcon) Render(r *Renderer, colors Colors, label string) image.Image {
	center := r.Dim / 2
	radius := center - r.Padding
	im := image.NewRGBA(image.Rect(0, 0, r.Dim, r.Dim))

	tmp := image.NewRGBA(image.Rect(0, 0, r.Dim, r.Dim))
	drawCircle(tmp, colors.Shadow, center, center+r.Padding, radius)
	draw.Draw(im, im.Bounds(), tmp, tmp.Bounds().Min, draw.Over)

	if r.Stroke > 0 {
		clearRGBA(tmp)
		drawCircle(tmp, colors.Outline, center, center, radius)
		draw.Draw(im, im.Bounds(), tmp, tmp.Bounds().Min, draw.Over)
	}

	clearRGBA(tmp)
	drawCircle(tmp, colors.Fill, center, center, radius-r.Stroke)
	draw.Draw(im, im.Bounds(), tmp, tmp.Bounds().Min, draw.Over)

	r.centerLabel(im, colors.Text, label)

	return im
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
		fy := float64(y) + 0.5 - float64(cy)
		for x := cr.Min.X; x < cr.Max.X; x++ {
			fx := float64(x) + 0.5 - float64(cx)

			rx := math.Sqrt(fx*fx + fy*fy)

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

type squareIcon struct{}

var Square Drawable = squareIcon{}

func (squareIcon) Render(r *Renderer, colors Colors, label string) image.Image {
	im := image.NewNRGBA(image.Rect(0, 0, r.Dim, r.Dim))

	// top and bottom lines
	for i := 0; i < r.Stroke; i++ {
		yt := r.Padding + i
		yb := r.Dim - r.Padding - i - 1
		for x := r.Padding; x < r.Dim-r.Padding; x++ {
			im.Set(x, yt, colors.Outline)
			im.Set(x, yb, colors.Outline)
		}
	}

	// middle section
	ps := r.Padding + r.Stroke
	xl := ps
	xr := r.Dim - ps
	for y := ps; y < r.Dim-ps; y++ {
		for i := 0; i < r.Stroke; i++ {
			im.Set(xl-i-1, y, colors.Outline)
			im.Set(xr+i, y, colors.Outline)
		}
		for x := xl; x < xr; x++ {
			im.Set(x, y, colors.Fill)
		}
	}

	// shadow
	for y := r.Dim - r.Padding; y < r.Dim; y++ {
		for x := r.Padding; x < r.Dim-r.Padding; x++ {
			im.Set(x, y, colors.Shadow)
		}
	}

	r.centerLabel(im, colors.Text, label)

	return im
}
