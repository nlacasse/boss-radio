package station

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/gif"
	"log"
	"os/exec"
	"sync"
)

//go:embed images/kxlu.gif
var kxluLogoBytes []byte

type kxluStatus struct {
	AirName    string `json:"air_name"`
	Album      string `json:"album"`
	Artist     string `json:"artist"`
	TrackTitle string `json:"track_title"`
}

type Kxlu struct {
	logo image.Image

	mu  sync.Mutex
	cmd *exec.Cmd
}

var _ Station = (*Kxlu)(nil)

func NewKxlu() *Kxlu {
	logo, _, err := image.Decode(bytes.NewReader(kxluLogoBytes))
	if err != nil {
		log.Fatalf("Could not decode KXLU logo: %w", err)
	}

	return &Kxlu{
		logo: logo,
	}
}

func (kxlu *Kxlu) Name() string {
	return "KXLU"
}

func (kxlu *Kxlu) Logo() image.Image {
	return kxlu.logo
}

func (kxlu *Kxlu) Stream() string {
	return "https://kxlu.streamguys1.com/kxlu-hi"
}

func (kxlu *Kxlu) Status() Status {
	var s Status
	return s
}
