package osc

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestServerMessageReceiving(t *testing.T) {
	finish := make(chan bool)
	start := make(chan bool)
	done := sync.WaitGroup{}
	done.Add(2)

	// Start the server in a go-routine
	go func() {
		server := &Server{}
		c, err := net.ListenPacket("udp", "localhost:6677")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()

		// Start the client
		start <- true

		packet, err := server.ReceivePacket(c)
		if err != nil {
			t.Errorf("Server error: %v", err)
			return
		}
		if packet == nil {
			t.Error("nil packet")
			return
		}
		packet, err = server.ReceivePacket(c)
		if err != nil {
			t.Errorf("Server error: %v", err)
			return
		}
		if packet == nil {
			t.Error("nil packet")
			return
		}
		packet, err = server.ReceivePacket(c)
		if err != nil {
			t.Errorf("Server error: %v", err)
			return
		}
		if packet == nil {
			t.Error("nil packet")
			return
		}

		msg := packet.(*Message)
		if len(msg.Arguments) != 2 {
			t.Errorf("Argument length should be 2 and is: %d\n", len(msg.Arguments))
		}
		if msg.Arguments[0].(int32) != 1122 {
			t.Error("Argument should be 1122 and is: " + string(msg.Arguments[0].(int32)))
		}
		if msg.Arguments[1].(int32) != 3344 {
			t.Error("Argument should be 3344 and is: " + string(msg.Arguments[1].(int32)))
		}

		c.Close()
		finish <- true
	}()

	go func() {
		timeout := time.After(5 * time.Second)
		select {
		case <-timeout:
		case <-start:
			client := NewClient("localhost", 6677)
			msg := NewMessage("/address/test")
			msg.Append(int32(1122))
			msg.Append(int32(3344))
			time.Sleep(500 * time.Millisecond)
			client.Send(msg)
			client.Send(msg)
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

func TestReadTimeout(t *testing.T) {
	start := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		select {
		case <-time.After(5 * time.Second):
			t.Fatal("timed out")
		case <-start:
			client := NewClient("localhost", 6677)
			msg := NewMessage("/address/test1")
			err := client.Send(msg)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(150 * time.Millisecond)
			msg = NewMessage("/address/test2")
			err = client.Send(msg)
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		server := &Server{ReadTimeout: 100 * time.Millisecond}
		c, err := net.ListenPacket("udp", "localhost:6677")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()

		start <- true
		p, err := server.ReceivePacket(c)
		if err != nil {
			t.Errorf("server error: %v", err)
			return
		}
		if got, want := p.(*Message).Address, "/address/test1"; got != want {
			t.Errorf("wrong address; got = %s, want = %s", got, want)
			return
		}

		// Second receive should time out since client is delayed 150 milliseconds
		if _, err = server.ReceivePacket(c); err == nil {
			t.Errorf("expected error")
			return
		}

		// Next receive should get it
		p, err = server.ReceivePacket(c)
		if err != nil {
			t.Errorf("server error: %v", err)
			return
		}
		if got, want := p.(*Message).Address, "/address/test2"; got != want {
			t.Errorf("wrong address; got = %s, want = %s", got, want)
			return
		}
	}()

	wg.Wait()
}

func BenchmarkReceivePacket(b *testing.B) {
	d := &dummyConn{m: msg}
	s := &Server{}
	var p Packet
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p, _ = s.ReceivePacket(d)
	}
	result = p
}
