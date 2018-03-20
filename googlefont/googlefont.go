package googlefont

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type Client struct {
	*http.Client
}

func (r *Client) Get(family string, weight int) ([]byte, error) {
	u, err := url.Parse("https://fonts.googleapis.com/css")
	if err != nil {
		panic("impossible")
	}

	uv := make(url.Values)
	uv.Set("family", fmt.Sprintf("%s:%d", family, weight))
	u.RawQuery = uv.Encode()

	css, err := r.getBody(u.String())
	if err != nil {
		return nil, err
	}

	fontUrl, err := cssFontUrl(css)
	if err != nil {
		return nil, err
	}

	return r.getBody(fontUrl)
}

func (r *Client) getBody(url string) ([]byte, error) {
	c := r.Client
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

var defaultClient = &Client{}

func Get(family string, weight int) ([]byte, error) {
	return defaultClient.Get(family, weight)
}

var cssFontRe = regexp.MustCompile(`@font-face\s*{[^}]*src:[^;]*(url\([^;]*)`)

func cssFontUrl(css []byte) (string, error) {
	r := cssFontRe.FindAllSubmatch(css, 1)
	if len(r) != 1 {
		return "", errors.Errorf("css parse error: multiple matches")
	}

	s := string(r[0][1])
	s = strings.TrimPrefix(s, "url(")
	if len(s) == 0 {
		return "", errors.Errorf("css parse error: unexpected eof")
	}
	if s[0] == '\'' || s[0] == '"' {
		return "", errors.Errorf("css parse error: not implemented")
	}

	i := strings.IndexRune(s, ')')
	if i < 0 {
		return "", errors.Errorf("css parse error: unexpected eof")
	}

	return s[:i], nil
}

/*

https://fonts.googleapis.com/css?family=Roboto:500

@font-face {
  font-family: 'Roboto';
  font-style: normal;
  font-weight: 500;
  src: local('Roboto Medium'), local('Roboto-Medium'), url(https://fonts.gstatic.com/s/roboto/v18/KFOlCnqEu92Fr1MmEU9fBBc9.ttf) format('truetype');
}

*/
