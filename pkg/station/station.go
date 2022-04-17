package station

import "image"

var AllStations = []Station{
	NewKfjc(),
	NewKxlu(),
	NewWfmu(),
	NewWmbr()}

type Station interface {
	Name() string

	Logo() image.Image

	Stream() string

	Status() Status
}

type Status struct {
	Show   string
	Album  string
	Artist string
	Track  string
}
