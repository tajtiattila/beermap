package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"
)

func httpErrorCode(w http.ResponseWriter, errc int) {
	m := http.StatusText(errc)
	if m == "" {
		m = fmt.Sprint("Error", errc)
	}
	http.Error(w, m, errc)
}

func requestWithPath(r *http.Request, path string) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	r2.URL.Path = path
	return r2
}

func loadTemplate(path string) (*template.Template, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read template")
	}

	t, err := template.New(filepath.Base(path)).Parse(string(src))
	if err != nil {
		return nil, errors.Wrapf(err, "parse template %q", path)
	}

	return t, nil
}

func serveTemplateWithData(w http.ResponseWriter, req *http.Request,
	path string, data interface{}) {

	src, err := ioutil.ReadFile(path)
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
}

func serveDirTemplate(path string, data interface{}) http.Handler {
	def := http.FileServer(http.Dir(path))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			def.ServeHTTP(w, req)
			return
		}

		serveTemplateWithData(w, req, filepath.Join(path, "index.html"), data)
	})
}
