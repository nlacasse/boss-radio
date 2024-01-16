module github.com/nlacasse/boss-radio

go 1.18

replace periph.io/x/devices/v3 => /home/nlacasse/go/src/periph.io/x/devices

require (
	github.com/itchyny/volume-go v0.2.1
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	golang.org/x/text v0.3.6
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/devices/v3 v3.6.13
	periph.io/x/host/v3 v3.7.2
)

require (
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/moutend/go-wca v0.2.0 // indirect
	github.com/warthog618/gpiod v0.8.2 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
