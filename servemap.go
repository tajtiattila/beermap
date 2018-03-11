package main

import (
	"bytes"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/tajtiattila/beermap/keyvalue"
)

const pubjson = "pubs.json"

func mapHandler(res, googlemapsapikey string, mdb *mapDB) http.Handler {
	dirh := http.FileServer(http.Dir(filepath.Join(res)))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key, rest := splitUIPath(req)

		mm, err := mdb.GetMeta(key)
		if err != nil {
			http.NotFound(w, req)
			return
		}

		p := rest.URL.Path
		if p == "/" {
			td := struct {
				GoogleMapsApiKey string
				Title            string
			}{
				googlemapsapikey,
				mm.Title,
			}

			const indexHtml = "index.html"
			serveTemplateWithData(w, req, filepath.Join(res, indexHtml), td)
			return
		}

		if p == "/"+pubjson || strings.HasPrefix(p, "/icon-") {
			k := "path|" + path.Join(mm.Key, p)
			raw, err := mdb.db.Get(k)
			if err != nil {
				log.Printf("data access %v: %v", k, err)
				if err == keyvalue.ErrNotFound {
					http.NotFound(w, req)
				} else {
					httpErrorCode(w, http.StatusInternalServerError)
				}
				return
			}
			http.ServeContent(w, req, path.Base(p), mm.ModTime, bytes.NewReader(raw))
			return
		}

		if p == "/mapstyle.json" {
			raw, err := mdb.db.Get(mm.styleKey())
			if err != nil {
				if err != keyvalue.ErrNotFound {
					log.Printf("mapstyle access %v: %v", req, err)
				}
				raw = []byte("[]")
			}
			http.ServeContent(w, req, path.Base(p), mm.ModTime, bytes.NewReader(raw))
			return
		}

		dirh.ServeHTTP(w, rest)
	})
}

func splitUIPath(r *http.Request) (mapKey string, sub *http.Request) {
	p := r.URL.Path
	if p == "" || p[0] != '/' {
		return "", r
	}
	i := strings.IndexRune(p[1:], '/') + 1
	if i <= 0 {
		i = len(p)
	}
	return p[1:i], requestWithPath(r, p[i:])
}
