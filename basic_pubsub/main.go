package main

import (
	"context"
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"time"
)

const addr = "tcp://127.0.0.1:5555"

type payload [][]byte

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 10; i++ {
		go func(ctx context.Context, idx int) {
			sub, err := goczmq.NewSub("tcp://127.0.0.1:5555,tcp://127.0.0.1:5556", "")
			if err != nil {
				log.Fatalf("[%d] new sub err: %v\n", idx, err)
			}
			success := 0
			failure := 0

			defer func() {
				sub.Destroy()
				log.Printf("[%d] sub stop. success=%d failure=%d\n", idx, success, failure)
			}()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := sub.RecvMessage()
					if err != nil {
						failure++
						break
					}
					success++
					//log.Printf("[%d] msg: %v", idx, msg)
				}
			}
		}(ctx, i)
	}

	for i := 0; i < 2; i++ {
		go func(ctx context.Context, idx int) {
			pub, err := goczmq.NewPub(fmt.Sprintf("tcp://*:%d", 5555+idx))
			if err != nil {
				log.Fatalf("[%d] new pub err: %v", idx, err)
			}
			success := 0
			failure := 0

			defer func() {
				pub.Destroy()
				log.Printf("[%d] pub stop. success=%d failure=%d\n", idx, success, failure)
			}()
			ticker := time.NewTicker(1 * time.Millisecond)

			for {
				select {
				case <-ctx.Done():
					return
				case ts := <-ticker.C:
					data := [][]byte{[]byte("foo"), []byte("bar"), []byte(fmt.Sprintf("%v", ts.UnixNano()))}
					if err := pub.SendMessage(data); err != nil {
						//log.Printf("[%d] pub err: %v", idx, err)
						failure++
						break
					}
					success++
				}
			}
		}(ctx, i)
	}

	time.Sleep(1 * time.Second)
}

func sample2() {
	ctx := context.Background()

	pub, _ := goczmq.NewPub(addr)
	defer pub.Destroy()
	sub1, _ := goczmq.NewSub(addr, "")
	defer sub1.Destroy()
	sub2, _ := goczmq.NewSub(addr, "foo")
	defer sub2.Destroy()

	publish := func(t time.Time) {
		// 1つめがtopic，いくつでも送れる
		data := payload{[]byte("foo"), []byte("test1"), []byte("test2"), []byte(fmt.Sprintf("%v", t.Unix()))}
		if err := pub.SendMessage(data); err != nil {
			log.Fatal(err)
		}

		data2 := payload{[]byte("bar"), []byte(fmt.Sprintf("%v", t.Unix()))}
		if err := pub.SendMessage(data2); err != nil {
			log.Fatal(err)
		}

		println("---")
	}

	go func() {
		publish(time.Now())
		for t := range time.Tick(time.Second) {
			publish(t)
		}
	}()

	go func() {
		for {
			msg, err := sub1.RecvMessage()
			if err != nil {
				log.Fatalf("sub1: %v", err)
			}
			fmt.Printf("sub1 msg: %v\n", msg)
		}
	}()
	go func() {
		for {
			msg, err := sub2.RecvMessage()
			if err != nil {
				log.Fatalf("sub2: %v", err)
			}
			fmt.Printf("sub2 msg: %v\n", msg)
		}
	}()

	time.Sleep(time.Second)
	sub3, _ := goczmq.NewSub(addr, "")
	defer sub3.Destroy()
	go func() {
		for {
			msg, err := sub3.RecvMessage()
			if err != nil {
				log.Fatalf("sub3: %v", err)
			}
			fmt.Printf("sub3 msg: %v\n", msg)
		}
	}()

	<-ctx.Done()
}

func sample() {
	// Create a router socket and bind it to port 5555.
	router, err := goczmq.NewRouter("tcp://*:5555")
	if err != nil {
		log.Fatal(err)
	}
	defer router.Destroy()

	log.Println("router created and bound")

	// Create a dealer socket and connect it to the router.
	dealer, err := goczmq.NewDealer("tcp://127.0.0.1:5555")
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()

	log.Println("dealer created and connected")

	// Send a 'Hello' message from the dealer to the router.
	// Here we send it as a frame ([]byte), with a FlagNone
	// flag to indicate there are no more frames following.
	err = dealer.SendFrame([]byte("Hello"), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("dealer sent 'Hello'")

	// Receve the message. Here we call RecvMessage, which
	// will return the message as a slice of frames ([][]byte).
	// Since this is a router socket that support async
	// request / reply, the first frame of the message will
	// be the routing frame.
	request, err := router.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("router received '%s' from '%v'", request[1], request[0])

	// Send a reply. First we send the routing frame, which
	// lets the dealer know which client to send the message.
	// The FlagMore flag tells the router there will be more
	// frames in this message.
	err = router.SendFrame(request[0], goczmq.FlagMore)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("router sent 'World'")

	// Next send the reply. The FlagNone flag tells the router
	// that this is the last frame of the message.
	err = router.SendFrame([]byte("World"), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	// Receive the reply.
	reply, err := dealer.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dealer received '%s'", string(reply[0]))
}
