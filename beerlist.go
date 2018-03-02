package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/pkg/errors"
	"github.com/tajtiattila/geocode"
)

func parseMultiLine(r io.Reader, f func([]string) error) error {
	scanner := bufio.NewScanner(r)
	var blk []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if err := f(blk); err != nil {
				return err
			}
			blk = blk[:0]
		} else {
			blk = append(blk, line)
		}
	}
	if len(blk) != 0 {
		if err := f(blk); err != nil {
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
	Num   int
	Title string
	Addr  string
	Geo   LatLong
	Tags  []string
	Desc  []string
}

func (p Pub) NumStr() string {
	return fmt.Sprintf("%03d", p.Num)
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
	prt("[%03d] %s\n", p.Num, p.Title)
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

func parsePubList(r io.Reader, gc geocode.Geocoder) ([]Pub, error) {
	var pubs []Pub
	err := parseMultiLine(r, func(v []string) error {
		var title, addr, tags string
		var rest []string
		for _, line := range v {
			if len(line) == 0 {
				continue
			}
			switch line[0] {
			case '[':
				title = line
			case '(':
				addr = line
			case '#':
				tags = line
			default:
				rest = append(rest, line)
			}
		}

		if title == "" || addr == "" {
			log.Println("missing title/addr")
			return nil
		}

		p, err := parsePub(gc, title, addr, tags, rest)
		if err != nil {
			log.Println(err)
			return nil
		}

		pubs = append(pubs, p)
		return nil
	})
	return pubs, err
}

func parsePub(gc geocode.Geocoder, title, addr, tags string, rest []string) (Pub, error) {
	var p Pub

	if _, err := fmt.Sscanf(title, "[%d]%s", &p.Num, &p.Title); err != nil {
		return Pub{}, errors.Wrapf(err, "error parsing title %q", title)
	}
	p.Title = strings.TrimSpace(p.Title)

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
