package main

import (
	_ "image/gif"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nlacasse/boss-radio/pkg/bradio"
	"github.com/nlacasse/boss-radio/pkg/station"
	"periph.io/x/host/v3"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatalf("host.Init failed: %v", err)
	}

	br, err := bradio.NewBossRadio(station.AllStations)
	if err != nil {
		log.Fatalf("NewBossRadio() failed: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		br.Destroy()
		os.Exit(0)
	}()

	if err := br.Run(); err != nil {
		log.Fatal(err)
	}
}
