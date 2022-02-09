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
	Handler     HandlerFunc
	ReadTimeout time.Duration
	conn        net.PacketConn
}

// ListenAndServe quickly sets up an OSC Server listening on addr.
func ListenAndServe(addr string, handler HandlerFunc) error {
	s := &Server{addr, handler, 0, nil}
	return s.ListenAndServe()
}

// ListenAndServe listens for OSC packets and sends them to the specified handler.
func (s *Server) ListenAndServe() error {
	if s.Handler == nil {
		return fmt.Errorf("ListenAndServe: missing handler")
	}

	var err error
	s.conn, err = net.ListenPacket("udp", s.Addr)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Serve()
}

// Serve retrieves incoming OSC packets.
func (s *Server) Serve() error {
	var falloff time.Duration
	for {
		msg, addr, err := s.readFromConnection()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if falloff == 0 {
					falloff = 5 * time.Millisecond
				} else {
					falloff *= 2
				}
				if max := 1 * time.Second; falloff > max {
					falloff = max
				}
				time.Sleep(falloff)
				continue
			} else if !ok {
				continue // TODO: allow logging of invalid packets
			}
			return err
		}
		falloff = 0
		go func() {
			defer recoverer(addr)
			s.Handler(msg, addr)
		}()
	}
}

func recoverer(a net.Addr) {
	if err := recover(); err != nil {
		buf := make([]byte, MaxPacketSize+29)
		buf = buf[:runtime.Stack(buf, false)]
		fmt.Printf("osc: panic handling from %s: %v\n%s\n", a, err, buf) // TODO: figure out a better logger
	}
}

// Close allows you to close the Server connection.
func (s *Server) Close() error {
	if err := s.conn.Close(); err != nil {
		return err
	}
	s.conn = nil
	return nil
}

// WriteTo allows you to reuse the Server connection for sending Packets.
func (s *Server) WriteTo(p Packet, addr string) (int, error) {
	b, err := p.MarshalBinary()
	if err != nil {
		return 0, err
	}
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return 0, err
	}
	return s.conn.WriteTo(b, a)
}

// ReceivePacketFromConn reads a single packet from conn.
func (s *Server) ReceivePacketFromConn(conn net.PacketConn) (Packet, net.Addr, error) {
	c := s.conn
	defer func() { s.conn = c }()
	s.conn = conn
	return s.readFromConnection()
}

// readFromConnection retrieves OSC packets.
func (s *Server) readFromConnection() (Packet, net.Addr, error) {
	if s.ReadTimeout != 0 {
		if err := s.conn.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
			return nil, nil, err
		}
	}

	b := bufPool.Get().(*[]byte)
	defer bufPool.Put(b)

	n, a, err := s.conn.ReadFrom(*b)
	if err != nil {
		return nil, a, err
	}
	bb := make([]byte, n)
	copy(bb, *b)

	p, err := parsePacket(bb)
	return p, a, err
}

// Handler is the interface for handling an OSC Packet.
type Handler interface {
	Handle(Packet, net.Addr)
}

// HandlerFunc implements the Handler interface. Type definition for an OSC handler function.
type HandlerFunc func(Packet, net.Addr)

// Handle calls itself with the given Packet. Implements the Handler interface.
func (f HandlerFunc) Handle(p Packet, a net.Addr) {
	f(p, a)
}
