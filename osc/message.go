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
		if _, err := GetTypeTag(a); err != nil {
			return err
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
		case int32:
			buf := make([]byte, bit32Size)
			binary.BigEndian.PutUint32(buf, uint32(t))
			b.Write(buf)
		case float32:
			buf := make([]byte, bit32Size)
			binary.BigEndian.PutUint32(buf, *(*uint32)(unsafe.Pointer(&t)))
			b.Write(buf)
		case int64:
			buf := make([]byte, bit64Size)
			binary.BigEndian.PutUint64(buf, uint64(t))
			b.Write(buf)
		case float64:
			buf := make([]byte, bit64Size)
			binary.BigEndian.PutUint64(buf, *(*uint64)(unsafe.Pointer(&t)))
			b.Write(buf)
		case string:
			writePaddedString(t, b)
		case []byte:
			if _, err := writeBlob(t, b); err != nil {
				return err
			}
		case Timetag:
			buf := make([]byte, bit64Size)
			binary.BigEndian.PutUint64(buf, uint64(t))
			b.Write(buf)
		}
	}

	if b.Len() >= MaxPacketSize {
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

	if data.Len() >= MaxPacketSize {
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

// UnmarshalBinary implements the encoding.BinaryUnmarshaler
func (m *Message) UnmarshalBinary(data []byte) error {
	if data[0] != '/' {
		return fmt.Errorf("UnmarshalBinary: data not a valid OSC message")
	}

	if (len(data) % bit32Size) != 0 {
		return fmt.Errorf("UnmarshalBinary: data isn't mod 4")
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
	if err = m.readArguments(b); err != nil {
		return fmt.Errorf("UnmarshalBinary: %w", err)
	}

	return nil
}

// readArguments from `reader` and add them to the OSC message `msg`.
func (m *Message) readArguments(reader *bytes.Buffer) error {
	// Read the type tag string
	typetags, _, err := readPaddedString(reader)
	if err != nil {
		return fmt.Errorf("readArguments: %w", err)
	}

	if len(typetags) == 0 {
		return nil
	}

	// If the typetag doesn't start with ',', it's not valid
	if typetags[0] != ',' {
		return fmt.Errorf("unsupported typetag string: %s", typetags)
	}

	m.Arguments = make([]interface{}, 0, len(typetags)-1)

	for _, c := range typetags[1:] {
		if reader.Len() < bit32Size {
			return fmt.Errorf("readArguments: not enough bits to read")
		}
		switch c {
		default:
			return fmt.Errorf("unsupported typetag: %c", c)

		case 'i': // int32
			m.Arguments = append(m.Arguments, int32(binary.BigEndian.Uint32(reader.Next(bit32Size))))

		case 'h': // int64
			m.Arguments = append(m.Arguments, int64(binary.BigEndian.Uint64(reader.Next(bit64Size))))

		case 'f': // float32
			f := binary.BigEndian.Uint32(reader.Next(bit32Size))
			m.Arguments = append(m.Arguments, *(*float32)(unsafe.Pointer(&f)))

		case 'd': // float64/double
			f := binary.BigEndian.Uint64(reader.Next(bit64Size))
			m.Arguments = append(m.Arguments, *(*float64)(unsafe.Pointer(&f)))

		case 's': // string
			str, err := reader.ReadString(0)
			if err != nil {
				return err
			}
			// Remove the padding bytes
			reader.Next(padBytesNeeded(len(str)))
			str = str[:len(str)-1]

			m.Arguments = append(m.Arguments, str)

		case 'b': // blob
			var buf []byte
			if buf, _, err = readBlob(reader); err != nil {
				return fmt.Errorf("readArguments: %w", err)
			}
			m.Arguments = append(m.Arguments, buf)

		case 't': // OSC time tag
			m.Arguments = append(m.Arguments, Timetag(binary.BigEndian.Uint64(reader.Next(bit64Size))))

		case 'N': // nil
			m.Arguments = append(m.Arguments, nil)

		case 'T': // true
			m.Arguments = append(m.Arguments, true)

		case 'F': // false
			m.Arguments = append(m.Arguments, false)
		}
	}

	return nil
}
