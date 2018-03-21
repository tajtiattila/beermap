package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type KMZ struct {
	Title      string
	Placemarks []kPlacemark

	seq int

	closer io.Closer
	z      *zip.Writer
}

func NewKMZ(w io.Writer, title string) *KMZ {
	return &KMZ{
		Title: title,
		z:     zip.NewWriter(w),
	}
}

func (k *KMZ) Close() error {
	f, err := k.z.Create("doc.kml")
	if err != nil {
		return err
	}
	if err := kmltmp.Execute(f, k); err != nil {
		return err
	}
	return k.z.Close()
}

type kPlacemark struct {
	Placemark
	StyleID  string
	IconPath string
}

type Placemark struct {
	Title     string
	Desc      string
	Lat, Long float64
}

func (k *KMZ) IconPlacemark(png []byte, pm Placemark) error {
	path := fmt.Sprintf("images/icon-%d.png", k.seq)
	style := fmt.Sprintf("icon-%d-BEE1157", k.seq)
	k.seq++

	f, err := k.z.Create(path)
	if err != nil {
		return errors.Wrap(err, "can't create image file in zip")
	}

	if _, err := f.Write(png); err != nil {
		return errors.Wrap(err, "can't write image into zip")
	}

	k.Placemarks = append(k.Placemarks, kPlacemark{
		Placemark: pm,
		StyleID:   style,
		IconPath:  path,
	})
	return nil
}

func needcdata(s string) bool {
	for _, r := range s {
		switch r {
		case '<', '>', '&', '\'', '"', '\r', '\n':
			return true
		}
	}
	return false
}

func xmlCharData(s string) string {
	if !needcdata(s) {
		return s
	}
	const cdata = "<![CDATA["
	const cend = "]]>"
	res := new(bytes.Buffer)
	f := func(s string) {
		s = strings.Replace(s, "\r", "", -1)
		s = strings.Replace(s, "\n", "<br>", -1)
		res.WriteString(cdata)
		res.WriteString(s)
		res.WriteString(cend)
	}

	for {
		i := strings.Index(s, cend)
		if i < 0 {
			break
		}
		f(s[:i+1])
		s = s[i+1:]
	}
	f(s)

	return res.String()
}

var kmltmp = template.Must(template.New("kmz").Funcs(template.FuncMap{
	"xmlCharData": xmlCharData,
}).Parse(`<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <name>{{.Title}}</name>
{{range .Placemarks}}{{if .StyleID}}
    <Style id="{{.StyleID}}-normal">
      <IconStyle>
        <scale>1</scale>
        <Icon>
          <href>{{.IconPath}}</href>
        </Icon>
      </IconStyle>
      <LabelStyle>
        <scale>0</scale>
      </LabelStyle>
    </Style>
    <Style id="{{.StyleID}}-highlight">
      <IconStyle>
        <scale>1</scale>
        <Icon>
          <href>{{.IconPath}}</href>
        </Icon>
      </IconStyle>
      <LabelStyle>
        <scale>1</scale>
      </LabelStyle>
    </Style>
    <StyleMap id="{{.StyleID}}">
      <Pair>
        <key>normal</key>
        <styleUrl>#{{.StyleID}}-normal</styleUrl>
      </Pair>
      <Pair>
        <key>highlight</key>
        <styleUrl>#{{.StyleID}}-highlight</styleUrl>
      </Pair>
    </StyleMap>
{{end}}{{end}}
{{range .Placemarks}}
    <Placemark>
      <name>{{.Title | xmlCharData}}</name>
{{- if .Desc}}
      <description>{{.Desc | xmlCharData}}</description>
{{- end}}
{{- if .StyleID}}
      <styleUrl>#{{.StyleID}}</styleUrl>
{{- end}}
      <Point>
        <coordinates>
			{{.Long}},{{.Lat}},0
        </coordinates>
      </Point>
    </Placemark>
{{end}}
  </Document>
</kml>
`))

/*
<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <name>Prágai kocsmák - számos</name>


    <Style id="icon-seq2-0-0-0288D1-normal">
      <IconStyle>
        <scale>1</scale>
        <Icon>
          <href>images/icon-1.png</href>
        </Icon>
      </IconStyle>
      <LabelStyle>
        <scale>0</scale>
      </LabelStyle>
    </Style>
    <Style id="icon-seq2-0-0-0288D1-highlight">
      <IconStyle>
        <scale>1</scale>
        <Icon>
          <href>images/icon-1.png</href>
        </Icon>
      </IconStyle>
      <LabelStyle>
        <scale>1</scale>
      </LabelStyle>
    </Style>
    <StyleMap id="icon-seq2-0-0-0288D1">
      <Pair>
        <key>normal</key>
        <styleUrl>#icon-seq2-0-0-0288D1-normal</styleUrl>
      </Pair>
      <Pair>
        <key>highlight</key>
        <styleUrl>#icon-seq2-0-0-0288D1-highlight</styleUrl>
      </Pair>
    </StyleMap>

	...

    <Placemark>
      <name>[001] U Zlatého Tygra</name>
      <description><![CDATA[[001] U Zlatého Tygra - az Arany Tigrishez<br>(Husova 17, Praha 1-Staré Město) <br>H-V:15-23<br>http://www.uzlatehotygra.cz/<br>- a legjobb pilseni itt van (A söröző leghíresebb törzsvendége Bohumil Hrabal író volt), nyitás előtt pár perccel már tobzódni kell, különben nem férünk be<br>http://www.pragaisorozok.hu/index.php?menu=sorozok&varosresz=stare-mesto-ovaros&id=u-zlateho-tygra-az-arany-tigrishez]]></description>
      <styleUrl>#icon-seq2-0-0-0288D1</styleUrl>
      <Point>
        <coordinates>
          14.418045,50.085772,0
        </coordinates>
      </Point>
    </Placemark>

  </Document>
</kml>
*/
