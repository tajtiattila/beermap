package main

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/pkg/errors"
	"github.com/tajtiattila/beermap/googlefont"
	"github.com/tajtiattila/beermap/keyvalue"
)

type FontCache struct {
	db keyvalue.DB

	prefix string

	maxAge time.Duration
}

func (fc *FontCache) Get(name string) ([]byte, error) {
	rawmeta, err := fc.db.Get(fc.key("meta", name))
	if err == keyvalue.ErrNotFound {
		return fc.fetch(name)
	}

	var fm fcMeta
	if err := json.Unmarshal(rawmeta, &fm); err != nil {
		log.Println("FontCache meta unmarshal:", err)
		return fc.fetch(name)
	}

	var fetchErr error
	if fm.Cached.Add(fc.maxAge).Before(time.Now()) {
		r, err := fc.fetch(name)
		if err == nil {
			return r, nil
		}
		log.Println("FontCache update failed:", err)
		fetchErr = err
	}

	raw, err := fc.db.Get(fc.key("font", name))
	if err == nil {
		return raw, nil
	}

	if fetchErr == nil {
		// try fetching in case of cache error
		return fc.fetch(name)
	}

	// both cache and fetch failed, return fetch error
	return nil, fetchErr
}

func (fc *FontCache) fetch(name string) ([]byte, error) {
	face, weight, err := splitFontName(name)
	if err != nil {
		return nil, err
	}
	raw, err := googlefont.Get(face, weight)
	if err != nil {
		return nil, errors.Wrap(err, "Fontcache fetch")
	}

	// don't cache unusable font
	if _, err := freetype.ParseFont(raw); err != nil {
		return nil, errors.Wrap(err, "Fontcache parse")
	}

	if err := fc.db.Set(fc.key("font", name), raw); err != nil {
		log.Println("FontCache store:", err)
		return raw, nil
	}

	fm, err := json.Marshal(fcMeta{
		Cached: time.Now(),
	})
	if err != nil {
		log.Println("FontCache metadata masrhal:", err)
		return raw, nil
	}
	if err := fc.db.Set(fc.key("meta", name), fm); err != nil {
		log.Println("FontCache metadata store:", err)
		return raw, nil
	}

	return raw, nil
}

type fcMeta struct {
	Cached time.Time `json:"cached"`
}

func (fc *FontCache) key(kind, k string) string {
	return fc.prefix + kind + "/" + k
}

func splitFontName(name string) (face string, weight int, err error) {
	i := strings.LastIndex(name, ":")
	if i == -1 {
		return "", 0, errors.New("fontcache: missing weight")
	}
	face = name[:i]
	weight, err = strconv.Atoi(name[i+1:])
	if err != nil {
		return "", 0, errors.New("fontcache: invalid weight")
	}
	return face, weight, nil
}
