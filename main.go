package main

import (
	"flag"
	"fmt"
	"html/template"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/tajtiattila/basedir"
	"github.com/tajtiattila/geocode"
)

func main() {
	addr := flag.String("addr", ":8080", "default listen address")
	res := flag.String("res", "./res", "resource path")
	flag.Parse()

	gmapsapikey := os.Getenv("GOOGLEMAPS_APIKEY")
	if gmapsapikey == "" {
		log.Fatal("GOOGLEMAPS_APIKEY environment variable unset")
	}

	if flag.NArg() > 1 {
		log.Fatal("at most one arg needed")
	}

	var publist string
	if flag.NArg() == 1 {
		publist = flag.Arg(0)
	} else {
		publist = "serlist.txt"
	}

	pubList, err := getPubList(publist, gmapsapikey)
	if err != nil {
		log.Fatalln("can't load pub list:", err)
	}
	log.Println(len(pubList), "pubs in list")

	if err := writeKMZPubListFile("serlist.kmz", *res, pubList); err != nil {
		log.Println(err)
	}

	td := struct {
		GoogleMapsApiKey string
	}{
		gmapsapikey,
	}
	http.Handle("/", serveDirTemplate(*res, td))

	http.Handle("/pubs.json", servePubData(pubList))

	log.Println("listening on", *addr)
	log.Println(http.ListenAndServe(*addr, nil))
}

func serveDirTemplate(path string, data interface{}) http.Handler {
	const indexHtml = "index.html"
	def := http.FileServer(http.Dir(path))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			def.ServeHTTP(w, req)
			return
		}

		src, err := ioutil.ReadFile(filepath.Join(path, indexHtml))
		if err != nil {
			log.Println("read template:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		name := filepath.Base(path)
		t, err := template.New(name).Parse(string(src))
		if err != nil {
			log.Println("parse template:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, data); err != nil {
			log.Println("execute template:", err)
		}
	})
}

func getPubList(fn, gmapsapikey string) ([]Pub, error) {
	cacheDir, err := basedir.Cache.EnsureDir("beermap", 0777)
	if err != nil {
		return nil, errors.Wrap(err, "can't get cache dir")
	}

	qc, err := geocode.LevelDB(filepath.Join(cacheDir, "geocode.leveldb"))
	if err != nil {
		return nil, errors.Wrap(err, "can't open geocache")
	}
	defer qc.Close()

	gc := geocode.Cache(geocode.StdGoogle(gmapsapikey), qc)

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parsePubList(f, gc)
}

func writeKMZPubListFile(outname, res string, pubs []Pub) error {
	r, err := NewIconRenderer(res)
	if err != nil {
		return err
	}

	f, err := os.Create(outname)
	if err != nil {
		return err
	}
	defer f.Close()

	kmz, err := NewKMZ(f, "serlist")
	if err != nil {
		return err
	}

	writeKMZPubList(kmz, pubs, r)

	return kmz.Close()
}

func writeKMZPubList(kmz *KMZ, pubs []Pub, r *IconRenderer) {
	for _, p := range pubs {
		var c color.Color
		//Kek 1e90ff, zold 9acd32, piros b22222
		switch {
		case p.Has("#closed"):
			c = color.NRGBA{0xb2, 0x22, 0x22, 0xff}
		case p.Has("#user"):
			c = color.NRGBA{0x22, 0x8b, 0x22, 0xff}
		default:
			c = color.NRGBA{0x1e, 0x90, 0xff, 0xff}
			//c = color.NRGBA{2, 136, 209, 255}
		}
		ci := CircleIcon{
			Outline: color.White,
			Fill:    c,
			Shadow:  color.NRGBA{0, 0, 0, 73},
			Text:    color.White,
			Label:   fmt.Sprint(p.Num),
		}
		err := kmz.IconPlacemark(ci.Render(r), Placemark{
			Title: fmt.Sprintf("[%03d] %s", p.Num, p.Title),
			Desc:  p.Addr + "\n" + strings.Join(p.Desc, "\n"),
			Lat:   p.Geo.Lat,
			Long:  p.Geo.Long,
		})
		if err != nil {
			log.Println(err)
		}
	}
}
