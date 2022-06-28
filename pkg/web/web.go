package web

import (
	"log"
	"net/http"
	"sync"
	"text/template"

	"github.com/nlacasse/boss-radio/pkg/events"
	"github.com/nlacasse/boss-radio/pkg/station"
)

var evMap = map[string]events.Event{
	"prev":     events.ButtonLeft,
	"next":     events.ButtonRight,
	"vol_up":   events.ButtonUp,
	"vol_down": events.ButtonDown,
	"power":    events.ButtonCenter,
}

type Status struct {
	Power  bool
	Name   string
	Status station.Status
}

type Web struct {
	status Status
	mu     sync.Mutex
}

func New() *Web {
	return &Web{}
}

func (w *Web) Update(status Status) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
}

func (w *Web) ListenAndUpdate(eventCh chan<- events.Event, statusCh <-chan Status) error {
	t, err := template.New("FreqM0d").Parse(tpl)
	if err != nil {
		return err
	}
	http.HandleFunc("/", func(res http.ResponseWriter, _ *http.Request) {
		log.Printf("serving /")
		if err := t.Execute(res, w.status); err != nil {
			log.Printf("Template failed: %v", err)
		}
	})
	for str, ev := range evMap {
		sstr := str
		sev := ev
		http.HandleFunc("/"+str, func(res http.ResponseWriter, req *http.Request) {
			w.mu.Lock()
			defer w.mu.Unlock()
			log.Printf("serving /%s", sstr)
			eventCh <- sev
			// Wait for return event.
			w.status = <-statusCh
			http.Redirect(res, req, "/", 307)
		})
	}

	go http.ListenAndServe(":8000", nil)

	return nil
}

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<title>FreqM0d</title>
		<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
		<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
		<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
		<link rel="manifest" href="/site.webmanifest">
		<style type="text/css">
			body {
				font-family: monospace;
				background-image: url('/skullfreq.jpg');
				background-repeat: no-repeat;
				background-attachment: fixed;
				background-size: contain;
				background-color: black;
			}
			h1 {
				font-size: 6em;
				color:red;
				text-shadow: -1px 0 black, 0 1px black, 1px 0 black, 0 -1px black;

			}
			h2 {
				font-size: 4em;
				color:red;
				text-shadow: -1px 0 black, 0 1px black, 1px 0 black, 0 -1px black;
			}
			a:link {
			  text-decoration: none;
			}

			a:visited {
			  text-decoration: none;
			}

			a:hover {
			  text-decoration: none;
			}

			a:active {
			  text-decoration: none;
			}
		</style>
		<meta http-equiv="refresh" content="10" />
	</head>
	<body>
		{{if .Power}}
			<h1>{{.Name}}</h1>
			<h2>{{.Status.Show}}</h2>
			<h2>{{.Status.Artist}}</h2>
			<h2>{{.Status.Track}}</h2>
			<h2>{{.Status.Album}}</h2>
			<br><br>
			<a href="/prev"><h1>PREV</h1></a>
			<a href="/next"><h1>NEXT</h1></a>
			<br>
			<a href="/vol_up"><h1>VOL UP</h1></a>
			<a href="/vol_down"><h1>VOL DOWN</h1></a>
			<br>
			<a href="/power"><h1>TURN OFF</h1></a><br>
		{{else}}
			<a href="/power"><h1>TURN ON</h1></a><br>
		{{end}}
	<body>
</html>
`
