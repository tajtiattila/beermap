package googlefont_test

import (
	"testing"

	"github.com/tajtiattila/beermap/googlefont"
)

const cssSample = `
@font-face {
  font-family: 'Roboto';
  font-style: normal;
  font-weight: 500;
  src: local('Roboto Medium'), local('Roboto-Medium'), url(https://fonts.gstatic.com/s/roboto/v18/KFOlCnqEu92Fr1MmEU9fBBc9.ttf) format('truetype');
}
`

func TestCSS(t *testing.T) {
	got, err := googlefont.Exp_cssFontUrl([]byte(cssSample))
	if err != nil {
		t.Fatal(err)
	}
	const want = "https://fonts.gstatic.com/s/roboto/v18/KFOlCnqEu92Fr1MmEU9fBBc9.ttf"
	if got != want {
		t.Fatalf("got font url %q, want %q", got, want)
	}
}

func TestFetchFont(t *testing.T) {
	_, err := googlefont.Get("Open Sans", 400)
	if err != nil {
		t.Fatal(err)
	}
}
