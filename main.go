package main

import (
	"flag"
	"fmt"
	"html/template"
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
	if len(pubList) == 0 {
		log.Fatalln("empty pub list")
	}
	log.Println(len(pubList), "pubs in list")

	styler, err := NewStylerPath(filepath.Join(*res, "iconstyle.json"))
	if err != nil {
		log.Fatalln("can't init icon styler:", err)
	}

	if err := writeKMZPubListFile("serlist.kmz", pubList, styler); err != nil {
		log.Println(err)
	}

	td := struct {
		GoogleMapsApiKey string
	}{
		gmapsapikey,
	}
	http.Handle("/", serveDirTemplate(*res, td))

	const pubIconPfx = "/pub-icon"
	http.Handle("/pubs.json", servePubData(pubList, pubIconPfx))
	http.Handle(pubIconPfx+"/", servePubIcons(pubList, styler, pubIconPfx))

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

func writeKMZPubListFile(outname string, pubs []Pub, styler *Styler) error {
	f, err := os.Create(outname)
	if err != nil {
		return err
	}
	defer f.Close()

	kmz, err := NewKMZ(f, "serlist")
	if err != nil {
		return err
	}

	writeKMZPubList(kmz, pubs, styler)

	return kmz.Close()
}

func writeKMZPubList(kmz *KMZ, pubs []Pub, styler *Styler) {
	for _, p := range pubs {
		err := kmz.IconPlacemark(styler.PubIcon(p), Placemark{
			Title: fmt.Sprintf("[%s] %s", p.Label, p.Title),
			Desc:  p.Addr + "\n" + strings.Join(p.Desc, "\n"),
			Lat:   p.Geo.Lat,
			Long:  p.Geo.Long,
		})
		if err != nil {
			log.Println(err)
		}
	}
}
