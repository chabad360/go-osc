package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
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
		if _, err := GetTypeTag(a); err != nil {
			return err
		}
	}
	m.Arguments = append(m.Arguments, args...)
	return nil
}

// Equals returns true if the given OSC Message `m` is equal to the current OSC
// Message. It checks if the OSC address and the arguments are equal. Returns
// true if the current object and `m` are equal.
func (m *Message) Equals(msg *Message) bool {
	return reflect.DeepEqual(m, msg)
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
		s, err := GetTypeTag(args)
		if err != nil {
			return "", err
		}
		tags = append(tags, s...)
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
		switch arg.(type) {
		case bool, int32, int64, float32, float64, string:
			fmt.Fprintf(strBuf, " %v", arg)

		case nil:
			strBuf.WriteString(" Nil")

		case []byte:
			strBuf.WriteString(" blob")

		case Timetag:
			timeTag := arg.(Timetag)
			fmt.Fprintf(strBuf, " %d", timeTag.TimeTag())
		}
	}

	return strBuf.String()
}

// CountArguments returns the number of arguments.
func (m *Message) CountArguments() int {
	return len(m.Arguments)
}

// MarshalBinary serializes the OSC message to a byte buffer. The byte buffer
// has the following format:
// 1. OSC Address Pattern
// 2. OSC Type Tag String
// 3. OSC Arguments
func (m *Message) MarshalBinary() (b []byte, err error) {
	// We can start with the OSC address and add it to the buffer
	data := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(data)
	data.Reset()

	if err = m.LightMarshalBinary(data); err != nil {
		return nil, err
	}
	return append(b, data.Bytes()...), nil
}

func (m *Message) LightMarshalBinary(data *bytes.Buffer) error {
	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()

	// Process the type tags and collect all arguments
	for _, arg := range m.Arguments {
		switch t := arg.(type) {
		default:
			return fmt.Errorf("LightMarshalBinary: unsupported type: %T", t)

		case bool, nil:
			continue
		case float32, float64, int32, int64:
			if err := binary.Write(b, binary.BigEndian, t); err != nil {
				return err
			}

		case string:
			writePaddedString(t, b)
		case []byte:
			if _, err := writeBlob(t, b); err != nil {
				return err
			}
		case Timetag:
			tt, err := t.MarshalBinary()
			if err != nil {
				return err
			}
			b.Write(tt)
		}
	}

	if b.Len() >= len(initBuf) {
		return fmt.Errorf("LightMarshalBinary: payload too large: %d", b.Len())
	}

	writePaddedString(m.Address, data)

	// Write the type tag string to the data buffer
	typetags, err := m.TypeTags()
	if err != nil {
		return err
	}
	writePaddedString(typetags, data)

	// Write the payload (OSC arguments) to the data buffer
	data.Write(b.Bytes())

	if data.Len() >= len(initBuf) {
		return fmt.Errorf("LightMarshalBinary: packet too large: %d", data.Len())
	}

	return nil
}

func NewMessageFromData(data []byte) (msg *Message, err error) {
	msg = &Message{}
	if err = msg.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *Message) UnmarshalBinary(data []byte) error {
	if data[0] != '/' {
		return fmt.Errorf("data not a valid OSC message")
	}

	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()

	b.Write(data)

	// First, read the OSC address
	addr, _, err := readPaddedString(b)
	if err != nil {
		return fmt.Errorf("UnmarshalBinary: %w", err)
	}

	// Read all arguments
	m.Address = addr
	if m.Arguments, err = readArguments(b); err != nil {
		return fmt.Errorf("UnmarshalBinary: %w", err)
	}

	return nil
}

// readArguments from `reader` and add them to the OSC message `msg`.
func readArguments(reader *bytes.Buffer) ([]interface{}, error) {
	// Read the type tag string
	typetags, _, err := readPaddedString(reader)
	if err != nil {
		return nil, fmt.Errorf("readArguments: %w", err)
	}

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
				return nil, fmt.Errorf("readArguments: %w", err)
			}
			args = append(args, i)

		case 'h': // int64
			var i int64
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return nil, fmt.Errorf("readArguments: %w", err)
			}
			args = append(args, i)

		case 'f': // float32
			var f float32
			if err = binary.Read(reader, binary.BigEndian, &f); err != nil {
				return nil, fmt.Errorf("readArguments: %w", err)
			}
			args = append(args, f)

		case 'd': // float64/double
			var d float64
			if err = binary.Read(reader, binary.BigEndian, &d); err != nil {
				return nil, fmt.Errorf("readArguments: %w", err)
			}
			args = append(args, d)

		case 's': // string
			var s string
			if s, _, err = readPaddedString(reader); err != nil {
				return nil, fmt.Errorf("readArguments: %w", err)
			}
			args = append(args, s)

		case 'b': // blob
			var buf []byte
			if buf, _, err = readBlob(reader); err != nil {
				return nil, fmt.Errorf("readArguments: %w", err)
			}
			args = append(args, buf)

		case 't': // OSC time tag
			var tt uint64
			if err = binary.Read(reader, binary.BigEndian, &tt); err != nil {
				return nil, fmt.Errorf("readArguments: %w", err)
			}
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
