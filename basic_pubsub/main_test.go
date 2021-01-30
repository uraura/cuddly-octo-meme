package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"testing"
	"time"
)

func Test_pubsub(t *testing.T) {

	t.Run("10pub/100sub", func(t *testing.T) {

		for i := 0; i < 10; i++ {
			go func(idx int) {
				sub, _ := goczmq.NewSub(addr, "")
				defer sub.Destroy()
				for {
					if _, err := sub.RecvMessage(); err != nil {
						t.Fatalf("[%d] sub err: %v", idx, err)
					}
				}
			}(i)
		}

		for i := 0; i < 10; i++ {
			go func(idx int) {
				pub, _ := goczmq.NewPub(addr)
				defer pub.Destroy()
				// wait
				time.Sleep(time.Second)

				for ts := range time.Tick(1 * time.Millisecond) {
					data := [][]byte{[]byte("foo"), []byte("bar"), []byte(fmt.Sprintf("%v", ts.Unix()))}
					if err := pub.SendMessage(data); err != nil {
						t.Fatalf("[%d] pub err: %v", idx, err)
					}
				}
			}(i)
		}

		time.Sleep(time.Second)
	})

}
