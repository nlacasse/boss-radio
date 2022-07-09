package bradio

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/itchyny/volume-go"

	"github.com/nlacasse/boss-radio/pkg/button"
	"github.com/nlacasse/boss-radio/pkg/events"
	"github.com/nlacasse/boss-radio/pkg/remote"
	"github.com/nlacasse/boss-radio/pkg/screen"
	"github.com/nlacasse/boss-radio/pkg/station"
	"github.com/nlacasse/boss-radio/pkg/web"
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
	btn    *button.Button
	remote *remote.Remote
	web    *web.Web
	scrn   *screen.Screen

	state     state
	cmd       *exec.Cmd
	stns      []station.Station
	stnIdx    int
	curStatus station.Status
}

func NewBossRadio(stns []station.Station) (*BossRadio, error) {
	scrn, err := screen.New()
	if err != nil {
		return nil, fmt.Errorf("screen.New failed: %v", err)
	}

	return &BossRadio{
		btn:    button.New(),
		remote: remote.New(),
		web:    web.New(),
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
	br.scrn.Freeze(250 * time.Millisecond)
	return nil
}

func (br *BossRadio) power() error {
	if br.state == stateOn {
		// Turning off.
		br.state = stateOff
		br.cmd.Process.Kill()
		br.cmd = nil
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
	eventCh := make(chan events.Event)
	webEventCh := make(chan events.Event)
	webStatusCh := make(chan web.Status)
	if err := br.btn.Listen(eventCh); err != nil {
		return fmt.Errorf("Button.Listen failed: %w", err)
	}
	if err := br.remote.Listen(eventCh); err != nil {
		return fmt.Errorf("Remote.Listen failed: %w", err)
	}
	if err := br.web.ListenAndUpdate(webEventCh, webStatusCh); err != nil {
		return fmt.Errorf("Web.Listen failed: %w", err)
	}
	defer br.stop()

	br.updateDisplay()

	// Tick every 30 seconds to update the status screen or clock.
	statusUpdateTicker := time.NewTicker(30 * time.Second)
	defer statusUpdateTicker.Stop()

	for {
		select {
		case ev := <-eventCh:
			log.Printf("got event %v", ev)
			if err := br.handleEvent(ev); err != nil {
				return err
			}

		case ev := <-webEventCh:
			log.Printf("got web event %v", ev)
			if err := br.handleEvent(ev); err != nil {
				return err
			}
			if br.state == stateOff {
				webStatusCh <- web.Status{}
			} else {
				stn := br.stns[br.stnIdx]
				webStatusCh <- web.Status{
					Power:  true,
					Name:   stn.Name(),
					Status: br.curStatus,
				}
			}

		case <-statusUpdateTicker.C:
			st := web.Status{}
			if br.state == stateOn {
				br.updateStatus()
				stn := br.stns[br.stnIdx]
				st = web.Status{
					Power:  true,
					Name:   stn.Name(),
					Status: br.curStatus,
				}
			}
			br.web.Update(st)
		}

		br.updateDisplay()
	}
}

func (br *BossRadio) handleEvent(ev events.Event) error {
	// Always handle power button (middle/play).
	if ev == events.ButtonCenter || ev == events.RemotePlay {
		return br.power()
	}

	// Ignore all other events if we are off.
	if br.isOff() {
		return nil
	}

	// Handle other events.
	switch ev {
	case events.ButtonLeft, events.RemoteLeft:
		return br.turnDial(dialTurnLeft)
	case events.ButtonRight, events.RemoteRight:
		return br.turnDial(dialTurnRight)
	case events.ButtonUp, events.RemoteUp:
		return br.turnVolume(5)
	case events.ButtonDown, events.RemoteDown:
		return br.turnVolume(-5)
	case events.RemoteMenu:
		// Noop for now.
		return nil
	case events.ButtonCenter, events.RemotePlay:
		// Already handled above.
		return nil
	default:
		return fmt.Errorf("unknown event: %v", ev)
	}
}

func (br *BossRadio) updateStatus() {
	log.Printf("update status")
	stn := br.stns[br.stnIdx]
	br.curStatus = stn.Status()
}

func (br *BossRadio) updateDisplay() {
	// Show main screen or clock.
	switch br.state {
	case stateOff:
		br.showClock()
	case stateOn:
		br.showStatus()
	default:
		panic(fmt.Sprintf("unknown state: %v", br.state))
	}

}

func (br *BossRadio) showStatus() {
	log.Printf("show status")
	stn := br.stns[br.stnIdx]
	namePad := (14 - len(stn.Name())) / 2
	if namePad < 0 {
		namePad = 0
	}
	br.scrn.SetText([6]string{
		strings.Repeat(" ", namePad) + stn.Name(),
		"",
		br.curStatus.Show,
		br.curStatus.Artist,
		br.curStatus.Track,
		br.curStatus.Album,
	})
	br.scrn.Draw()
}

func (br *BossRadio) showClock() {
	log.Printf("show clock")
	now := time.Now()
	br.scrn.ClearText()
	br.scrn.SetTextLine(1, "  "+now.Format("Jan 2 15:04"))
	if ip := getIP(); len(ip) > 0 {
		br.scrn.SetTextLine(4, "  "+ip.String())
	}
	br.scrn.Draw()
}

func (br *BossRadio) play() error {
	stn := br.stns[br.stnIdx]
	br.stop()
	br.cmd = stn.StreamCmd()
	if err := br.cmd.Start(); err != nil {
		return err
	}

	// Flash new station logo for a second.
	br.scrn.DrawImage(stn.Logo())
	br.scrn.Freeze(250 * time.Millisecond)
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
