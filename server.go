package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"nanomsg.org/go-mangos"
	"nanomsg.org/go-mangos/protocol/rep"
	"nanomsg.org/go-mangos/transport/ws"
)

func die(format string, v ...interface{}) {
	log.Printf(format, v...)
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func reqHandler(sock mangos.Socket) {
	count := 0
	for {
		log.Printf("Receive from socket...")
		req, e := sock.Recv()
		if e != nil {
			die("Cannot get request: %v", e)
		}
		log.Printf("Received request: %s", string(req))
		reply := fmt.Sprintf("REPLY #%d %s", count, time.Now().String())
		if e := sock.Send([]byte(reply)); e != nil {
			die("Cannot send reply: %v", e)
		}
		count++
	}
}

func addReqHandler(r *mux.Router, port int) {
	log.Printf("adding handler")
	sock, _ := rep.NewSocket()

	sock.AddTransport(ws.NewTransport())

	url := fmt.Sprintf("ws://localhost:%d/req", port)
	if l, e := sock.NewListener(url, nil); e != nil {
		die("bad listener: %v", e)
	} else if h, e := l.GetOption(ws.OptionWebSocketHandler); e != nil {
		die("bad handler: %v", e)
	} else {
		l.SetOption(ws.OptionWebSocketCheckOrigin, false)
		r.Handle("/req", h.(http.Handler))
		l.Listen()
	}
	log.Printf("Run mangos rep")
	go reqHandler(sock)
}

func server(port int) {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	addReqHandler(r, port)
	e := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	die("Http server died: %v", e)
}

func main() {
	server(8080)
}
