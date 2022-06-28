package station

import (
	"image"
	"os/exec"
)

var AllStations = []Station{
	NewKfjc(),
	NewKxlu(),
	NewWfmu(),
	NewWmbr(),
	NewNts1(),
	NewNts2(),
	NewAporee(),
	NewBluetooth(),
}

type Station interface {
	Name() string

	Logo() image.Image

	StreamCmd() *exec.Cmd

	Status() Status
}

type Status struct {
	Show   string
	Album  string
	Artist string
	Track  string
}
