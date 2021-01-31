package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"

	"log"
)

func main() {
	frontend, _ := zmq.NewSocket(zmq.XSUB)
	defer frontend.Close()
	frontend.Connect("tcp://127.0.0.1:5556")

	backend, _ := zmq.NewSocket(zmq.XPUB)
	defer backend.Close()
	backend.Bind("tcp://127.0.0.1:5555")

	capture, _ := zmq.NewSocket(zmq.PUSH)
	defer capture.Close()
	capture.Bind("inproc://tap")

	go func() {
		p, _ := zmq.NewSocket(zmq.PULL)
		defer p.Close()
		p.Connect("inproc://tap")

		for {
			msg, _ := p.RecvMessage(0)
			fmt.Printf("tap: %v\n", msg)
		}
	}()

	err := zmq.Proxy(frontend, backend, capture)
	log.Fatalln("Proxy interrupted:", err)
}
