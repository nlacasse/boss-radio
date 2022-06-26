package main

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"log"
	"os"

	"github.com/nlacasse/boss-radio/pkg/screen"
	"periph.io/x/host/v3"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("disp <image>")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("image.Decode: %v", err)
	}

	if _, err := host.Init(); err != nil {
		log.Fatalf("host.Init failed: %v", err)
	}
	scrn, err := screen.New()
	if err != nil {
		log.Fatalf("screen.New: %v", err)
	}

	scrn.DrawImage(img)
}
