package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"testing"
)

const k = 1024
const m = 1024 * k

func BenchmarkCompare(b *testing.B) {
	b.Run("pubsub 1k", func(b *testing.B) {
		benchmarkProxyPubSub(k, b)
	})
	b.Run("pushpull 1k", func(b *testing.B) {
		benchmarkProxyPullPush(k, b)
	})
}

func BenchmarkPullPush(b *testing.B) {
	b.Run("1k", func(b *testing.B) {
		benchmarkProxyPullPush(k, b)
	})
	b.Run("2k", func(b *testing.B) {
		benchmarkProxyPullPush(2*k, b)
	})
	b.Run("4k", func(b *testing.B) {
		benchmarkProxyPullPush(4*k, b)
	})
	b.Run("8k", func(b *testing.B) {
		benchmarkProxyPullPush(8*k, b)
	})
	b.Run("16k", func(b *testing.B) {
		benchmarkProxyPullPush(16*k, b)
	})
	b.Run("32k", func(b *testing.B) {
		benchmarkProxyPullPush(32*k, b)
	})
	b.Run("64k", func(b *testing.B) {
		benchmarkProxyPullPush(64*k, b)
	})
	b.Run("128k", func(b *testing.B) {
		benchmarkProxyPullPush(128*k, b)
	})
	b.Run("256k", func(b *testing.B) {
		benchmarkProxyPullPush(256*k, b)
	})
	b.Run("512k", func(b *testing.B) {
		benchmarkProxyPullPush(512*k, b)
	})
	b.Run("1m", func(b *testing.B) {
		benchmarkProxyPullPush(1*m, b)
	})
}

func benchmarkProxyPullPush(size int, b *testing.B) {
	go func() {
		frontend, _ := zmq.NewSocket(zmq.PULL)
		defer frontend.Close()
		frontend.Bind(fmt.Sprintf("inproc://front%d", size))

		backend, _ := zmq.NewSocket(zmq.PUSH)
		defer backend.Close()
		backend.Bind(fmt.Sprintf("inproc://back%d", size))

		zmq.Proxy(frontend, backend, nil)
	}()

	go func() {
		pushSock, _ := zmq.NewSocket(zmq.PUSH)
		defer pushSock.Close()
		pushSock.Connect(fmt.Sprintf("inproc://front%d", size))

		payload := make([]byte, size)
		for i := 0; i < b.N; i++ {
			_, err := pushSock.SendBytes(payload, 0)
			if err != nil {
				log.Printf("send error: %v", err)
				panic(err)
			}
		}
	}()

	pullSock, _ := zmq.NewSocket(zmq.PULL)
	defer pullSock.Close()
	pullSock.Connect(fmt.Sprintf("inproc://back%d", size))
	for i := 0; i < b.N; i++ {
		msg, err := pullSock.RecvBytes(0)
		if err != nil {
			log.Printf("recv error: %v", err)
			panic(err)
		}
		if len(msg) != size {
			log.Printf("size error: %v", err)
			panic("msg too small")
		}
		b.SetBytes(int64(size))
	}
}

func BenchmarkPubSub(b *testing.B) {
	b.Run("1k", func(b *testing.B) {
		benchmarkProxyPubSub(1*k, b)
	})
	b.Run("2k", func(b *testing.B) {
		benchmarkProxyPubSub(2*k, b)
	})
	b.Run("4k", func(b *testing.B) {
		benchmarkProxyPubSub(4*k, b)
	})
	b.Run("8k", func(b *testing.B) {
		benchmarkProxyPubSub(8*k, b)
	})
	b.Run("16k", func(b *testing.B) {
		benchmarkProxyPubSub(16*k, b)
	})
	b.Run("32k", func(b *testing.B) {
		benchmarkProxyPubSub(32*k, b)
	})
	b.Run("64k", func(b *testing.B) {
		benchmarkProxyPubSub(64*k, b)
	})
	b.Run("128k", func(b *testing.B) {
		benchmarkProxyPubSub(128*k, b)
	})
	b.Run("256k", func(b *testing.B) {
		benchmarkProxyPubSub(256*k, b)
	})
	b.Run("512k", func(b *testing.B) {
		benchmarkProxyPubSub(512*k, b)
	})
	b.Run("1m", func(b *testing.B) {
		benchmarkProxyPubSub(1*m, b)
	})
}

func benchmarkProxyPubSub(size int, b *testing.B) {
	go func() {
		frontend, _ := zmq.NewSocket(zmq.XSUB)
		defer frontend.Close()
		frontend.Bind(fmt.Sprintf("inproc://front2%d", size))
		//frontend.Connect("tcp://127.0.0.1:5555")

		backend, _ := zmq.NewSocket(zmq.XPUB)
		defer backend.Close()
		backend.Bind(fmt.Sprintf("inproc://back2%d", size))
		//backend.Bind("tcp://127.0.0.1:5556")

		zmq.Proxy(frontend, backend, nil)
	}()

	for i := 0; i < 2; i++ {
		go func() {
			pub, _ := zmq.NewSocket(zmq.PUB)
			defer pub.Close()
			pub.Connect(fmt.Sprintf("inproc://front2%d", size))
			//pub.Bind("tcp://*:5555")

			payload := make([]byte, size)
			for {
				_, err := pub.SendBytes(payload, 0)
				if err != nil {
					log.Printf("send error: %v", err)
					panic(err)
				}
			}
		}()
	}

	sub, _ := zmq.NewSocket(zmq.SUB)
	defer sub.Close()
	sub.Connect(fmt.Sprintf("inproc://back2%d", size))
	sub.SetSubscribe("") // all

	for i := 0; i < b.N; i++ {
		msg, err := sub.RecvBytes(0)
		if err != nil {
			log.Printf("recv error: %v", err)
			panic(err)
		}
		if len(msg) != size {
			log.Printf("size error: %v", err)
			panic("msg too small")
		}
		b.SetBytes(int64(size))
	}
}
