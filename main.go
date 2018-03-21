package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tajtiattila/basedir"
	"github.com/tajtiattila/beermap/icon"
	"github.com/tajtiattila/beermap/keyvalue"
	"github.com/tajtiattila/geocode"
)

func main() {
	addr := flag.String("addr", ":8080", "default listen address")
	res := flag.String("res", "./res", "resource path")
	dbpath := flag.String("db", "./db", "database path")
	flag.Parse()

	gmapsapikey := os.Getenv("GOOGLEMAPS_APIKEY")
	if gmapsapikey == "" {
		log.Fatal("GOOGLEMAPS_APIKEY environment variable unset")
	}

	if flag.NArg() > 1 {
		log.Fatal("at most one arg needed")
	}

	ir, err := icon.NewRendererFontPath(filepath.Join(*res, "Roboto-Medium.ttf"))
	if err != nil {
		log.Fatalln("can't start icon renderer", err)
	}

	db, err := keyvalue.OpenLevelDB(*dbpath)
	if err != nil {
		log.Fatalln("can't open db", err)
	}
	defer db.Close()

	mdb := &mapDB{db: db}

	fontcache := &FontCache{
		db:     db,
		prefix: "fontdata/",
		maxAge: time.Hour,
	}

	gc, closer, err := newGeocoder(gmapsapikey)
	if err != nil {
		log.Fatalln("can't start geocoder", err)
	}
	defer closer.Close()

	editor := newEditor("/edit/", filepath.Join(*res, "ui/edit"), mdb, gc)
	editor.defaultIconRenderer = ir
	editor.fontSrc = fontcache.Get
	http.Handle("/edit/", editor)

	http.Handle("/map/", http.StripPrefix("/map",
		mapHandler(filepath.Join(*res, "ui/map"), gmapsapikey, mdb)))

	http.Handle("/", http.FileServer(http.Dir(filepath.Join(*res, "ui/root"))))

	log.Println("listening on", *addr)
	log.Println(http.ListenAndServe(*addr, nil))
}

func newGeocoder(gmapsapikey string) (geocode.Geocoder, io.Closer, error) {
	cacheDir, err := basedir.Cache.EnsureDir("beermap", 0777)
	if err != nil {
		return nil, nil, errors.Wrap(err, "can't get cache dir")
	}

	qc, err := geocode.LevelDB(filepath.Join(cacheDir, "geocode.leveldb"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "can't open geocache")
	}

	gc := geocode.LatLong(geocode.OpenLocationCode(geocode.Cache(geocode.StdGoogle(gmapsapikey), qc)))
	return gc, qc, nil
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
