package bradio

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/itchyny/volume-go"
	"github.com/nlacasse/boss-radio/pkg/button"
	"github.com/nlacasse/boss-radio/pkg/remote"
	"github.com/nlacasse/boss-radio/pkg/screen"
	"github.com/nlacasse/boss-radio/pkg/station"
)

// Max time to wait before status updates.
const statusUpdateTime = 30 * time.Second

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
	btn    *button.Button
	remote *remote.Remote
	scrn   *screen.Screen

	state         state
	cmd           *exec.Cmd
	stns          []station.Station
	stnIdx        int
	curStatus     station.Status
	curStatusTime time.Time
}

func NewBossRadio(stns []station.Station) (*BossRadio, error) {
	scrn, err := screen.New()
	if err != nil {
		return nil, fmt.Errorf("screen.New failed: %v", err)
	}

	return &BossRadio{
		btn:    button.New(),
		remote: remote.New(),
		scrn:   scrn,
		state:  stateOff,
		stns:   stns,
	}, nil

}

func (br *BossRadio) turnDial(dt dialTurn) error {
	br.stnIdx = (br.stnIdx + int(dt)) % len(br.stns)
	if br.stnIdx < 0 {
		br.stnIdx += len(br.stns)
	}
	if err := br.play(); err != nil {
		return err
	}
	br.updateStatus()
	return nil
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
	if err := br.play(); err != nil {
		return err
	}
	br.updateStatus()
	return nil
}

func (br *BossRadio) isOff() bool {
	return br.state == stateOff
}

func (br *BossRadio) Run() error {
	pushChan, err := br.btn.Listen()
	if err != nil {
		return fmt.Errorf("Button.Listen failed: %w", err)
	}
	remoteChan, err := br.remote.Listen()
	if err != nil {
		return fmt.Errorf("Remote.Listen failed: %w", err)
	}
	defer br.stop()

	// We start in off mode. Just show the clock.
	br.showClock()

	// Tick every 10 seconds to update the status screen or clock.
	statusUpdateTicker := time.NewTicker(10 * time.Second)
	defer statusUpdateTicker.Stop()

	for {
		select {
		case push := <-pushChan:
			// Map the PushDirection to remoteButton. We key the behavior off
			// remote.Button type.
			pushMap := map[button.PushDirection]remote.Button{
				button.Left:   remote.Left,
				button.Right:  remote.Right,
				button.Up:     remote.Up,
				button.Down:   remote.Down,
				button.Center: remote.Play,
			}
			rb, ok := pushMap[push]
			if !ok {
				return fmt.Errorf("unknown push %v", push)
			}
			if err := br.handleButton(rb); err != nil {
				return err
			}

		case rb := <-remoteChan:
			if err := br.handleButton(rb); err != nil {
				return err
			}

		case <-statusUpdateTicker.C:
			// Refresh screen info below.
		}

		// Are we off, or did we just turn off?
		if br.isOff() {
			br.showClock()
			continue
		}

		// Update status if it's been long enough.
		if time.Now().After(br.curStatusTime.Add(statusUpdateTime)) {
			br.updateStatus()
		}

		br.showStatus()
	}
}

func (br *BossRadio) handleButton(rb remote.Button) error {
	// Always handle power button (middle/play).
	if rb == remote.Play {
		return br.power()
	}

	// Ignore all other buttons if we are off.
	if br.isOff() {
		return nil
	}

	// Handle other buttons.
	switch rb {
	case remote.Left:
		return br.turnDial(dialTurnLeft)
	case remote.Right:
		return br.turnDial(dialTurnRight)
	case remote.Up:
		return br.turnVolume(5)
	case remote.Down:
		return br.turnVolume(-5)
	case remote.Menu:
		// Noop for now.
		return nil
	case remote.Play:
		// Already handled above.
		return nil
	default:
		return fmt.Errorf("unknown button push: %v", rb)
	}
}

func (br *BossRadio) updateStatus() {
	log.Print("updateStatus()")
	stn := br.stns[br.stnIdx]
	br.curStatus = stn.Status()
	br.curStatusTime = time.Now()
}

func (br *BossRadio) showStatus() {
	stn := br.stns[br.stnIdx]
	br.scrn.SetText([6]string{
		"     " + stn.Name(),
		"",
		br.curStatus.Show,
		br.curStatus.Artist,
		br.curStatus.Track,
		br.curStatus.Album,
	})
	br.scrn.Draw()
}

func (br *BossRadio) showClock() {
	now := time.Now()
	br.scrn.ClearText()
	br.scrn.SetTextLine(1, "  "+now.Format("Jan 2 15:04"))
	br.scrn.SetTextLine(4, "  "+getIP().String())
	br.scrn.Draw()
}

func (br *BossRadio) play() error {
	stn := br.stns[br.stnIdx]
	br.stop()
	br.cmd = exec.Command("mpv", "-no-video", stn.Stream())
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
	br.scrn.Clear()
}

// Get preferred outbound ip of this machine
func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
