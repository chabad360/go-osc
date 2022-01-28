package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	bundleTagString = "#bundle"
)

// Bundle represents an OSC bundle. It consists of the OSC-string "#bundle"
// followed by an OSC Time Tag, followed by zero or more OSC bundle/message
// elements. The OSC-timetag is a 64-bit fixed point time tag. See
// http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag  Timetag
	Elements []Packet
}

// Verify that Bundle implements the Packet interface.
var _ Packet = (*Bundle)(nil)

// MarshalBinary implements the encoding.BinaryMarshaler
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

// LightMarshalBinary allows you to marshal a bundle into a bytes.Buffer to avoid an allocation.
func (b *Bundle) LightMarshalBinary(data *bytes.Buffer) error {
	writePaddedString("#bundle", data)

	// Add the time tag
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(b.Timetag))
	data.Write(buf)

	// Process all Bundle elements
	for _, m := range b.Elements {
		buf, err := m.MarshalBinary()
		if err != nil {
			return err
		}

		// Write the size of the bundle
		b := make([]byte, bit32Size)
		binary.BigEndian.PutUint32(b, uint32(len(buf)))
		data.Write(b)

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
	return &Bundle{Timetag: NewTimetagFromTime(time)}
}

// Append appends an OSC bundle or OSC message to the bundle.
func (b *Bundle) Append(pck Packet) error {
	switch t := pck.(type) {
	default:
		return fmt.Errorf("unsupported OSC packet type: only Bundle and Message are supported")

	case *Bundle, *Message:
		b.Elements = append(b.Elements, t)
	}

	return nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *Bundle) UnmarshalBinary(d []byte) error {
	if (len(d) % bit32Size) != 0 {
		return fmt.Errorf("UnmarshalBinary: data isn't padded properly")
	}

	if len(d) < 20 {
		return fmt.Errorf("UnmarshalBinary: bundle is too short")
	}

	data := make([]byte, len(d))
	copy(data, d)

	// Read the '#bundle' OSC string
	startTag, n, err := readPaddedString(data)
	if err != nil {
		return err
	}
	data = data[n:]

	if startTag != bundleTagString {
		return fmt.Errorf("invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	// Create a new bundle
	b.Timetag = Timetag(binary.BigEndian.Uint64(data[:bit64Size]))
	data = data[bit64Size:]

	// Read until the end of the buffer
	for len(data) > 0 {
		// Read the size of the bundle element
		length := int(binary.BigEndian.Uint32(data[:bit32Size]))
		if len(data) < length {
			return fmt.Errorf("invalid bundle element length: %d", length)
		}
		data = data[bit32Size:]

		p, err := ReadPacket(data[:length])
		if err != nil {
			return err
		}
		data = data[length:]
		b.Append(p)
	}

	return nil
}
