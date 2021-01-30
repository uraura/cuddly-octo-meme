package main

import (
	"context"
	"fmt"
	"github.com/zeromq/goczmq"
	"time"
)

type payload [][]byte

func main() {
	ctx := context.Background()

	addr := "tcp://127.0.0.1:5555"
	pub := goczmq.NewPubChanneler(addr)
	defer pub.Destroy()
	sub1 := goczmq.NewSubChanneler(addr, "")
	defer sub1.Destroy()
	sub2 := goczmq.NewSubChanneler(addr, "foo")
	defer sub2.Destroy()

	publish := func(t time.Time) {
		// 1つめがtopic，いくつでも送れる
		data := payload{[]byte("foo"), []byte("test1"), []byte("test2"), []byte(fmt.Sprintf("%v", t.Unix()))}
		pub.SendChan <- data

		data2 := payload{[]byte("bar"), []byte(fmt.Sprintf("%v", t.Unix()))}
		pub.SendChan <- data2

		println("---")
	}

	go func() {
		publish(time.Now())
		for t := range time.Tick(time.Second) {
			publish(t)
		}
	}()

	go func() {
		for msg := range sub1.RecvChan {
			fmt.Printf("sub1 msg: %v\n", msg)
		}
	}()
	go func() {
		for msg := range sub2.RecvChan {
			fmt.Printf("sub2 msg: %v\n", msg)
		}
	}()

	time.Sleep(time.Second)
	sub3 := goczmq.NewSubChanneler(addr, "")
	defer sub3.Destroy()
	go func() {
		for msg := range sub3.RecvChan {
			fmt.Printf("sub3 msg: %v\n", msg)
		}
	}()

	<-ctx.Done()
}
