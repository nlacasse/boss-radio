package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/nlacasse/boss-radio/pkg/events"
)

var evMap = map[string]events.Event{
	"prev":     events.ButtonLeft,
	"next":     events.ButtonRight,
	"vol_up":   events.ButtonUp,
	"vol_down": events.ButtonDown,
	"power":    events.ButtonCenter,
}

type Web struct{}

func New() *Web {
	return &Web{}
}

func (w *Web) Listen(ch chan<- events.Event) error {
	t, err := template.New("FreqM0d").Parse(tpl)
	if err != nil {
		return err
	}
	http.HandleFunc("/", func(res http.ResponseWriter, _ *http.Request) {
		if err := t.Execute(res, nil); err != nil {
			log.Printf("Template failed: %v", err)
		}
	})
	for str, ev := range evMap {
		// Scope ev.
		sev := ev
		http.HandleFunc("/"+str, func(res http.ResponseWriter, _ *http.Request) {
			ch <- sev
			if err := t.Execute(res, nil); err != nil {
				log.Printf("Template failed: %v", err)
			}
		})
	}

	go http.ListenAndServe(":http", nil)
	return nil
}

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<title>FreqM0d</title>
	</head>
	<body>
		<h1>{{.Station}}</h1>
		<a href="/power"><h1>POWER</h1></a><br>
		<a href="/prev"><h1>PREV</h1></a>
		<a href="/next"><h1>NEXT</h1></a><br>
		<a href="/vol_up"><h1>VOL_UP</h1></a>
		<a href="/vol_down"><h1>VOL_DOWN</h1></a><br>
	<body>
</html>
`
