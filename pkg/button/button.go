package button

import (
	"fmt"
	"time"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"

	"github.com/nlacasse/boss-radio/pkg/events"
)

const gpiochip = "gpiochip0"

type Button struct{}

func New() *Button {
	return &Button{}
}

func (b *Button) Listen(ch chan<- events.Event) error {
	chip, err := gpiod.NewChip(gpiochip)
	if err != nil {
		return fmt.Errorf("NewChip(%q) failed: %v", gpiochip, err)
	}

	btns := map[events.Event]string{
		events.ButtonLeft:   "gpio14",
		events.ButtonRight:  "gpio24",
		events.ButtonUp:     "gpio23",
		events.ButtonDown:   "gpio8",
		events.ButtonCenter: "gpio15",
	}

	for ev, pinName := range btns {
		pin, err := rpi.Pin(pinName)
		if err != nil {
			return fmt.Errorf("error getting pin %q: %v", pinName, err)
		}
		evv := ev
		handler := func(_ gpiod.LineEvent) {
			ch <- evv
		}
		if _, err := chip.RequestLine(pin,
			gpiod.AsInput,
			gpiod.WithPullUp,
			gpiod.WithFallingEdge,
			gpiod.WithDebounce(30*time.Millisecond),
			gpiod.WithEventHandler(handler)); err != nil {
			return fmt.Errorf("error getting line for pin %q(%d): %v", pinName, pin, err)
		}
	}

	return nil
}
