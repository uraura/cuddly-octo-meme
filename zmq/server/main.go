package main

import (
	zmq "github.com/pebbe/zmq4"
	"time"

	"fmt"
)

func main() {
	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind("tcp://*:5556")

	for t := range time.Tick(time.Second) {
		topic := "foo"
		if t.Unix()%3 == 0 {
			topic = "bar"
		}

		publisher.SendMessage([]string{topic, fmt.Sprintf("%v", t.Unix())})
	}
}
