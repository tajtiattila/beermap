package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"image/png"
	"log"
	"net/http"
	"path"
	"regexp"
	"time"
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
	Label   string  `json:"label"`
	Title   string  `json:"title"`
	Lat     float64 `json:"lat"`
	Long    float64 `json:"lng"`
	Icon    string  `json:"icon"`
	Content string  `json:"content"`
}

func pubListJSON(pubs []Pub, iconpfx string) []byte {
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
			Icon: path.Join(iconpfx, p.IconBasename()),
		}
		err := contentTmpl.Execute(buf, xp)
		if err != nil {
			log.Fatal(err)
		}

		jp := jpub{
			Label:   p.Label,
			Title:   p.Title,
			Lat:     p.Geo.Lat,
			Long:    p.Geo.Long,
			Icon:    xp.Icon,
			Content: buf.String(),
		}
		md.Pubs = append(md.Pubs, jp)
	}
	raw, err := json.Marshal(md)
	if err != nil {
		log.Fatal(err)
	}
	return raw
}

func servePubData(pubs []Pub, iconpfx string) http.Handler {
	raw := pubListJSON(pubs, iconpfx)
	now := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeContent(w, req, "pubs.json", now, bytes.NewReader(raw))
	})
}

var contentTmpl = template.Must(template.New("info").Funcs(template.FuncMap{
	"addLinks": addLinks,
}).Parse(`<h1 class="pubinfo-title">
<span class="pubinfo-icon"><img src="{{.Icon}}"></span>
<span class="pubinfo-titletext">{{.Title}}</span>
</h1>
<p class="pubinfo-addr">{{.Addr}}</p>
<p class="pubinfo-desc">{{range .Desc}}
{{. | addLinks}}<br>
{{end}}</p>
`))

var linkRe = regexp.MustCompile(`(http|ftp|https)://([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)

func addLinks(s string) template.HTML {
	s = template.HTMLEscapeString(s)
	return template.HTML(linkRe.ReplaceAllString(s, `<a target="pub" href="$0">$0</a>`))
}

// pubIconData generates PNG image data for pub using styler
func pubIconData(pub Pub, styler *Styler) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, styler.PubIcon(pub))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
