package station

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image"
	_ "image/gif"
	"io"
	"log"
	"net/http"
	"os/exec"
)

//go:embed images/aporee.gif
var aporeeLogoBytes []byte

type aporeeStatus struct {
	Date  string `json:"aporee_date"`
	Lat   string `json:"aporee_lat"`
	Lng   string `json:"aporee_lng"`
	Title string `json:"aporee_title"`
}

type Aporee struct {
	logo image.Image
}

var _ Station = (*Aporee)(nil)

func NewAporee() *Aporee {
	logo, _, err := image.Decode(bytes.NewReader(aporeeLogoBytes))
	if err != nil {
		log.Fatalf("Could not decode Aporee logo: %w", err)
	}

	return &Aporee{
		logo: logo,
	}
}

func (aporee *Aporee) Name() string {
	return "Aporee"
}

func (aporee *Aporee) Logo() image.Image {
	return aporee.logo
}

func (aporee *Aporee) StreamCmd() *exec.Cmd {
	str := "http://radio.aporee.org:8000/aporee_high.m3u"
	return exec.Command("mpv", "-no-video", str)
}

func (aporee *Aporee) Status() Status {
	var s Status

	resp, err := http.Get("https://radio.aporee.org/spool/meta.js")
	if err != nil {
		s.Show = err.Error()
		return s
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Show = err.Error()
		return s
	}

	var as aporeeStatus
	if err := json.Unmarshal(body, &as); err != nil {
		s.Show = err.Error()
		return s
	}

	return Status{
		Show:  as.Title,
		Track: as.Lat,
		Album: as.Lng,
	}
}
