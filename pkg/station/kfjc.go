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
	"sync"
)

//go:embed images/kfjc-devil.gif
var kfjcLogoBytes []byte

type kfjcStatus struct {
	AirName    string `json:"air_name"`
	Album      string `json:"album"`
	Artist     string `json:"artist"`
	TrackTitle string `json:"track_title"`
}

type Kfjc struct {
	logo image.Image

	mu  sync.Mutex
	cmd *exec.Cmd
}

var _ Station = (*Kfjc)(nil)

func NewKfjc() *Kfjc {
	logo, _, err := image.Decode(bytes.NewReader(kfjcLogoBytes))
	if err != nil {
		log.Fatalf("Could not decode KFJC logo: %w", err)
	}

	return &Kfjc{
		logo: logo,
	}
}

func (kfjc *Kfjc) Name() string {
	return "KFJC"
}

func (kfjc *Kfjc) Logo() image.Image {
	return kfjc.logo
}

func (kfjc *Kfjc) Stream() string {
	return "http://netcast.kfjc.org/kfjc-320k-aac"
}

func (kfjc *Kfjc) Status() Status {
	var s Status

	resp, err := http.Get("https://kfjc.org/api/playlists/current.php")
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

	var ks kfjcStatus
	if err := json.Unmarshal(body, &ks); err != nil {
		s.Show = err.Error()
		return s
	}

	return Status{
		Show:   ks.AirName,
		Album:  ks.Album,
		Artist: ks.Artist,
		Track:  ks.TrackTitle,
	}
}
