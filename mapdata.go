package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
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
	Lat     float64 `json:"lat"`
	Long    float64 `json:"lng"`
	Visited bool    `json:"visited"`
	Closed  bool    `json:"closed"`
	Content string  `json:"content"`
}

func servePubData(pubs []Pub) http.Handler {
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
		err := contentTmpl.Execute(buf, p)
		if err != nil {
			log.Fatal(err)
		}

		jp := jpub{
			Label:   p.NumStr(),
			Lat:     p.Geo.Lat,
			Long:    p.Geo.Long,
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

var contentTmpl = template.Must(template.New("info").Parse(`<h1>[{{.NumStr}}] {{.Title}}</h1>
<p>{{.Addr}}</p>
<p>{{range .Desc}}
{{.}}<br>
{{end}}</p>
`))
