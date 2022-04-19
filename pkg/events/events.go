package events

import "fmt"

type Event int

const (
	ButtonUp Event = iota
	ButtonDown
	ButtonLeft
	ButtonRight
	ButtonCenter
	RemoteUp
	RemoteDown
	RemoteLeft
	RemoteRight
	RemotePlay
	RemoteMenu
)

func (e Event) String() string {
	switch e {
	case ButtonUp:
		return "ButtonUp"
	case ButtonDown:
		return "ButtonDown"
	case ButtonLeft:
		return "ButtonLeft"
	case ButtonRight:
		return "ButtonRight"
	case ButtonCenter:
		return "ButtonCenter"
	case RemoteUp:
		return "RemoteUp"
	case RemoteDown:
		return "RemoteDown"
	case RemoteLeft:
		return "RemoteLeft"
	case RemoteRight:
		return "RemoteRight"
	case RemotePlay:
		return "RemotePlay"
	case RemoteMenu:
		return "RemoteMenu"
	default:
		return fmt.Sprintf("unknown event %v", e)
	}
}
