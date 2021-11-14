package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	initBuf = make([]byte, 65535)
	//buf     = bytes.NewBuffer(make([]byte, 65535))
	l       sync.Mutex
	bufPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

// Server represents an OSC server. The server listens on Address and Port for
// incoming OSC packets and bundles.
type Server struct {
	Addr        string
	Dispatcher  Dispatcher
	ReadTimeout time.Duration
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
	defer ln.Close()

	return s.Serve(ln)
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
	l.Lock()
	defer l.Unlock()
	if s.ReadTimeout != 0 {
		if err := c.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
			return nil, err
		}
	}

	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()
	n, err := b.ReadFrom(&eofReader{c})
	if err != nil {
		return nil, err
	}

	var start int
	return ReadPacket(b, &start, int(n))
}

// ParsePacket parses the given msg string and returns a Packet
func ParsePacket(msg string) (Packet, error) {
	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()

	b.WriteString(msg)

	var start int
	p, err := ReadPacket(b, &start, b.Len())
	return p, err
}

// ReadPacket parses an OSC packet from the given reader.
func ReadPacket(reader *bytes.Buffer, start *int, end int) (Packet, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	err = reader.UnreadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '/': // An OSC Message starts with a '/'
		return readMessage(reader, start)
	case '#': // An OSC bundle starts with a '#'
		return readBundle(reader, start, end)
	}

	return nil, fmt.Errorf("ReadPacket: invalid packet")
}

// readBundle reads a Bundle from reader.
func readBundle(reader *bytes.Buffer, start *int, end int) (*Bundle, error) {
	// Read the '#bundle' OSC string
	startTag, n, err := readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	if startTag != bundleTagString {
		return nil, fmt.Errorf("invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	var timeTag uint64
	if err = binary.Read(reader, binary.BigEndian, &timeTag); err != nil {
		return nil, err
	}
	*start += 8

	// Create a new bundle
	bundle := Bundle{Timetag: *NewTimetagFromTimetag(timeTag)}

	// Read until the end of the buffer
	for *start < end {
		// Read the size of the bundle element
		var length int32
		if err = binary.Read(reader, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		*start += 4

		var p Packet
		p, err = ReadPacket(reader, start, end)
		if err != nil {
			return nil, err
		}
		if err = bundle.Append(p); err != nil {
			return nil, err
		}
	}

	return &bundle, nil
}

var msgs = &Message{}

// readMessage from `reader`.
func readMessage(reader *bytes.Buffer, start *int) (*Message, error) {
	// First, read the OSC address
	addr, n, err := readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	// Read all arguments
	msg := Message{}
	msg.Address = addr
	//msgs.Arguments = msgs.Arguments[:0]
	if msg.Arguments, err = readArguments(reader, start); err != nil {
		return nil, err
	}

	//msg := *msgs
	return &msg, nil
}

// readArguments from `reader` and add them to the OSC message `msg`.
func readArguments(reader *bytes.Buffer, start *int) ([]interface{}, error) {
	// Read the type tag string
	typetags, n, err := readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	if len(typetags) == 0 {
		return nil, nil
	}

	// If the typetag doesn't start with ',', it's not valid
	if typetags[0] != ',' {
		return nil, fmt.Errorf("unsupported type tag string: %s", typetags)
	}

	// Remove ',' from the type tag
	tt := typetags[1:]

	args := make([]interface{}, 0, len(tt))

	for _, c := range tt {
		switch c {
		default:
			return nil, fmt.Errorf("unsupported type tag: %c", c)

		case 'i': // int32
			var i int32
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return nil, err
			}
			*start += 4
			args = append(args, i)

		case 'h': // int64
			var i int64
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return nil, err
			}
			*start += 8
			args = append(args, i)

		case 'f': // float32
			var f float32
			if err = binary.Read(reader, binary.BigEndian, &f); err != nil {
				return nil, err
			}
			*start += 4
			args = append(args, f)

		case 'd': // float64/double
			var d float64
			if err = binary.Read(reader, binary.BigEndian, &d); err != nil {
				return nil, err
			}
			*start += 8
			args = append(args, d)

		case 's': // string
			var s string
			var n int
			if s, n, err = readPaddedString(reader); err != nil {
				return nil, err
			}
			*start += n
			args = append(args, s)

		case 'b': // blob
			var buf []byte
			var n int
			if buf, n, err = readBlob(reader); err != nil {
				return nil, err
			}
			*start += n
			args = append(args, buf)

		case 't': // OSC time tag
			var tt uint64
			if err = binary.Read(reader, binary.BigEndian, &tt); err != nil {
				return nil, err
			}
			*start += 8
			args = append(args, *NewTimetagFromTimetag(tt))

		case 'N': // nil
			args = append(args, nil)

		case 'T': // true
			args = append(args, true)

		case 'F': // false
			args = append(args, false)
		}
	}

	return args, nil
}
