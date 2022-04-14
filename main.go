package main

import (
	"fmt"
	_ "image/gif"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	volume "github.com/itchyny/volume-go"
	"periph.io/x/host/v3"

	"github.com/nlacasse/boss-radio/pkg/button"
	"github.com/nlacasse/boss-radio/pkg/screen"
	"github.com/nlacasse/boss-radio/pkg/station"
)

func loop(scrn *screen.Screen, btn *button.Button) error {
	pushChan, err := btn.Listen()
	if err != nil {
		return fmt.Errorf("Button.Listen failed: %w", err)
	}

	var (
		on     bool
		stnIdx int
		cmd    *exec.Cmd
	)
	stn := station.Stations[stnIdx]

	defer func() {
		if cmd != nil {
			cmd.Process.Kill()
		}
	}()

	// TODO: Move this.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if cmd != nil {
			cmd.Process.Kill()
		}
		os.Exit(0)
	}()

	for push := range pushChan {
		var (
			err        error
			volChanged bool
			stnChanged bool
		)
		switch push {
		case button.Center:
			on = !on
			if !on {
				cmd.Process.Kill()
				cmd = nil
				scrn.Clear()
				continue
			}
			stnChanged = true
		case button.Left:
			if !on {
				continue
			}
			stnIdx = (stnIdx - 1) % len(station.Stations)
			if stnIdx < 0 {
				stnIdx += len(station.Stations)
			}
			stn = station.Stations[stnIdx]
			stnChanged = true
		case button.Right:
			if !on {
				continue
			}
			stnIdx = (stnIdx + 1) % len(station.Stations)
			stnChanged = true
			stn = station.Stations[stnIdx]
		case button.Up:
			if !on {
				continue
			}
			err = volume.IncreaseVolume(5)
			volChanged = true
		case button.Down:
			if !on {
				continue
			}
			err = volume.IncreaseVolume(-5)
			volChanged = true
		}
		if err != nil {
			return err
		}

		if volChanged {
			// Flash new volume.
			vol, err := volume.GetVolume()
			if err != nil {
				return err
			}
			scrn.SetText([6]string{
				"",
				"",
				fmt.Sprintf("Volume: %d", vol),
			})
			scrn.Draw()
			time.Sleep(250 * time.Millisecond)
		}

		if stnChanged {
			if cmd != nil {
				cmd.Process.Kill()
			}
			cmd = exec.Command("mpv", "-no-video", stn.Stream())
			log.Printf("COMMAND: %+v", cmd)
			cmd.Stdout = os.Stdout
			if err := cmd.Start(); err != nil {
				return err
			}

			// Flash new station logo for a second.
			scrn.DrawImage(stn.Logo())
			time.Sleep(250 * time.Millisecond)
		}

		// Back to station screen.
		status := stn.Status()
		scrn.SetText([6]string{
			"     " + stn.Name(),
			"",
			status.Show,
			status.Artist,
			status.Track,
			status.Album,
		})
		scrn.Draw()
	}
	panic("unreachable")
}

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal("host.Init failed: %v", err)
	}

	scrn, err := screen.New()
	if err != nil {
		log.Fatalf("screen.New failed: %v", err)
	}

	btn := button.New()
	if err := loop(scrn, btn); err != nil {
		log.Fatal("loog failed: %v", err)
	}
}
