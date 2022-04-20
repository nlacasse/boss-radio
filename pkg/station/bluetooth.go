package station

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/gif"
	"log"
	"os/exec"
	"strings"
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
	sc := `bluetoothctl paired-devices |
cut -f2 -d' '|
while read -r uuid
do
    info=$(bluetoothctl info $uuid)
	if echo "$info" | grep -q "Connected: yes"; then
	   echo "$info" | grep "Name" | cut -f2 -d' '
   fi
done
`

	cmd := exec.Command("bash", "-c", sc)
	out, err := cmd.Output()
	log.Printf("got bluetooth status %q %v", string(out), err)
	if err != nil {
		log.Printf("bluetooth status failed: %v", err)
		return Status{}
	}

	return Status{
		Show: strings.TrimSpace(string(out)),
	}
}
