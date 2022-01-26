package osc

import (
	"bytes"
	"io"
	"net"
	"time"
)

// Server represents an OSC server. The server listens on Address and Port for
// incoming OSC packets and bundles.
type Server struct {
	Addr        string
	Dispatcher  Dispatcher
	ReadTimeout time.Duration
	close       func() error
}

// ListenAndServe retrieves incoming OSC packets and dispatches the retrieved
// OSC packets.
func (s *Server) ListenAndServe() error {
	if s.Dispatcher == nil {
		s.Dispatcher = NewStandardDispatcher()
	}

	ln, err := net.ListenPacket("udp", s.Addr)
	if err != nil {
		return err
	}
	s.close = ln.Close
	defer ln.Close()

	return s.Serve(ln)
}

// Close shuts down the server and closes it's upd connection
func (s *Server) Close() error {
	if s.close != nil {
		return s.close()
	}
	return nil
}

// Serve retrieves incoming OSC packets from the given connection and dispatches
// retrieved OSC packets. If something goes wrong an error is returned.
func (s *Server) Serve(c net.PacketConn) error {
	var tempDelay time.Duration
	for {
		msg, err := s.readFromConnection(c)
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
			}
			return err
		}
		tempDelay = 0
		go s.Dispatcher.Dispatch(msg)
	}
}

// ReceivePacket listens for incoming OSC packets and returns the packet if one is received.
func (s *Server) ReceivePacket(c net.PacketConn) (Packet, error) {
	return s.readFromConnection(c)
}

type eofReader struct {
	net.PacketConn
}

func (g eofReader) Read(buf []byte) (int, error) {
	n, _, err := g.ReadFrom(buf)
	if err == nil {
		return n, io.EOF
	}
	return n, err
}

// readFromConnection retrieves OSC packets.
func (s *Server) readFromConnection(c net.PacketConn) (Packet, error) {
	if s.ReadTimeout != 0 {
		if err := c.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
			return nil, err
		}
	}

	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()
	_, err := b.ReadFrom(eofReader{c})
	if err != nil {
		return nil, err
	}

	return ReadPacket(b.Bytes())
}
