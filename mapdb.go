package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/tajtiattila/beermap/keyvalue"
)

type mapDB struct {
	db keyvalue.DB
}

func (m mapDB) GetMeta(key string) (mapMeta, error) {
	var mm mapMeta
	raw, err := m.db.Get("meta|" + key)
	if err != nil {
		return mm, err
	}
	if err := json.Unmarshal(raw, &mm); err != nil {
		return mm, err
	}
	mm.Key = key
	return mm, nil
}

func (m mapDB) StoreMeta(mm mapMeta) error {
	if !mapKeyGen.validKey(mm.Key) {
		return errors.New("invalid map key")
	}
	raw, err := json.Marshal(mm)
	if err != nil {
		return err
	}
	return m.db.Set("meta|"+mm.Key, raw)
}

func (m mapDB) Access(key string) error {
	if !mapKeyGen.validKey(key) {
		return errors.New("invalid map key")
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	return m.db.Set("access|"+key, []byte(ts))
}

// mapMeta holds map metadata
type mapMeta struct {
	Key      string `json:"-"`
	EditPass string `json:"editPass"`

	Title string `json:"title"`

	ModTime time.Time `json:"modTime"`

	PubCount      int `json:"pubCount"`
	TotalPubCount int `json:"totalPubCount"`
}

func (mm mapMeta) styleKey() string {
	return "mapstyle|" + mm.Key
}
