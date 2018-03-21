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
	prefix := flag.String("prefix", "", "optional server prefix")
	flag.Parse()

	if p := *prefix; p != "" {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		serveHttpPrefix = p
	}

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
	httpHandle("/edit/", editor)

	httpHandle("/map/", http.StripPrefix("/map",
		mapHandler(filepath.Join(*res, "ui/map"), gmapsapikey, mdb)))

	httpHandle("/", http.FileServer(http.Dir(filepath.Join(*res, "ui/root"))))

	var withPrefix string
	if serveHttpPrefix != "" {
		withPrefix = fmt.Sprintf("with prefix %q", serveHttpPrefix)
	}
	log.Printf("listening on %v%s", *addr, withPrefix)

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
