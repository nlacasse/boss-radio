package button

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"

	"github.com/nlacasse/boss-radio/pkg/events"
)

type Button struct{}

func New() *Button {
	return &Button{}
}

func (b *Button) Listen(ch chan<- events.Event) error {
	btns := map[events.Event]gpio.PinIO{
		events.ButtonLeft:   gpioreg.ByName("GPIO14"),
		events.ButtonRight:  gpioreg.ByName("GPIO24"),
		events.ButtonUp:     gpioreg.ByName("GPIO23"),
		events.ButtonDown:   gpioreg.ByName("GPIO8"),
		events.ButtonCenter: gpioreg.ByName("GPIO15"),
	}

	for ev, pin := range btns {
		if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
			return fmt.Errorf("pin.In failed: %v", err)
		}
		dbPin, err := gpioutil.Debounce(pin, 3*time.Millisecond, 30*time.Millisecond, gpio.BothEdges)
		if err != nil {
			return fmt.Errorf("gpioutil.Debounce failed: %v", err)
		}
		go func(ev events.Event, dbPin gpio.PinIO) {
			for {
				edge := dbPin.WaitForEdge(1 * time.Second)
				if edge && dbPin.Read() == gpio.Low {
					ch <- ev
				}
			}
		}(ev, dbPin)
	}

	return nil
}
