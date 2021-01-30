package main

import (
	"github.com/zeromq/goczmq"
	"testing"
	"time"
)

func Benchmark_bench(b *testing.B) {

	pub := goczmq.NewPubChanneler(addr)
	defer pub.Destroy()
	sub := goczmq.NewSubChanneler(addr, "")
	defer sub.Destroy()

	// wait
	time.Sleep(time.Second)

	data := [][]byte{[]byte("foo"), []byte("bar")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pub.SendChan <- data
	}
}
