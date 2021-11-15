package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

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

// MarshalBinary serializes the OSC bundle to a byte array with the following
// format:
// 1. Bundle string: '#bundle'
// 2. OSC timetag
// 3. Length of first OSC bundle element
// 4. First bundle element
// 5. Length of n OSC bundle element
// 6. n bundle element
func (b *Bundle) MarshalBinary() (bb []byte, err error) {
	// Add the '#bundle' string
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()

	if err = b.LightMarshalBinary(buf); err != nil {
		return nil, err
	}
	return append(bb, buf.Bytes()...), nil
}

// LightMarshalBinary allows you to marshal a bundle into a bytes.Buffer to avoid an allocation
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

// NewBundleFromData returns a new OSC bundle created from the parsed data.
func NewBundleFromData(data []byte) (b *Bundle, err error) {
	b = &Bundle{}
	if err = b.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return b, nil
}

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

// UnmarshalBinary implements the BinaryUnmarshaler interface.
func (b *Bundle) UnmarshalBinary(data []byte) error {
	reader := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(reader)
	reader.Reset()

	reader.Write(data)

	// Read the '#bundle' OSC string
	startTag, _, err := readPaddedString(reader)
	if err != nil {
		return err
	}

	if startTag != bundleTagString {
		return fmt.Errorf("invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	var timeTag uint64
	if err = binary.Read(reader, binary.BigEndian, &timeTag); err != nil {
		return err
	}

	// Create a new bundle
	b.Timetag = *NewTimetagFromTimetag(timeTag)

	// Read until the end of the buffer
	for reader.Len() > 0 {
		// Read the size of the bundle element
		var length int32
		if err = binary.Read(reader, binary.BigEndian, &length); err != nil {
			return err
		}

		var p Packet
		p, err = ReadPacket(reader.Bytes())
		if err != nil {
			return err
		}
		b.Append(p)
	}

	return nil
}
