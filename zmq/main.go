package main

import (
	zmq "github.com/pebbe/zmq4"
	"time"
)

func main() {
	go func() {
		frontend, _ := zmq.NewSocket(zmq.XSUB)
		defer frontend.Close()
		frontend.Bind("inproc://front")
		//frontend.Connect("tcp://127.0.0.1:5555")

		backend, _ := zmq.NewSocket(zmq.XPUB)
		defer backend.Close()
		backend.Bind("inproc://back")
		//backend.Bind("tcp://127.0.0.1:5556")

		zmq.Proxy(frontend, backend, nil)
	}()

	for i := 0; i < 3; i++ {
		go func(idx int) {
			pub, _ := zmq.NewSocket(zmq.PUB)
			defer pub.Close()
			pub.Connect("inproc://front")
			//pub.Bind("tcp://*:5555")

			payload := make([]byte, 1024)
			for range time.Tick(time.Second) {
				_, err := pub.SendBytes(payload, 0)
				if err != nil {
					panic(err)
				}
				println("send 1024 bytes")
			}
		}(i)
	}

	sub, _ := zmq.NewSocket(zmq.SUB)
	defer sub.Close()
	sub.Connect("inproc://back")
	//sub.Connect("tcp://127.0.0.1:5556")
	sub.SetSubscribe("") // all

	for {
		msg, err := sub.RecvBytes(0)
		if err != nil {
			break
		}
		if len(msg) != 1024 {
			panic("msg too small")
		}
		println("recv 1024 bytes")
	}
}
