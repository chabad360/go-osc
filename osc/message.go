package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Message represents a single OSC message. An OSC message consists of an OSC
// address pattern and zero or more arguments.
type Message struct {
	Address   string
	Arguments []interface{}
}

// Verify that Messages implements the Packet interface.
var _ Packet = (*Message)(nil)

func (m *Message) Clear() {
	m.Address = ""
	m.Arguments = m.Arguments[:0]
}

// NewMessage returns a new Message. The address parameter is the OSC address.
func NewMessage(addr string, args ...interface{}) *Message {
	return &Message{Address: addr, Arguments: args}
}

// Append appends the given arguments to the arguments list.
func (m *Message) Append(args ...interface{}) error {
	for _, a := range args {
		if t := ToTypeTag(a); t == 0 {
			return fmt.Errorf("unsupported type: %T", a)
		}
	}
	m.Arguments = append(m.Arguments, args...)
	return nil
}

// Match returns true, if the OSC address pattern of the OSC Message matches the given
// address. The match is case sensitive!
func (m *Message) Match(addr string) bool {
	regexp, err := getRegEx(m.Address)
	if err != nil {
		return false
	}
	return regexp.MatchString(addr)
}

// TypeTags returns the type tag string.
func (m *Message) TypeTags() (string, error) {
	if m == nil {
		return "", fmt.Errorf("TypeTags: message is nil")
	}

	tags := make([]byte, 0, len(m.Arguments)+1)
	tags = append(tags, ',')
	for _, args := range m.Arguments {
		s := ToTypeTag(args)
		if s == 0 {
			return "", fmt.Errorf("unsupported type: %T", args)
		}
		tags = append(tags, byte(s))
	}

	return *(*string)(unsafe.Pointer(&tags)), nil
}

// String implements the fmt.Stringer interface.
func (m *Message) String() string {
	if m == nil {
		return ""
	}

	tags, _ := m.TypeTags()

	strBuf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(strBuf)
	strBuf.Reset()

	strBuf.WriteString(m.Address)
	if len(tags) == 0 {
		return strBuf.String()
	}

	strBuf.WriteByte(' ')
	strBuf.WriteString(tags)

	for _, arg := range m.Arguments {
		switch arg := arg.(type) {
		case bool, int32, int64, float32, float64, string:
			fmt.Fprintf(strBuf, " %v", arg)

		case nil:
			strBuf.WriteString(" Nil")

		case []byte:
			strBuf.WriteString(" blob")

		case Timetag:
			fmt.Fprintf(strBuf, " %d", arg.TimeTag())
		}
	}

	return strBuf.String()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (m *Message) MarshalBinary() ([]byte, error) {
	buf := bPool.Get().([]byte)

	n := writePaddedString(m.Address, buf)

	// Write the type tag string to the data buffer
	nn, err := writeTypeTags(m.Arguments, buf[n:])
	if err != nil {
		return nil, err
	}

	n += nn

	// Process the type tags and collect all arguments
	for _, arg := range m.Arguments {
		switch t := arg.(type) {
		default:
			return nil, fmt.Errorf("MarshalBinary: unsupported type: %T", t)

		case bool, nil:
			continue
		case int32:
			binary.BigEndian.PutUint32(buf[n:], uint32(t))
			n += bit32Size
		case float32:
			binary.BigEndian.PutUint32(buf[n:], *(*uint32)(unsafe.Pointer(&t)))
			n += bit32Size
		case int64:
			binary.BigEndian.PutUint64(buf[n:], uint64(t))
			n += bit64Size
		case float64:
			binary.BigEndian.PutUint64(buf[n:], *(*uint64)(unsafe.Pointer(&t)))
			n += bit64Size
		case string:
			n += writePaddedString(t, buf[n:])
		case []byte:
			if len(t) > MaxPacketSize-n {
				return nil, fmt.Errorf("MarshalBinary: blob makes packet too large")
			}
			n += writeBlob(t, buf[n:])
		case Timetag:
			binary.BigEndian.PutUint64(buf[n:], uint64(t))
			n += bit64Size
		}
	}

	b := make([]byte, n)
	copy(b, buf)
	bPool.Put(buf)

	return b, nil
}

// NewMessageFromData returns a new OSC message created from the parsed data.
func NewMessageFromData(data []byte) (msg *Message, err error) {
	msg = &Message{}
	if err = msg.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return msg, nil
}

// newMessageFromData assumes that the bytes have already been copied.
func newMessageFromData(data []byte) (msg *Message, err error) {
	msg = &Message{}
	if err = msg.unmarshalBinary(data); err != nil {
		return nil, err
	}
	return msg, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (m *Message) UnmarshalBinary(d []byte) error {
	data := make([]byte, len(d))
	copy(data, d)

	return m.unmarshalBinary(data)
}

// unmarshalBinary is the actual implementation, it doesn't copy, so we can use a single copy for bundles.
func (m *Message) unmarshalBinary(data []byte) error {
	if data[0] != '/' {
		return fmt.Errorf("UnmarshalBinary: data not a valid OSC message")
	}

	if (len(data) % bit32Size) != 0 {
		return fmt.Errorf("UnmarshalBinary: data isn't mod 4") //TODO: error hygiene
	}

	// First, read the OSC address
	addr, n, err := parsePaddedString(data)
	if err != nil {
		return fmt.Errorf("UnmarshalBinary: %w", err)
	}

	// parse all arguments
	m.Address = addr
	data = data[n:]
	if err = m.parseArguments(data); err != nil {
		return fmt.Errorf("UnmarshalBinary: %w", err)
	}

	return nil
}

// parseArguments parses a []byte with OSC arguments and add them to the OSC message `msg`.
func (m *Message) parseArguments(data []byte) error {
	// parse the type tag string
	typetags, n, err := parsePaddedString(data)
	if err != nil {
		return fmt.Errorf("parseArguments: %w", err)
	}
	data = data[n:]

	if len(typetags) == 0 {
		return nil
	}

	// If the typetag doesn't start with ',', it's not valid
	if typetags[0] != ',' {
		return fmt.Errorf("unsupported typetag string: %s", typetags)
	}

	typetags = typetags[1:]

	m.Arguments = make([]interface{}, 0, len(typetags))

	for _, c := range typetags {
		if len(data) < bit32Size {
			return fmt.Errorf("parseArguments: not enough bits to read")
		}
		switch TypeTag(c) {
		default:
			return fmt.Errorf("unsupported typetag: %c", c)

		case Int32:
			m.Arguments = append(m.Arguments, int32(binary.BigEndian.Uint32(data[:bit32Size])))
			data = data[bit32Size:]

		case Int64:
			if len(data) < bit64Size {
				return fmt.Errorf("readArguments: not enough bits to read")
			}
			m.Arguments = append(m.Arguments, int64(binary.BigEndian.Uint64(data[:bit64Size])))
			data = data[bit64Size:]

		case Float32:
			b := binary.BigEndian.Uint32(data[:bit32Size])
			m.Arguments = append(m.Arguments, *(*float32)(unsafe.Pointer(&b)))
			data = data[bit32Size:]

		case Float64:
			if len(data) < bit64Size {
				return fmt.Errorf("readArguments: not enough bits to read")
			}
			f := binary.BigEndian.Uint64(data[:bit64Size])
			m.Arguments = append(m.Arguments, *(*float64)(unsafe.Pointer(&f)))
			data = data[bit64Size:]

		case String:
			str, n, err := parsePaddedString(data)
			if err != nil {
				return fmt.Errorf("readArguments: %w", err)
			}
			m.Arguments = append(m.Arguments, str)
			data = data[n:]

		case Blob:
			buf, n, err := parseBlob(data)
			if err != nil {
				return fmt.Errorf("readArguments: %w", err)
			}
			m.Arguments = append(m.Arguments, buf)
			data = data[n:]

		case TimeTag:
			if len(data) < bit64Size {
				return fmt.Errorf("readArguments: not enough bits to read")
			}
			m.Arguments = append(m.Arguments, Timetag(binary.BigEndian.Uint64(data[:bit64Size])))
			data = data[bit64Size:]

		case Nil:
			m.Arguments = append(m.Arguments, nil)

		case True:
			m.Arguments = append(m.Arguments, true)

		case False:
			m.Arguments = append(m.Arguments, false)
		}
	}

	return nil
}
