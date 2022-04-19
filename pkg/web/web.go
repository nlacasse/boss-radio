package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/nlacasse/boss-radio/pkg/events"
)

var evMap = map[string]events.Event{
	"left":   events.ButtonLeft,
	"right":  events.ButtonRight,
	"up":     events.ButtonUp,
	"down":   events.ButtonDown,
	"center": events.ButtonCenter,
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
		<a href="/center">CENTER</a><br>
		<a href="/left">LEFT</a><br>
		<a href="/right">RIGHT</a><br>
		<a href="/up">UP</a><br>
		<a href="/down">DOWN</a><br>
	<body>
</html>
`
