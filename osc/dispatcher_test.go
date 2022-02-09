package osc

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestAddMethodFunc(t *testing.T) {
	d := &Dispatcher{}
	err := d.AddMethodFunc("/address/test", func(msg *Message) {})
	if err != nil {
		t.Error("Expected that OSC address '/address/test' is valid")
	}
}

func TestAddMethodFuncFail(t *testing.T) {
	d := &Dispatcher{}
	err := d.AddMethodFunc("/address*/test", func(msg *Message) {})
	if err == nil {
		t.Error("Expected error with '/address*/test'")
	}
}

func TestServerMessageDispatching(t *testing.T) {
	finish := make(chan bool)
	start := make(chan bool)
	done := sync.WaitGroup{}
	done.Add(2)

	// Start the OSC server in a new go-routine
	go func() {
		conn, err := net.ListenPacket("udp", "localhost:6677")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		d := &Dispatcher{}
		err = d.AddMethodFunc("/address/test", func(msg *Message) {
			if len(msg.Arguments) != 1 {
				t.Error("Argument length should be 1 and is: " + string(rune(len(msg.Arguments))))
			}

			if msg.Arguments[0].(int32) != 1122 {
				t.Error("Argument should be 1122 and is: " + string(msg.Arguments[0].(int32)))
			}

			// Stop OSC server
			conn.Close()
			finish <- true
		})
		if err != nil {
			t.Error("Error adding message handler")
		}

		server := &Server{Addr: "localhost:6677", Handler: d.Dispatch}
		start <- true
		server.Serve()
	}()

	go func() {
		timeout := time.After(5 * time.Second)
		select {
		case <-timeout:
		case <-start:
			time.Sleep(500 * time.Millisecond)
			client := NewClient("localhost", 6677)
			msg := NewMessage("/address/test")
			msg.Append(int32(1122))
			client.Send(msg)
		}

		done.Done()

		select {
		case <-timeout:
		case <-finish:
		}
		done.Done()
	}()

	done.Wait()
}
