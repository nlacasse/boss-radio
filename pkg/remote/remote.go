package remote

import (
	"fmt"
	"log"

	"periph.io/x/conn/v3/ir"
	"periph.io/x/devices/v3/lirc"
)

type Button int

const (
	Up Button = iota
	Down
	Left
	Right
	Play
	Menu
)

func (b Button) String() string {
	switch b {
	case Up:
		return "Up"
	case Down:
		return "Down"
	case Left:
		return "Left"
	case Right:
		return "Right"
	case Play:
		return "Play"
	case Menu:
		return "Menu"
	default:
		panic(fmt.Sprintf("unknown button %v", b))
	}
}

var btnMap = map[ir.Key]Button{
	ir.KEY_UP:    Up,
	ir.KEY_DOWN:  Down,
	ir.KEY_LEFT:  Left,
	ir.KEY_RIGHT: Right,
	ir.KEY_PLAY:  Play,
	ir.KEY_MENU:  Menu,
}

type Remote struct {
}

func New() *Remote {
	return &Remote{}
}

func (r *Remote) Listen() (<-chan Button, error) {
	// Open a handle to lircd:
	conn, err := lirc.New()
	if err != nil {
		return nil, err
	}

	ch := make(chan Button)
	go func() {
		for msg := range conn.Channel() {
			if msg.Repeat {
				continue
			}
			btn, ok := btnMap[msg.Key]
			if !ok {
				log.Printf("unknown key: %v", msg.Key)
				continue
			}
			ch <- btn
		}
	}()
	return ch, nil
}
