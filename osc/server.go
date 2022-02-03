package osc

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

// Server represents an OSC server. The server listens on Address and Port for incoming OSC packets and bundles.
type Server struct {
	Addr        string
	Dispatcher  Dispatcher
	ReadTimeout time.Duration
}

// ListenAndServe retrieves incoming OSC packets and dispatches the retrieved OSC packets.
func (s *Server) ListenAndServe() error {
	if s.Dispatcher == nil {
		s.Dispatcher = NewStandardDispatcher()
	}

	ln, err := net.ListenPacket("udp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	return s.Serve(ln)
}

// Serve retrieves incoming OSC packets from the given connection and dispatches retrieved OSC packets.
// If something goes wrong an error is returned.
func (s *Server) Serve(c net.PacketConn) error {
	var tempDelay time.Duration
	for {
		msg, addr, err := s.readFromConnection(c)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			} else if !ok {
				continue // TODO: allow logging of packet errors
			}
			return err
		}
		tempDelay = 0
		go s.serve(msg, addr)
	}
}

func (s *Server) serve(m Packet, a net.Addr) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, MaxPacketSize+29)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("osc: panic handling from %s: %v\n%s\n", a, err, buf) // TODO: figure out a better logger
		}
	}()
	s.Dispatcher.Dispatch(m)
}

// ReceivePacket listens for incoming OSC packets and returns the packet if one is received.
func (s *Server) ReceivePacket(c net.PacketConn) (Packet, net.Addr, error) {
	return s.readFromConnection(c)
}

// readFromConnection retrieves OSC packets.
func (s *Server) readFromConnection(c net.PacketConn) (Packet, net.Addr, error) {
	if s.ReadTimeout != 0 {
		if err := c.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
			return nil, nil, err
		}
	}

	b := bufPool.Get().(*[]byte)
	defer bufPool.Put(b)

	n, a, err := c.ReadFrom(*b)
	if err != nil {
		return nil, a, err
	}
	bb := make([]byte, n)
	copy(bb, *b)

	p, err := parsePacket(bb)
	return p, a, err
}
