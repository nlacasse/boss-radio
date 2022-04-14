package button

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"
)

type PushDirection int

const (
	Left PushDirection = iota
	Right
	Up
	Down
	Center
)

func (pd PushDirection) String() string {
	switch pd {
	case Left:
		return "Left"
	case Right:
		return "Right"
	case Up:
		return "Up"
	case Down:
		return "Down"
	case Center:
		return "Center"
	default:
		panic(fmt.Sprintf("unknown PushDirection: %v", pd))
	}
}

type Button struct {
}

func New() *Button {
	return &Button{}
}

func (b *Button) Listen() (<-chan PushDirection, error) {
	ch := make(chan PushDirection)
	btns := map[PushDirection]gpio.PinIO{
		Left:   gpioreg.ByName("GPIO14"),
		Right:  gpioreg.ByName("GPIO24"),
		Up:     gpioreg.ByName("GPIO23"),
		Down:   gpioreg.ByName("GPIO8"),
		Center: gpioreg.ByName("GPIO15"),
	}

	for pd, pin := range btns {
		if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
			return nil, fmt.Errorf("pin.In failed: %v", err)
		}
		dbPin, err := gpioutil.Debounce(pin, 3*time.Millisecond, 30*time.Millisecond, gpio.BothEdges)
		if err != nil {
			return nil, fmt.Errorf("gpioutil.Debounce failed: %v", err)
		}
		go func(pd PushDirection, dbPin gpio.PinIO) {
			for {
				edge := dbPin.WaitForEdge(1 * time.Second)
				if edge && dbPin.Read() == gpio.Low {
					ch <- pd
				}
			}
		}(pd, dbPin)
	}

	return ch, nil
}
