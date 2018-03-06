package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"path"
	"regexp"
	"time"

	"github.com/tajtiattila/beermap/icon"
)

type mapData struct {
	Bounds jbounds `json:"bounds"`
	Pubs   []jpub  `json:"pubs"`
}

type jbounds struct {
	N float64 `json:"north"`
	E float64 `json:"east"`
	S float64 `json:"south"`
	W float64 `json:"west"`
}

type jpub struct {
	Title   string  `json:"title"`
	Lat     float64 `json:"lat"`
	Long    float64 `json:"lng"`
	Icon    string  `json:"icon"`
	Visited bool    `json:"visited"`
	Closed  bool    `json:"closed"`
	Content string  `json:"content"`
}

func servePubData(pubs []Pub, iconpfx string) http.Handler {
	var md mapData
	for i, p := range pubs {
		if i == 0 {
			md.Bounds.N = p.Geo.Lat
			md.Bounds.S = p.Geo.Lat
			md.Bounds.E = p.Geo.Long
			md.Bounds.W = p.Geo.Long
		} else {
			if p.Geo.Lat > md.Bounds.N {
				md.Bounds.N = p.Geo.Lat
			}
			if p.Geo.Lat < md.Bounds.S {
				md.Bounds.S = p.Geo.Lat
			}
			if p.Geo.Long > md.Bounds.E {
				md.Bounds.E = p.Geo.Long
			}
			if p.Geo.Long < md.Bounds.W {
				md.Bounds.W = p.Geo.Long
			}
		}

		buf := new(bytes.Buffer)
		xp := struct {
			Pub
			Icon string
		}{
			Pub:  p,
			Icon: path.Join(iconpfx, fmt.Sprintf("icon-%d.png", p.Num)),
		}
		err := contentTmpl.Execute(buf, xp)
		if err != nil {
			log.Fatal(err)
		}

		jp := jpub{
			Title:   p.Title,
			Lat:     p.Geo.Lat,
			Long:    p.Geo.Long,
			Icon:    xp.Icon,
			Visited: p.Has("#user"),
			Closed:  p.Has("#closed"),
			Content: buf.String(),
		}
		md.Pubs = append(md.Pubs, jp)
	}
	raw, err := json.Marshal(md)
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeContent(w, req, "pubs.json", now, bytes.NewReader(raw))
	})
}

var contentTmpl = template.Must(template.New("info").Funcs(template.FuncMap{
	"addLinks": addLinks,
}).Parse(`<h1><img src="{{.Icon}}" width="28" height="28">{{.Title}}</h1>
<p>{{.Addr}}</p>
<p>{{range .Desc}}
{{. | addLinks}}<br>
{{end}}</p>
`))

var linkRe = regexp.MustCompile(`(http|ftp|https)://([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)

func addLinks(s string) template.HTML {
	s = template.HTMLEscapeString(s)
	return template.HTML(linkRe.ReplaceAllString(s, `<a target="pub" href="$0">$0</a>`))
}

func getPubIcon(p Pub, r *icon.Renderer) image.Image {
	var c color.Color
	//Kek 1e90ff, zold 9acd32, piros b22222
	switch {
	case p.Has("#closed"):
		c = color.NRGBA{0xb2, 0x22, 0x22, 0xff}
	case p.Has("#user"):
		c = color.NRGBA{0x22, 0x8b, 0x22, 0xff}
	default:
		c = color.NRGBA{0x1e, 0x90, 0xff, 0xff}
	}
	return r.Render(icon.Circle, icon.SimpleColors(c), fmt.Sprint(p.Num))
}

func servePubIcons(pubs []Pub, r *icon.Renderer, pfx string) http.Handler {
	icons := make(map[string][]byte)
	for _, p := range pubs {
		buf := new(bytes.Buffer)
		err := png.Encode(buf, getPubIcon(p, r))
		if err != nil {
			log.Fatal(err)
		}
		icons[path.Join(pfx, fmt.Sprintf("icon-%d.png", p.Num))] = buf.Bytes()
	}
	now := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p := req.URL.Path
		raw, ok := icons[p]
		if !ok {
			log.Println(p)
			http.NotFound(w, req)
			return
		}
		http.ServeContent(w, req, path.Base(p), now, bytes.NewReader(raw))
	})
}
