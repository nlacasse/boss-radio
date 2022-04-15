package bradio

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/itchyny/volume-go"
	"github.com/nlacasse/boss-radio/pkg/button"
	"github.com/nlacasse/boss-radio/pkg/screen"
	"github.com/nlacasse/boss-radio/pkg/station"
)

type state int

const (
	stateOff state = iota
	stateOn
)

type dialTurn int

var (
	dialTurnLeft  dialTurn = -1
	dialTurnRight dialTurn = 1
)

type BossRadio struct {
	// immutable
	btn  *button.Button
	scrn *screen.Screen

	state  state
	cmd    *exec.Cmd
	stns   []station.Station
	stnIdx int
}

func NewBossRadio(stns []station.Station) (*BossRadio, error) {
	scrn, err := screen.New()
	if err != nil {
		return nil, fmt.Errorf("screen.New failed: %v", err)
	}

	return &BossRadio{
		btn:   button.New(),
		scrn:  scrn,
		state: stateOff,
		stns:  stns,
	}, nil

}

func (br *BossRadio) turnDial(dt dialTurn) error {
	br.stnIdx = (br.stnIdx + int(dt)) % len(br.stns)
	if br.stnIdx < 0 {
		br.stnIdx += len(br.stns)
	}
	return br.play()
}

func (br *BossRadio) turnVolume(delta int) error {
	if err := volume.IncreaseVolume(delta); err != nil {
		return err
	}

	// Flash new volume.
	vol, err := volume.GetVolume()
	if err != nil {
		return err
	}
	br.scrn.ClearText()
	br.scrn.SetTextLine(2, fmt.Sprintf("Volume: %d", vol))
	br.scrn.Draw()
	time.Sleep(250 * time.Millisecond)
	return nil
}

func (br *BossRadio) power() error {
	if br.state == stateOn {
		// Turning off.
		br.state = stateOff
		br.cmd.Process.Kill()
		br.cmd = nil
		br.showClock()
		return nil
	}

	// Turning on.
	br.state = stateOn
	return br.play()
}

func (br *BossRadio) isOff() bool {
	return br.state == stateOff
}

func (br *BossRadio) Run() error {
	pushChan, err := br.btn.Listen()
	if err != nil {
		return fmt.Errorf("Button.Listen failed: %w", err)
	}
	defer br.stop()

	// We start in off mode. Just show the clock.
	br.showClock()

	// Tick every 30 seconds to update the status screen or clock.
	statusUpdateTicker := time.NewTicker(30 * time.Second)
	defer statusUpdateTicker.Stop()

	for {
		var push button.PushDirection

		select {
		case <-statusUpdateTicker.C:
			if br.isOff() {
				log.Printf("Updating clock due to ticker")
				br.showClock()
				continue
			}
			log.Printf("Updating status due to ticker")
			br.showStatus()
			continue

		case push = <-pushChan:
			// Handled below:
		}

		// Always handle power button.
		if push == button.Center {
			if err := br.power(); err != nil {
				return err
			}
		}

		// Ignore all other buttons if we are off.
		if br.isOff() {
			continue
		}

		// Handle other buttons.
		switch push {
		case button.Left:
			if err := br.turnDial(dialTurnLeft); err != nil {
				return err
			}
		case button.Right:
			if err := br.turnDial(dialTurnRight); err != nil {
				return err
			}
		case button.Up:
			if err := br.turnVolume(5); err != nil {
				return err
			}
		case button.Down:
			if err := br.turnVolume(-5); err != nil {
				return err
			}
		case button.Center:
			// Already handled above.
		default:
			return fmt.Errorf("unknown button push: %v", push)
		}

		// Back to staus screen.
		br.showStatus()
	}
	panic("impossible")
}

func (br *BossRadio) showStatus() {
	stn := br.stns[br.stnIdx]
	status := stn.Status()
	br.scrn.SetText([6]string{
		"     " + stn.Name(),
		"",
		status.Show,
		status.Artist,
		status.Track,
		status.Album,
	})
	br.scrn.Draw()
}

func (br *BossRadio) showClock() {
	now := time.Now()
	br.scrn.ClearText()
	br.scrn.SetTextLine(3, "  "+now.Format("Jan 2 15:04"))
	br.scrn.Draw()
}

func (br *BossRadio) play() error {
	stn := br.stns[br.stnIdx]
	br.stop()
	br.cmd = exec.Command("mpv", "-no-video", stn.Stream())
	log.Printf("COMMAND: %+v", br.cmd)
	if err := br.cmd.Start(); err != nil {
		return err
	}

	// Flash new station logo for a second.
	br.scrn.DrawImage(stn.Logo())
	time.Sleep(250 * time.Millisecond)
	return nil
}

func (br *BossRadio) stop() {
	if br.cmd == nil {
		return
	}
	br.cmd.Process.Kill()
	br.cmd.Process.Release()
}

func (br *BossRadio) Destroy() {
	br.stop()
}
