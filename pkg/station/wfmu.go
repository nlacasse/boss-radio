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

//go:embed images/wfmu.gif
var wfmuLogoBytes []byte

type wfmuStatus struct {
	Show   string `json:"show"`
	Album  string `json:"album"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type Wfmu struct {
	logo image.Image
}

var _ Station = (*Wfmu)(nil)

func NewWfmu() *Wfmu {
	logo, _, err := image.Decode(bytes.NewReader(wfmuLogoBytes))
	if err != nil {
		log.Fatalf("Could not decode WFMU logo: %w", err)
	}

	return &Wfmu{
		logo: logo,
	}
}

func (wfmu *Wfmu) Name() string {
	return "WFMU"
}

func (wfmu *Wfmu) Logo() image.Image {
	return wfmu.logo
}

func (wfmu *Wfmu) StreamCmd() *exec.Cmd {
	str := "https://wfmu.org/wfmu.pls"
	return exec.Command("mpv", "-no-video", str)
}

func (wfmu *Wfmu) Status() Status {
	var s Status

	resp, err := http.Get("https://wfmu.org/wp-content/themes/wfmu-theme/status/main.json")
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

	var f interface{}
	if err := json.Unmarshal(body, &f); err != nil {
		s.Show = err.Error()
		return s
	}

	m, ok := f.(map[string]interface{})
	if !ok {
		return s
	}

	if show, ok := m["show"].(string); ok {
		s.Show = show
	}
	if album, ok := m["album"].(string); ok {
		s.Album = album
	}
	if artist, ok := m["artist"].(string); ok {
		s.Artist = artist
	}
	if track, ok := m["title"].(string); ok {
		s.Artist = track
	}
	return s

}
