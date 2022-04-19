package events

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
