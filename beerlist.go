package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/tajtiattila/geocode"
)

func parseMultiLine(r io.Reader, f func(lineno int, lines []string) error) error {
	scanner := bufio.NewScanner(r)
	var blk []string
	lineno := 1
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if err := f(lineno, blk); err != nil {
				return err
			}
			blk = blk[:0]
		} else {
			blk = append(blk, line)
		}
		lineno++
	}
	if len(blk) != 0 {
		if err := f(lineno, blk); err != nil {
			return err
		}
	}
	return scanner.Err()
}

type LatLong struct {
	Lat  float64
	Long float64
}

type Pub struct {
	Label string
	Title string
	Addr  string
	Geo   LatLong
	Tags  []string
	Desc  []string
}

func (p Pub) Has(tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (p Pub) String() string {
	buf := new(bytes.Buffer)
	p.WriteTo(buf)
	return buf.String()
}

func (p Pub) WriteTo(w io.Writer) (n int, err error) {
	prt := func(format string, v ...interface{}) {
		if err == nil {
			var m int
			m, err = fmt.Fprintf(w, format, v...)
			n += m
		}
	}
	prt("[%s] %s\n", p.Label, p.Title)
	prt("(%s)\n", p.Addr)
	if len(p.Tags) != 0 {
		prt("%s\n", strings.Join(p.Tags, " "))
	}
	for _, d := range p.Desc {
		prt("%s\n", d)
	}
	prt("\n")
	return n, err
}

// IconBasename returns the icon filename to use with this pub.
func (p Pub) IconBasename() string {
	return fmt.Sprintf("icon-%s.png", p.Label)
}

func parsePubList(r io.Reader, gc geocode.Geocoder, errh func(err error) error) ([]Pub, error) {
	var pubs []Pub
	seen := make(map[string]struct{})
	err := parseMultiLine(r, func(lineno int, v []string) error {
		var title, addr, tags string
		var rest []string
		for _, line := range v {
			if len(line) == 0 {
				continue
			}
			switch {
			case line[0] == '[' && title == "":
				title = line
			case line[0] == '(' && addr == "":
				addr = line
			case line[0] == '#':
				tags += " " + line
			default:
				rest = append(rest, line)
			}
		}

		if title == "" || addr == "" {
			return errh(errors.Errorf("line %d: missing title/addr", lineno))
		}

		p, err := parsePub(gc, title, addr, tags, rest)
		if err != nil {
			return errh(errors.Wrapf(err, "line %d", lineno))
		}

		if _, dup := seen[p.Label]; dup {
			return errh(errors.Errorf("line %d: duplicate label %q", lineno, p.Label))
		}
		seen[p.Label] = struct{}{}

		pubs = append(pubs, p)
		return nil
	})
	if err != nil {
		err = errh(err)
	}
	return pubs, err
}

func parsePub(gc geocode.Geocoder, title, addr, tags string, rest []string) (Pub, error) {
	var p Pub

	i := strings.IndexRune(title, ']')
	if title[0] != '[' || i < 0 {
		return Pub{}, errors.Errorf("error parsing title %q", title)
	}
	p.Label = strings.TrimSpace(title[1:i])

	p.Title = strings.TrimSpace(title[i+1:])

	p.Addr = strings.TrimSpace(strings.TrimRight(strings.TrimLeft(addr, "("), ")"))

	r, err := gc.Geocode(p.Addr)
	if err != nil {
		return Pub{}, errors.Wrapf(err, "geocode failed for %q", p.Addr)
	}
	p.Geo.Lat = r.Lat
	p.Geo.Long = r.Long

	p.Tags = strings.Fields(tags)
	p.Desc = rest

	return p, nil
}
