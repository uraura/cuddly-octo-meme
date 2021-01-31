package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	// Create and configure our proxy
	proxy := goczmq.NewProxy()
	defer proxy.Destroy()

	var err error

	if testing.Verbose() {
		err = proxy.Verbose()
		if err != nil {
			t.Error(err)
		}
	}

	err = proxy.SetFrontend(goczmq.Pull, "inproc://frontend")
	if err != nil {
		t.Error(err)
	}

	err = proxy.SetBackend(goczmq.Push, "inproc://backend")
	if err != nil {
		t.Error(err)
	}

	err = proxy.SetCapture("inproc://capture")
	if err != nil {
		t.Error(err)
	}

	// connect application sockets to proxy
	faucet := goczmq.NewSock(goczmq.Push)
	err = faucet.Connect("inproc://frontend")
	if err != nil {
		t.Error(err)
	}
	defer faucet.Destroy()

	sink := goczmq.NewSock(goczmq.Pull)
	err = sink.Connect("inproc://backend")
	if err != nil {
		t.Error(err)
	}
	defer sink.Destroy()

	tap := goczmq.NewSock(goczmq.Pull)
	_, err = tap.Bind("inproc://capture")
	if err != nil {
		t.Error(err)
	}
	defer tap.Destroy()

	// send some messages and check they arrived
	err = faucet.SendFrame([]byte("Hello"), goczmq.FlagNone)
	if err != nil {
		t.Error(err)
	}

	err = faucet.SendFrame([]byte("World"), goczmq.FlagNone)
	if err != nil {
		t.Error(err)
	}

	// check the tap
	b, f, err := tap.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, have := false, f == goczmq.FlagMore; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "Hello", string(b); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	b, f, err = tap.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, have := false, f == goczmq.FlagMore; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "World", string(b); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	b, f, err = sink.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, have := false, f == goczmq.FlagMore; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "Hello", string(b); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	_, f, err = sink.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if f == goczmq.FlagMore {
		t.Error("FlagMore set and should not be")
	}

	if want, have := false, f == goczmq.FlagMore; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	err = proxy.Pause()
	if err != nil {
		t.Error(err)
	}

	err = faucet.SendFrame([]byte("Belated Hello"), goczmq.FlagNone)
	if err != nil {
		t.Error(err)
	}

	if want, have := false, sink.Pollin(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := false, tap.Pollin(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	err = proxy.Resume()
	if err != nil {
		t.Error(err)
	}

	b, f, err = sink.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, have := false, f == goczmq.FlagMore; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "Belated Hello", string(b); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	b, f, err = tap.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, have := false, f == goczmq.FlagMore; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "Belated Hello", string(b); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	proxy.Destroy()
}

func BenchmarkProxySendFrame1k(b *testing.B)  { benchmarkProxySendFrame(1024, b) }
func BenchmarkProxySendFrame4k(b *testing.B)  { benchmarkProxySendFrame(4096, b) }
func BenchmarkProxySendFrame8k(b *testing.B)  { benchmarkProxySendFrame(8192, b) }
func BenchmarkProxySendFrame16k(b *testing.B) { benchmarkProxySendFrame(16384, b) }
func BenchmarkProxySendFrame32k(b *testing.B) { benchmarkProxySendFrame(32768, b) }

func benchmarkProxySendFrame(size int, b *testing.B) {
	proxy := goczmq.NewProxy()
	defer proxy.Destroy()

	err := proxy.SetFrontend(goczmq.Pull, fmt.Sprintf("inproc://benchProxyFront%d", size))
	if err != nil {
		//panic(err)
		fmt.Printf("err: %v\n", err)
	}

	err = proxy.SetBackend(goczmq.Push, fmt.Sprintf("inproc://benchProxyBack%d", size))
	if err != nil {
		//panic(err)
		fmt.Printf("err: %v\n", err)
	}

	pullSock := goczmq.NewSock(goczmq.Pull)
	defer pullSock.Destroy()

	err = pullSock.Connect(fmt.Sprintf("inproc://benchProxyBack%d", size))
	if err != nil {
		//panic(err)
		fmt.Printf("err: %v\n", err)
	}

	go func() {
		pushSock := goczmq.NewSock(goczmq.Push)
		defer pushSock.Destroy()
		err := pushSock.Connect(fmt.Sprintf("inproc://benchProxyFront%d", size))
		if err != nil {
			//panic(err)
			fmt.Printf("err: %v\n", err)
		}

		payload := make([]byte, size)
		for i := 0; i < b.N; i++ {
			err = pushSock.SendFrame(payload, goczmq.FlagNone)
			if err != nil {
				//panic(err)
				fmt.Printf("err: %v\n", err)
			}
		}
	}()

	for i := 0; i < b.N; i++ {
		msg, _, err := pullSock.RecvFrame()
		if err != nil {
			//panic(err)
			fmt.Printf("err: %v\n", err)
		}
		if len(msg) != size {
			//panic("msg too small")
			fmt.Printf("err: %v\n", "msg too small")
		}
		b.SetBytes(int64(size))
	}
}

func Test_pubsub(t *testing.T) {
	t.Skip()

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
