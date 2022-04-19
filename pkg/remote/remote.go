package remote

import (
	"log"

	"periph.io/x/conn/v3/ir"
	"periph.io/x/devices/v3/lirc"

	"github.com/nlacasse/boss-radio/pkg/events"
)

// Codes for Apple A1156.
var btnMap = map[ir.Key]events.Event{
	ir.KEY_KPPLUS:      events.RemoteUp,
	ir.KEY_KPMINUS:     events.RemoteDown,
	ir.KEY_REWIND:      events.RemoteLeft,
	ir.KEY_FASTFORWARD: events.RemoteRight,
	ir.KEY_PLAY:        events.RemotePlay,
	ir.KEY_MENU:        events.RemoteMenu,
}

type Remote struct{}

func New() *Remote {
	return &Remote{}
}

func (r *Remote) Listen(ch chan<- events.Event) error {
	// Open a handle to lircd:
	conn, err := lirc.New()
	if err != nil {
		return err
	}

	go func() {
		for msg := range conn.Channel() {
			if msg.Repeat {
				continue
			}
			ev, ok := btnMap[msg.Key]
			if !ok {
				log.Printf("unknown key: %v", msg.Key)
				continue
			}
			ch <- ev
		}
	}()
	return nil
}
