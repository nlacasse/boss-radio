package station

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	"io"
	"log"
	"net/http"
	"os/exec"
)

//go:embed images/nts1.gif
var nts1LogoBytes []byte

//go:embed images/nts2.gif
var nts2LogoBytes []byte

type ntsGenre struct {
	Value string `json:"value"`
}

type ntsDetails struct {
	Location string     `json:"location_long"`
	Name     string     `json:"name"`
	Genres   []ntsGenre `json:"genres"`
}

type ntsEmbeds struct {
	Details ntsDetails `json:"details"`
}

type ntsShow struct {
	Embeds ntsEmbeds `json:"embeds"`
}

type ntsResult struct {
	Now ntsShow `json:"now"`
}

type ntsStatus struct {
	Results []ntsResult `json:"results"`
}

type Nts struct {
	channel int
	logo    image.Image
	stream  string
}

var _ Station = (*Nts)(nil)

func NewNts1() *Nts {
	logo, _, err := image.Decode(bytes.NewReader(nts1LogoBytes))
	if err != nil {
		log.Fatalf("Could not decode NTS1 logo: %w", err)
	}

	return &Nts{
		channel: 1,
		logo:    logo,
		stream:  "https://stream-relay-geo.ntslive.net/stream",
	}
}

func NewNts2() *Nts {
	logo, _, err := image.Decode(bytes.NewReader(nts2LogoBytes))
	if err != nil {
		log.Fatalf("Could not decode NTS2 logo: %w", err)
	}

	return &Nts{
		channel: 2,
		logo:    logo,
		stream:  "https://stream-relay-geo.ntslive.net/stream2",
	}
}

func (nts *Nts) Name() string {
	return fmt.Sprintf("NTS %d", nts.channel)
}

func (nts *Nts) Logo() image.Image {
	return nts.logo
}

func (nts *Nts) StreamCmd() *exec.Cmd {
	return exec.Command("mpv", "-no-video", nts.stream)
}

func (nts *Nts) Status() Status {
	var s Status

	resp, err := http.Get("https://www.nts.live/api/v2/live")
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

	var ns ntsStatus
	if err := json.Unmarshal(body, &ns); err != nil {
		s.Show = err.Error()
		return s
	}

	if len(ns.Results) < 2 {
		s.Show = "Not enough results"
		return s
	}

	dets := ns.Results[nts.channel-1].Now.Embeds.Details
	var genre string
	if len(dets.Genres) > 0 {
		genre = dets.Genres[0].Value
	}

	return Status{
		Show:   dets.Name,
		Artist: "",
		Track:  dets.Location,
		Album:  genre,
	}
}
