package main

import (
	zmq "github.com/pebbe/zmq4"
	"os"

	"fmt"
)

func main() {
	subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
	subscriber.Connect("tcp://localhost:5555")

	filter := ""
	if len(os.Args) > 1 {
		filter = os.Args[1]
	}

	subscriber.SetSubscribe(filter)

	for {
		msg, _ := subscriber.RecvMessage(0)

		fmt.Printf("received: %v\n", msg)
	}
}
