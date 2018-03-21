package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tajtiattila/beermap/icon"
	"github.com/tajtiattila/beermap/keyvalue"
	"github.com/tajtiattila/geocode"
)

type editor struct {
	resdir string
	prefix string

	mdb *mapDB

	// gc is used to look up addresses
	gc geocode.Geocoder

	// fontSrc retursn raw TTF fonts
	fontSrc func(fontname string) ([]byte, error)

	defaultIconRenderer *icon.Renderer

	base http.Handler
}

func newEditor(prefix, resdir string, mdb *mapDB, gc geocode.Geocoder) *editor {
	return &editor{
		prefix: prefix,
		resdir: resdir,
		mdb:    mdb,
		gc:     gc,
		fontSrc: func(fn string) ([]byte, error) {
			return ioutil.ReadFile(filepath.Join("res", fn))
		},

		base: http.FileServer(http.Dir(resdir)),
	}
}

func (e *editor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, e.prefix) {
		http.NotFound(w, req)
		return
	}

	sub := strings.TrimPrefix(req.URL.Path, e.prefix)
	if sub == "new" {
		e.serveNew(w, req)
		return
	}
	i := strings.IndexRune(sub, '/')
	var key string
	if i >= 0 {
		key, sub = sub[:i], sub[i:]
	} else {
		key, sub = sub, ""
	}

	if !mapKeyGen.validKey(key) {
		http.NotFound(w, req)
		return
	}

	mm, err := e.mdb.GetMeta(key)
	if err != nil {
		log.Printf("can't get map metadata for %q: %v", key, err)
		httpErrorCode(w, http.StatusForbidden)
		return
	}

	switch sub {
	case "":
		var u url.URL
		u = *req.URL
		u.Path += "/"
		http.Redirect(w, req, u.String(), http.StatusMovedPermanently)
	case "/":
		t, err := loadTemplate(filepath.Join(e.resdir, "index.html"))
		if err != nil {
			log.Println(err)
			httpErrorCode(w, http.StatusInternalServerError)
			return
		}
		e.serveEdit(w, req, mm, t)
	default:
		e.base.ServeHTTP(w, requestWithPath(req, sub))
	}
}

const editPassName = "editPass"

func (e *editor) serveNew(w http.ResponseWriter, req *http.Request) {
	mm := mapMeta{
		Title:   "New map",
		ModTime: time.Now(),
	}

	var err error
	mm.Key, err = mapKeyGen.generateKey()
	if err != nil {
		log.Println("can't generate map key:", err)
		httpErrorCode(w, http.StatusInternalServerError)
		return
	}

	mm.EditPass, err = editPassGen.generateKey()
	if err != nil {
		log.Println("can't generate editpass:", err)
		httpErrorCode(w, http.StatusInternalServerError)
		return
	}

	if err := e.mdb.Access(mm.Key); err != nil {
		log.Println("can't store timestamp:", err)
		httpErrorCode(w, http.StatusInternalServerError)
		return
	}

	if err := e.mdb.StoreMeta(mm); err != nil {
		httpErrorCode(w, http.StatusInternalServerError)
		return
	}

	u := fmt.Sprintf("%s%s/?%s=%s", e.prefix, mm.Key, editPassName, mm.EditPass)
	http.Redirect(w, req, u, http.StatusTemporaryRedirect)
}

func (e *editor) serveEdit(w http.ResponseWriter, req *http.Request, mm mapMeta, t *template.Template) {
	if req.Method != "POST" && req.Method != "GET" {
		httpErrorCode(w, http.StatusBadRequest)
		return
	}

	editPass := req.URL.Query().Get(editPassName)
	if !editPassGen.validKey(editPass) {
		log.Printf("invalid editPass")
		httpErrorCode(w, http.StatusForbidden)
		return
	}

	if mm.EditPass != editPass {
		log.Printf("unauthorized editpass for %q: %v != %v", mm.Key, editPass, mm.EditPass)
		httpErrorCode(w, http.StatusForbidden)
		return
	}

	td := struct {
		Title  string
		Errors []string

		Msg []string

		MapTarget string
		MapLink   string
	}{
		Title: mm.Title,
	}

	if req.Method == "POST" {
		e.handlePost(&mm, func(e error) {
			td.Errors = append(td.Errors, e.Error())
		}, req)
		if len(td.Errors) == 0 {
			td.Msg = append(td.Msg, "Upload successful.")
		}
	}

	if mm.PubCount > 0 {
		td.Msg = append(td.Msg, fmt.Sprintf("Map has %d points", mm.PubCount))
		td.MapTarget = fmt.Sprintf("map-%s", mm.Key)
		td.MapLink = fmt.Sprintf("../../map/%s/", mm.Key)
	} else {
		td.Msg = append(td.Msg, "This map is empty.")
	}

	if err := t.Execute(w, td); err != nil {
		log.Println(err)
	}
}

func (e *editor) handlePost(mm *mapMeta, errh func(error), req *http.Request) {
	form, err := parseMultipartForm(req, 1<<20, func(formName string) (maxLen, maxMultiLen int64) {
		switch formName {
		case "listtxt", "iconstyle", "mapstyle":
			return 1 << 20, 0
		}
		return 0, 0
	})
	if err != nil {
		errh(err)
		return
	}

	mm.ModTime = time.Now()

	if t := form.Values.Get("title"); t != "" {
		mm.Title = t
	}

	batch := e.mdb.db.Batch()
	e.handleUIMapSave(mm, batch, form, errh)

	if f, ok := form.File("mapstyle"); ok {
		var l []interface{}
		err := json.Unmarshal(f.Content, &l)
		if err != nil {
			errh(errors.Wrap(err, "map style"))
		} else {
			batch.Set(mm.styleKey(), f.Content)
		}
	}

	p, err := json.Marshal(mm)
	if err != nil {
		errh(errors.Wrap(err, "can't marshal map metadata"))
		return
	}
	batch.Set("meta|"+mm.Key, p)

	if err := batch.Commit(); err != nil {
		errh(errors.Wrap(err, "map save error"))
	}
}

func (e *editor) handleUIMapSave(mm *mapMeta, batch keyvalue.Batch, form *multipartForm, errh func(error)) {
	listFile, newList := form.File("listtxt")
	styleFile, newStyle := form.File("iconstyle")

	if !newList && !newStyle {
		return
	}

	db := e.mdb.db

	var listBytes []byte
	listKey := "list|" + mm.Key
	if newList {
		listBytes = listFile.Content
	} else {
		var err error
		listBytes, err = db.Get(listKey)
		if err != nil {
			errh(err)
		}
	}

	pubs, err := parsePubList(bytes.NewReader(listBytes), e.gc, func(err error) error {
		errh(err)
		return nil
	})
	if len(pubs) == 0 {
		errh(errors.New("empty pub list"))
		if newList {
			// can't generate UI map for empty list
			return
		}
	}

	mm.PubCount = len(pubs)

	if newList {
		batch.Set(listKey, listBytes)
	}

	var styleBytes []byte
	styleKey := "iconstyle|" + mm.Key
	if newStyle {
		styleBytes = styleFile.Content
	} else {
		var err error
		styleBytes, err = db.Get(styleKey)
		if err != nil {
			errh(err)
		}
	}

	styler, err := NewStyler(bytes.NewReader(styleBytes), e.fontSrc)
	if err != nil {
		errh(err)
	}
	if styler == nil {
		if e.defaultIconRenderer == nil {
			// can't render icons
			return
		}
		styler = &Styler{
			r: e.defaultIconRenderer,
		}
	}

	if newStyle {
		batch.Set(styleKey, styleBytes)
	}

	pfx := "path|" + mm.Key + "/"
	it := db.Iterator(pfx, "")
	defer it.Close()
	for it.Next() {
		if !strings.HasPrefix(it.Key(), pfx) {
			break
		}
		batch.Delete(it.Key())
	}
	if err := it.Err(); err != nil {
		errh(err)
	}

	batch.Set("path|"+path.Join(mm.Key, pubjson), pubListJSON(pubs, ""))
	for _, p := range pubs {
		data, err := pubIconData(p, styler)
		if err != nil {
			log.Fatal(err)
		}
		batch.Set("path|"+path.Join(mm.Key, p.IconBasename()), data)
	}
}
