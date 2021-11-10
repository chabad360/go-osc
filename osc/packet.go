package osc

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

// Packet is the interface for Message and Bundle.
type Packet interface {
	encoding.BinaryMarshaler
}

// Message represents a single OSC message. An OSC message consists of an OSC
// address pattern and zero or more arguments.
type Message struct {
	Address   string
	Arguments []interface{}
}

// Verify that Messages implements the Packet interface.
var _ Packet = (*Message)(nil)

// Bundle represents an OSC bundle. It consists of the OSC-string "#bundle"
// followed by an OSC Time Tag, followed by zero or more OSC bundle/message
// elements. The OSC-timetag is a 64-bit fixed point time tag. See
// http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag  Timetag
	Messages []*Message
	Bundles  []*Bundle
}

// Verify that Bundle implements the Packet interface.
var _ Packet = (*Bundle)(nil)

////
// Message
////

// NewMessage returns a new Message. The address parameter is the OSC address.
func NewMessage(addr string, args ...interface{}) *Message {
	return &Message{Address: addr, Arguments: args}
}

// Append appends the given arguments to the arguments list.
func (msg *Message) Append(args ...interface{}) {
	msg.Arguments = append(msg.Arguments, args...)
}

// Equals returns true if the given OSC Message `m` is equal to the current OSC
// Message. It checks if the OSC address and the arguments are equal. Returns
// true if the current object and `m` are equal.
func (msg *Message) Equals(m *Message) bool {
	return reflect.DeepEqual(msg, m)
}

// Clear clears the OSC address and all arguments.
func (msg *Message) Clear() {
	msg.Address = ""
	msg.ClearData()
}

// ClearData removes all arguments from the OSC Message.
func (msg *Message) ClearData() {
	msg.Arguments = msg.Arguments[len(msg.Arguments):]
}

// Match returns true, if the OSC address pattern of the OSC Message matches the given
// address. The match is case sensitive!
func (msg *Message) Match(addr string) bool {
	regexp, err := getRegEx(msg.Address)
	if err != nil {
		return false
	}
	return regexp.MatchString(addr)
}

// TypeTags returns the type tag string.
func (msg *Message) TypeTags() (string, error) {
	if msg == nil {
		return "", fmt.Errorf("TypeTags: message is nil")
	}

	if len(msg.Arguments) == 0 {
		return "", nil
	}

	tags := make([]byte, 0, len(msg.Arguments)+1)
	tags = append(tags, ',')
	for _, m := range msg.Arguments {
		s, err := GetTypeTag(m)
		if err != nil {
			return "", err
		}
		tags = append(tags, s...)
	}

	return *(*string)(unsafe.Pointer(&tags)), nil
}

var strBuf []byte

// String implements the fmt.Stringer interface.
func (msg *Message) String() string {
	if msg == nil {
		return ""
	}

	tags, _ := msg.TypeTags()

	strBuf = strBuf[:0]
	strBuf = append(strBuf, msg.Address...)
	if len(tags) == 0 {
		return string(strBuf)
	}

	strBuf = append(strBuf, ' ')
	strBuf = append(strBuf, tags...)

	for _, arg := range msg.Arguments {
		switch arg.(type) {
		case bool, int32, int64, float32, float64, string:
			strBuf = append(strBuf, fmt.Sprintf(" %v", arg)...)

		case nil:
			strBuf = append(strBuf, " Nil"...)

		case []byte:
			strBuf = append(strBuf, " blob"...)

		case Timetag:
			timeTag := arg.(Timetag)
			strBuf = append(strBuf, fmt.Sprintf(" %d", timeTag.TimeTag())...)
		}
	}

	return string(strBuf)
}

// CountArguments returns the number of arguments.
func (msg *Message) CountArguments() int {
	return len(msg.Arguments)
}

// MarshalBinary serializes the OSC message to a byte buffer. The byte buffer
// has the following format:
// 1. OSC Address Pattern
// 2. OSC Type Tag String
// 3. OSC Arguments
func (msg *Message) MarshalBinary() ([]byte, error) {
	// We can start with the OSC address and add it to the buffer
	data := new(bytes.Buffer)
	if err := msg.LightMarshalBinary(data); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func (msg *Message) LightMarshalBinary(data *bytes.Buffer) error {
	// Type tag string starts with ","
	typetags := []byte{','}

	// Process the type tags and collect all arguments
	for _, arg := range msg.Arguments {
		switch t := arg.(type) {
		default:
			return fmt.Errorf("LightMarshalBinary: unsupported type: %T", t)

		case bool:
			if t {
				typetags = append(typetags, 'T')
			} else {
				typetags = append(typetags, 'F')
			}

		case nil:
			typetags = append(typetags, 'N')

		case int32:
			typetags = append(typetags, 'i')
			if err := binary.Write(data, binary.BigEndian, t); err != nil {
				return err
			}

		case float32:
			typetags = append(typetags, 'f')
			if err := binary.Write(data, binary.BigEndian, t); err != nil {
				return err
			}

		case string:
			typetags = append(typetags, 's')
			writePaddedString(t, data)
		case []byte:
			typetags = append(typetags, 'b')
			if _, err := writeBlob(t, data); err != nil {
				return err
			}

		case int64:
			typetags = append(typetags, 'h')
			if err := binary.Write(data, binary.BigEndian, t); err != nil {
				return err
			}

		case float64:
			typetags = append(typetags, 'd')
			if err := binary.Write(data, binary.BigEndian, t); err != nil {
				return err
			}

		case Timetag:
			typetags = append(typetags, 't')
			b, err := t.MarshalBinary()
			if err != nil {
				return err
			}
			data.Write(b)
		}
	}

	if data.Len() >= len(initBuf) {
		return fmt.Errorf("LightMarshalBinary: payload too large: %d", data.Len())
	}

	b := initBuf[:data.Len()]
	data.Read(b)

	data.Reset()
	writePaddedString(msg.Address, data)

	// Write the type tag string to the data buffer
	writePaddedString(*(*string)(unsafe.Pointer(&typetags)), data)

	// Write the payload (OSC arguments) to the data buffer
	data.Write(b)

	if data.Len() >= len(initBuf) {
		return fmt.Errorf("LightMarshalBinary: packet too large: %d", data.Len())
	}

	return nil
}

////
// Bundle
////

// NewBundle returns an OSC Bundle. Use this function to create a new OSC
// Bundle.
func NewBundle(time time.Time) *Bundle {
	return &Bundle{Timetag: *NewTimetag(time)}
}

// Append appends an OSC bundle or OSC message to the bundle.
func (b *Bundle) Append(pck Packet) error {
	switch t := pck.(type) {
	default:
		return fmt.Errorf("unsupported OSC packet type: only Bundle and Message are supported")

	case *Bundle:
		b.Bundles = append(b.Bundles, t)

	case *Message:
		b.Messages = append(b.Messages, t)
	}

	return nil
}

// MarshalBinary serializes the OSC bundle to a byte array with the following
// format:
// 1. Bundle string: '#bundle'
// 2. OSC timetag
// 3. Length of first OSC bundle element
// 4. First bundle element
// 5. Length of n OSC bundle element
// 6. n bundle element
func (b *Bundle) MarshalBinary() ([]byte, error) {
	// Add the '#bundle' string
	buf := new(bytes.Buffer)
	if err := b.LightMarshalBinary(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Bundle) LightMarshalBinary(data *bytes.Buffer) error {
	writePaddedString("#bundle", data)

	// Add the time tag
	if err := b.Timetag.LightMarshalBinary(data); err != nil {
		return err
	}

	// Process all OSC Messages
	for _, m := range b.Messages {
		buf, err := m.MarshalBinary()
		if err != nil {
			return err
		}

		// Append the length of the OSC message
		if err = binary.Write(data, binary.BigEndian, int32(len(buf))); err != nil {
			return err
		}

		// Append the OSC message
		data.Write(buf)
	}

	// Process all OSC Bundles
	for _, b := range b.Bundles {
		buf, err := b.MarshalBinary()
		if err != nil {
			return err
		}

		// Write the size of the bundle
		if err = binary.Write(data, binary.BigEndian, int32(len(buf))); err != nil {
			return err
		}

		// Append the bundle
		data.Write(buf)
	}

	return nil
}
