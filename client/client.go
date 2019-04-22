// MIT Licensed

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gopherjs/gopherwasm/js"
	wasm "github.com/strowk/mangos-in-browser/client/wasm"
	"nanomsg.org/go/mangos/v2/protocol/req"
)

func die(format string, v ...interface{}) {
	log.Printf(format, v...)
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func reqClient(port int) {
	sock, e := req.NewSocket()
	if e != nil {
		die("cannot make req socket: %v", e)
	}
	defer sock.Close()

	// new mangos does not want this
	// sock.AddTransport(transport)

	url := fmt.Sprintf("ws://localhost:%d/req", port)

	log.Printf("Start dial to %s", url)
	if e = sock.Dial(url); e != nil {
		die("cannot dial req url: %v", e)
	}
	log.Printf("Dialed")

	log.Printf("Start sending Hello")
	if e = sock.Send([]byte("Hello")); e != nil {
		die("Cannot send req: %v", e)
	}

	log.Printf("Start waiting for reply")
	if m, e := sock.Recv(); e != nil {
		die("Cannot recv reply: %v", e)
	} else {
		log.Printf("Received reply")
		msg := string(m)
		fmt.Printf("%s\n", msg)
		js.Global().Get("document").Call("getElementById", "result").Set("innerHTML", msg)
	}
}

func main() {
	wasm.Init()
	port := 8080
	callback := js.NewCallback(func(args []js.Value) {
		go reqClient(port)
	})
	defer callback.Release()
	setReq := js.Global().Get("setReq")
	setReq.Invoke(callback)
	select {}
}
