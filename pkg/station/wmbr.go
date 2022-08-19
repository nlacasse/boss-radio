package station

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"sync"

	"golang.org/x/text/encoding/ianaindex"
)

//go:embed images/wmbr.gif
var wmbrLogoBytes []byte

type wmbrInfo struct {
	XMLName   xml.Name `xml:"wmbrinfo"`
	Showname  string   `xml:"showname_ascii"`
	Showhosts string   `xml:showhosts_ascii"`
	Temp      string   `xml:"temp"`
	Weather   string   `xml:"wx"`
}

type Wmbr struct {
	logo image.Image

	mu  sync.Mutex
	cmd *exec.Cmd
}

var _ Station = (*Wmbr)(nil)

func NewWmbr() *Wmbr {
	logo, _, err := image.Decode(bytes.NewReader(wmbrLogoBytes))
	if err != nil {
		log.Fatalf("Could not decode WMBR logo: %w", err)
	}

	return &Wmbr{
		logo: logo,
	}
}

func (wmbr *Wmbr) Name() string {
	return "WMBR"
}

func (wmbr *Wmbr) Logo() image.Image {
	return wmbr.logo
}

func (wmbr *Wmbr) StreamCmd() *exec.Cmd {
	str := "http://wmbr.org:8000/hi"
	return exec.Command("mpv", "-no-video", str)
}

func (wmbr *Wmbr) Status() Status {
	var s Status

	resp, err := http.Get("https://wmbr.org/cgi-bin/xmlinfo")
	if err != nil {
		s.Show = err.Error()
		return s
	}
	defer resp.Body.Close()

	var wi wmbrInfo
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = func(charset string, reader io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, fmt.Errorf("charset %s: %s", charset, err.Error())
		}
		if enc == nil {
			// Assume it's compatible with (a subset of) UTF-8 encoding
			// Bug: https://github.com/golang/go/issues/19421
			return reader, nil
		}
		return enc.NewDecoder().Reader(reader), nil
	}
	if err := decoder.Decode(&wi); err != nil {
		log.Print(err)
		s.Show = err.Error()
		return s
	}

	re := regexp.MustCompile("[[:^ascii:]]")

	return Status{
		Show:   wi.Showname,
		Artist: wi.Showhosts,
		Track:  re.ReplaceAllLiteralString(wi.Temp, ""),
		Album:  wi.Weather,
	}
}
