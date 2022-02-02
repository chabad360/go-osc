package osc

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Message represents a single OSC Message.
type Message struct {
	Address   string
	Arguments []interface{}
}

// Verify that Messages implements the Packet interface.
var _ Packet = (*Message)(nil)

func (m *Message) Clear() {
	m.Address = ""
	m.Arguments = nil
}

// NewMessage returns a new Message.
// addr is the OSC address, args are the OSC to add to message.
func NewMessage(addr string, args ...interface{}) *Message {
	if len(args) == 0 {
		return &Message{Address: addr, Arguments: []interface{}{}}
	}
	return &Message{Address: addr, Arguments: args}
}

// Append appends the given arguments to the arguments list.
func (m *Message) Append(args ...interface{}) error {
	for _, a := range args {
		if t := ToTypeTag(a); t == TypeInvalid {
			return fmt.Errorf("unsupported type: %T", a)
		}
	}
	m.Arguments = append(m.Arguments, args...)
	return nil
}

// Match returns true, if the OSC address pattern matches addr. The match is case-sensitive!
func (m *Message) Match(addr string) bool {
	regexp, err := getRegEx(m.Address)
	if err != nil {
		return false
	}
	return regexp.MatchString(addr)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (m *Message) MarshalBinary() ([]byte, error) {
	buf := bPool.Get().(*[]byte)
	copy(*buf, empty[:])

	n := writePaddedString(m.Address, *buf)

	// Write the type tag string to the data buffer
	nn, err := writeTypeTags(m.Arguments, (*buf)[n:])
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
			binary.BigEndian.PutUint32((*buf)[n:], uint32(t))
			n += bit32Size
		case int64:
			binary.BigEndian.PutUint64((*buf)[n:], uint64(t))
			n += bit64Size
		case float32:
			binary.BigEndian.PutUint32((*buf)[n:], *(*uint32)(unsafe.Pointer(&t)))
			n += bit32Size
		case float64:
			binary.BigEndian.PutUint64((*buf)[n:], *(*uint64)(unsafe.Pointer(&t)))
			n += bit64Size
		case string:
			n += writePaddedString(t, (*buf)[n:])
		case []byte:
			if len(t) > MaxPacketSize-n {
				return nil, fmt.Errorf("MarshalBinary: blob makes packet too large")
			}
			n += writeBlob(t, (*buf)[n:])
		case Timetag:
			binary.BigEndian.PutUint64((*buf)[n:], uint64(t))
			n += bit64Size
		}
	}

	b := make([]byte, n)
	copy(b, *buf)
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
		return fmt.Errorf("bad typetag string: %s", typetags)
	}

	typetags = typetags[1:]

	m.Arguments = make([]interface{}, 0, len(typetags))

	for _, c := range typetags {
		switch TypeTag(c) {
		default:
			return fmt.Errorf("unsupported typetag: %c", c)

		case TypeInt32:
			if len(data) < bit32Size {
				return fmt.Errorf("parseArguments: not enough bits to read: %v", data)
			}
			m.Arguments = append(m.Arguments, int32(binary.BigEndian.Uint32(data[:bit32Size])))
			data = data[bit32Size:]

		case TypeInt64:
			if len(data) < bit64Size {
				return fmt.Errorf("readArguments: not enough bits to read: %v", data)
			}
			m.Arguments = append(m.Arguments, int64(binary.BigEndian.Uint64(data[:bit64Size])))
			data = data[bit64Size:]

		case TypeFloat32:
			if len(data) < bit32Size {
				return fmt.Errorf("parseArguments: not enough bits to read: %v", data)
			}
			b := binary.BigEndian.Uint32(data[:bit32Size])
			m.Arguments = append(m.Arguments, *(*float32)(unsafe.Pointer(&b)))
			data = data[bit32Size:]

		case TypeFloat64:
			if len(data) < bit64Size {
				return fmt.Errorf("readArguments: not enough bits to read: %v", data)
			}
			f := binary.BigEndian.Uint64(data[:bit64Size])
			m.Arguments = append(m.Arguments, *(*float64)(unsafe.Pointer(&f)))
			data = data[bit64Size:]

		case TypeString:
			if len(data) < bit32Size {
				return fmt.Errorf("parseArguments: not enough bits to read: %v", data)
			}
			str, n, err := parsePaddedString(data)
			if err != nil {
				return fmt.Errorf("readArguments: %w", err)
			}
			m.Arguments = append(m.Arguments, str)
			data = data[n:]

		case TypeBlob:
			if len(data) < bit64Size {
				return fmt.Errorf("parseArguments: not enough bits to read: %v", data)
			}
			buf, n, err := parseBlob(data)
			if err != nil {
				return fmt.Errorf("readArguments: %w", err)
			}
			m.Arguments = append(m.Arguments, buf)
			data = data[n:]

		case TypeTimeTag:
			if len(data) < bit64Size {
				return fmt.Errorf("readArguments: not enough bits to read: %v", data)
			}
			m.Arguments = append(m.Arguments, Timetag(binary.BigEndian.Uint64(data[:bit64Size])))
			data = data[bit64Size:]

		case TypeNil:
			m.Arguments = append(m.Arguments, nil)

		case TypeTrue:
			m.Arguments = append(m.Arguments, true)

		case TypeFalse:
			m.Arguments = append(m.Arguments, false)
		}
	}

	return nil
}
