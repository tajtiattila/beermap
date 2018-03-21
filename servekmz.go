package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/tajtiattila/beermap/keyvalue"
)

func writeKMZ(w io.Writer, db keyvalue.DB, mm mapMeta) error {
	srcKey := "src|" + mm.Key
	src, err := db.Get(srcKey)
	if err != nil {
		return err
	}

	var pubs []Pub
	if err := json.Unmarshal(src, &pubs); err != nil {
		return err
	}

	kmz := NewKMZ(w, mm.Title)
	for _, p := range pubs {
		iconKey := "path|" + path.Join(mm.Key, p.IconBasename())
		icon, err := db.Get(iconKey)
		if err != nil {
			return err
		}

		pm := Placemark{
			Title: fmt.Sprintf("[%s] %s", p.Label, p.Title),
			Desc:  p.Addr + "\n" + strings.Join(p.Desc, "\n"),
			Lat:   p.Geo.Lat,
			Long:  p.Geo.Long,
		}
		if err := kmz.IconPlacemark(icon, pm); err != nil {
			return err
		}
	}

	return kmz.Close()
}
