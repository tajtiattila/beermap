package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type multipartFile struct {
	Filename string
	Content  []byte
}

type multipartForm struct {
	Values url.Values
	Files  map[string][]multipartFile
}

func (m multipartForm) File(name string) (f multipartFile, ok bool) {
	v, ok := m.Files[name]
	if ok {
		return v[0], true
	}
	return multipartFile{}, false
}

// wantFormFile reports if a file with formName is wanted.
// if maxLen == 0, the parseMultipartForm will return an unwanted file error.
// if maxMultiLen > 0, then multiple files are accepted up to the specified
// max total length.
type wantFormFile func(name string) (maxLen, maxMultiLen int64)

func parseMultipartForm(req *http.Request, maxValueMem int64, f wantFormFile) (*multipartForm, error) {

	mr, err := req.MultipartReader()
	if err != nil {
		return nil, err
	}

	form := new(multipartForm)
	sumLen := make(map[string]int64)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		name := part.FormName()
		if name == "" {
			continue
		}
		filename := part.FileName()

		_, hasContentTypeHeader := part.Header["Content-Type"]
		if !hasContentTypeHeader && filename == "" {
			// value, store as string in memory
			var buf bytes.Buffer
			n, err := io.CopyN(&buf, part, maxValueMem+1)
			if err != nil && err != io.EOF {
				return nil, err
			}
			maxValueMem -= n
			if maxValueMem < 0 {
				return nil, errors.New("form message too large")
			}
			if form.Values == nil {
				form.Values = make(url.Values)
			}
			form.Values.Add(name, buf.String())
		} else {
			ml, mml := f(name)
			if ml == 0 {
				return nil, errors.Errorf("unwanted form file %q received", name)
			}
			if mml == 0 {
				if len(form.Files[name]) > 0 {
					return nil, errors.Errorf("duplicate form file %q received", name)
				}
				mml = ml
			}
			room := mml - sumLen[name]
			var buf bytes.Buffer
			n, err := io.CopyN(&buf, part, room+1)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if n > room {
				return nil, errors.Errorf("form file %q too long", name)
			}
			if n == 0 {
				continue
			}
			if form.Files == nil {
				form.Files = make(map[string][]multipartFile)
			}
			form.Files[name] = append(form.Files[name], multipartFile{
				Filename: filename,
				Content:  buf.Bytes(),
			})
			sumLen[name] += n
		}
	}
	return form, nil
}
