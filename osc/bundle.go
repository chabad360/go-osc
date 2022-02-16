package osc

import (
	"encoding/binary"
	"fmt"
	"time"
)

const (
	bundleTagString = "#bundle"
)

// Bundle represents an OSC Bundle.
type Bundle struct {
	Timetag  Timetag
	Elements []Packet
}

// Verify that Bundle implements the Packet interface.
var _ Packet = (*Bundle)(nil)

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *Bundle) MarshalBinary() ([]byte, error) {
	buf := bufPool.Get().(*[]byte)
	defer bufPool.Put(buf)
	copy(*buf, empty[:])

	// Add the '#bundle' string
	n := writePaddedString("#bundle", *buf)

	// Add the time tag
	binary.BigEndian.PutUint64((*buf)[n:n+bit64Size], uint64(b.Timetag))
	n += bit64Size

	// Process all Bundle elements
	for _, m := range b.Elements {
		bb, err := m.MarshalBinary()
		if err != nil {
			return nil, err
		}

		// Write the size of the element
		binary.BigEndian.PutUint32((*buf)[n:n+bit32Size], uint32(len(bb)))
		n += bit32Size

		// Append the bundle
		n += copy((*buf)[n:], bb)
	}

	bb := make([]byte, n)
	copy(bb, *buf)

	return bb, nil
}

// NewBundleFromData returns a new OSC bundle created from the parsed data.
func NewBundleFromData(data []byte) (b *Bundle, err error) {
	b = &Bundle{}
	if err = b.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return b, nil
}

// newBundleFromData assumes that the bytes have already been copied.
func newBundleFromData(data []byte) (b *Bundle, err error) {
	b = &Bundle{}
	if err = b.unmarshalBinary(data); err != nil {
		return nil, err
	}
	return b, nil
}

// NewBundleWithTime returns an OSC Bundle. Use this function to create a new OSC Bundle.
func NewBundleWithTime(time time.Time, elems ...Packet) *Bundle {
	return &Bundle{Timetag: NewTimetagFromTime(time), Elements: elems}
}

// NewBundle returns a new Bundle with elems as the elements.
func NewBundle(elems ...Packet) *Bundle {
	return &Bundle{Timetag: NewTimetag(), Elements: elems}
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
	data := make([]byte, len(d))
	copy(data, d)

	return b.unmarshalBinary(data)
}

// unmarshalBinary is the actual implementation, it doesn't copy, so we can use a single copy for bundles.
func (b *Bundle) unmarshalBinary(data []byte) error {
	if (len(data) % bit32Size) != 0 {
		return fmt.Errorf("UnmarshalBinary: data isn't padded properly")
	}

	if len(data) < 16 {
		return fmt.Errorf("UnmarshalBinary: bundle is too short")
	}

	// Read the '#bundle' OSC string
	startTag, n, err := parsePaddedString(data)
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
		data = data[bit32Size:]
		if len(data) < length {
			return fmt.Errorf("UnmarshalBinary: invalid bundle element length %d: remaining data in packet: %d", length, len(data))
		}

		p, err := parsePacket(data[:length])
		if err != nil {
			return err
		}
		data = data[length:]
		b.Append(p)
	}

	return nil
}
