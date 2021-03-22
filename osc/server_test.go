package osc

import (
	"net"
	"reflect"
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
			t.Error("Server error")
			return
		}
		if packet == nil {
			t.Error("nil packet")
			return
		}

		msg := packet.(*Message)
		if msg.CountArguments() != 2 {
			t.Errorf("Argument length should be 2 and is: %d\n", msg.CountArguments())
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

func TestParsePacket(t *testing.T) {
	for _, tt := range []struct {
		desc string
		msg  string
		pkt  Packet
		ok   bool
	}{
		{"no_args",
			"/a/b/c" + nulls(2) + "," + nulls(3),
			makePacket("/a/b/c", nil),
			true},
		{"string_arg",
			"/d/e/f" + nulls(2) + ",s" + nulls(2) + "foo" + nulls(1),
			makePacket("/d/e/f", []string{"foo"}),
			true},
		{"empty", "", nil, false},
	} {
		pkt, err := ParsePacket(tt.msg)
		if err != nil && tt.ok {
			t.Errorf("%s: ParsePacket() returned unexpected error; %s", tt.desc, err)
		}
		if err == nil && !tt.ok {
			t.Errorf("%s: ParsePacket() expected error", tt.desc)
		}
		if !tt.ok {
			continue
		}

		pktBytes, err := pkt.MarshalBinary()
		if err != nil {
			t.Errorf("%s: failure converting pkt to byte array; %s", tt.desc, err)
			continue
		}
		ttpktBytes, err := tt.pkt.MarshalBinary()
		if err != nil {
			t.Errorf("%s: failure converting tt.pkt to byte array; %s", tt.desc, err)
			continue
		}
		if got, want := pktBytes, ttpktBytes; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: ParsePacket() as bytes = '%s', want = '%s'", tt.desc, got, want)
			continue
		}
	}
}

var result Packet
var msg = []byte{47, 99, 111, 109, 112, 111, 115, 105, 116, 105, 111, 110, 47, 108, 97, 121, 101, 114, 115, 47, 49, 47, 99, 108, 105, 112, 115, 47, 49, 47, 116, 114, 97, 110, 115, 112, 111, 114, 116, 47, 112, 111, 115, 105, 116, 105, 111, 110, 0, 0, 0, 0, 44, 102, 0, 0, 62, 166, 157, 119, 0, 0, 0, 0}

//func BenchmarkParsePacket(b *testing.B) {
//	b.ResetTimer()
//	b.ReportAllocs()
//	var p Packet
//	for n := 0; n < b.N; n++ {
//		p, _ = ParsePacket(string(msg))
//	}
//	result = p
//}

type dummyConn struct {
	net.PacketConn
}

func (d *dummyConn) ReadFrom(buf []byte) (n int, addr net.Addr, err error) {
	n = copy(buf, msg)
	return
}

func (d *dummyConn) WriteTo(buf []byte, addr net.Addr) (n int, err error) { return }

func (d *dummyConn) Close() (err error) { return }

func (d *dummyConn) LocalAddr() (addr net.Addr) { return }

func (d *dummyConn) SetDeadline(t time.Time) (err error) { return }

func (d *dummyConn) SetReadDeadline(t time.Time) (err error) { return }

func (d *dummyConn) SetWriteDeadline(t time.Time) (err error) { return }

func BenchmarkReceivePacket(b *testing.B) {
	d := &dummyConn{}
	s := &Server{}
	var p Packet
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p, _ = s.ReceivePacket(d)
	}
	result = p
}
