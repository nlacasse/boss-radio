package station

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/gif"
	"log"
	"os/exec"
)

//go:embed images/bluetooth.gif
var bluetoothLogoBytes []byte

type Bluetooth struct {
	logo image.Image
}

var _ Station = (*Bluetooth)(nil)

func NewBluetooth() *Bluetooth {
	logo, _, err := image.Decode(bytes.NewReader(bluetoothLogoBytes))
	if err != nil {
		log.Fatalf("Could not decode Bluetooth logo: %w", err)
	}

	return &Bluetooth{
		logo: logo,
	}
}

func (bt *Bluetooth) Name() string {
	return "Bluetooth"
}

func (bt *Bluetooth) Logo() image.Image {
	return bt.logo
}

func (bt *Bluetooth) StreamCmd() *exec.Cmd {
	return exec.Command("bluealsa-aplay")
}

func (bt *Bluetooth) Status() Status {
	return Status{}
}
