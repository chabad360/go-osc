package osc

import (
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

type dummyConn struct {
	net.Conn
	m []byte
}

func (d *dummyConn) ReadFrom(buf []byte) (n int, addr net.Addr, err error) {
	n = copy(buf, d.m)
	return
}

func (d *dummyConn) Read(buf []byte) (n int, err error) {
	n = copy(buf, d.m)
	return
}

func (d *dummyConn) WriteTo(_ []byte, _ net.Addr) (n int, err error) { return }

func (d *dummyConn) Close() (err error) { return }

func (d *dummyConn) LocalAddr() (addr net.Addr) { return }

func (d *dummyConn) SetDeadline(_ time.Time) (err error) { return }

func (d *dummyConn) SetReadDeadline(_ time.Time) (err error) { return }

func (d *dummyConn) SetWriteDeadline(_ time.Time) (err error) { return }

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

		packet, _, err := server.ReceivePacketFromConn(c)
		if err != nil {
			t.Errorf("Server error: %v", err)
			return
		}
		if packet == nil {
			t.Error("nil packet")
			return
		}
		packet, _, err = server.ReceivePacketFromConn(c)
		if err != nil {
			t.Errorf("Server error: %v", err)
			return
		}
		if packet == nil {
			t.Error("nil packet")
			return
		}
		packet, _, err = server.ReceivePacketFromConn(c)
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
			client, _ := Dial("localhost:6677")
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
			client, _ := Dial("localhost:6677")
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
		p, _, err := server.ReceivePacketFromConn(c)
		if err != nil {
			t.Errorf("server error: %v", err)
			return
		}
		if got, want := p.(*Message).Address, "/address/test1"; got != want {
			t.Errorf("wrong address; got = %s, want = %s", got, want)
			return
		}

		// Second receive should time out since client is delayed 150 milliseconds
		if _, _, err = server.ReceivePacketFromConn(c); err == nil {
			t.Errorf("expected error")
			return
		}

		// Next receive should get it
		p, _, err = server.ReceivePacketFromConn(c)
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

//func TestServer_Close(t *testing.T) {
//	s := &Server{Handler: func(_ Packet, _ net.Addr) {}, Addr: ":10000"}
//	go s.ListenAndServe()
//	if s.conn == nil {
//		t.Fatal("Close(): server not started")
//	}
//	if err := s.Close(); (err != nil) != false {
//		t.Errorf("Close() error = %v, wantErr %v", err, false)
//	}
//	if s.conn != nil {
//		t.Error("Close(): failed to nil conn")
//	}
//}

func TestServer_ListenAndServe(t *testing.T) {
	type fields struct {
		Addr        string
		Handler     HandlerFunc
		ReadTimeout time.Duration
		conn        net.PacketConn
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Addr:        tt.fields.Addr,
				Handler:     tt.fields.Handler,
				ReadTimeout: tt.fields.ReadTimeout,
				conn:        tt.fields.conn,
			}
			if err := s.ListenAndServe(); (err != nil) != tt.wantErr {
				t.Errorf("ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_ReceivePacketFromConn(t *testing.T) {
	type fields struct {
		Addr        string
		Handler     HandlerFunc
		ReadTimeout time.Duration
		conn        net.PacketConn
	}
	type args struct {
		conn net.PacketConn
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Packet
		want1   net.Addr
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Addr:        tt.fields.Addr,
				Handler:     tt.fields.Handler,
				ReadTimeout: tt.fields.ReadTimeout,
				conn:        tt.fields.conn,
			}
			got, got1, err := s.ReceivePacketFromConn(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceivePacketFromConn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReceivePacketFromConn() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ReceivePacketFromConn() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestServer_Serve(t *testing.T) {
	type fields struct {
		Addr        string
		Handler     HandlerFunc
		ReadTimeout time.Duration
		conn        net.PacketConn
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Addr:        tt.fields.Addr,
				Handler:     tt.fields.Handler,
				ReadTimeout: tt.fields.ReadTimeout,
				conn:        tt.fields.conn,
			}
			if err := s.Serve(); (err != nil) != tt.wantErr {
				t.Errorf("Serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_WriteTo(t *testing.T) {
	type fields struct {
		Addr        string
		Handler     HandlerFunc
		ReadTimeout time.Duration
		conn        net.PacketConn
	}
	type args struct {
		p    Packet
		addr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Addr:        tt.fields.Addr,
				Handler:     tt.fields.Handler,
				ReadTimeout: tt.fields.ReadTimeout,
				conn:        tt.fields.conn,
			}
			got, err := s.WriteTo(tt.args.p, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WriteTo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_readFromConnection(t *testing.T) {
	type fields struct {
		Addr        string
		Handler     HandlerFunc
		ReadTimeout time.Duration
		conn        net.PacketConn
	}
	tests := []struct {
		name    string
		fields  fields
		want    Packet
		want1   net.Addr
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Addr:        tt.fields.Addr,
				Handler:     tt.fields.Handler,
				ReadTimeout: tt.fields.ReadTimeout,
				conn:        tt.fields.conn,
			}
			got, got1, err := s.readFromConnection()
			if (err != nil) != tt.wantErr {
				t.Errorf("readFromConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readFromConnection() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("readFromConnection() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

// The following test relies on having a more solid logging framework
//func Test_recoverer(t *testing.T) {
//	type args struct {
//		a net.Addr
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			recoverer(tt.args.a)
//		})
//	}
//}

func BenchmarkReceivePacketFromConn(b *testing.B) {
	d := &dummyConn{m: msg}
	s := &Server{}
	var p Packet
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p, _, _ = s.ReceivePacketFromConn(d)
	}
	result = p
}
