package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

////
// De/Encoding functions
////

// parseBlob parses an OSC blob from the blob byte array. Padding bytes are
// removed from the reader and not returned.
func parseBlob(data []byte) ([]byte, int, error) {
	// First, get the length
	blobLen := int(binary.BigEndian.Uint32(data[:bit32Size]))
	n := bit32Size + blobLen
	data = data[bit32Size:]

	if blobLen < 1 || blobLen > len(data) {
		return nil, 0, fmt.Errorf("parseBlob: invalid blob length %d", blobLen)
	}

	return data[:blobLen], n + padBytesNeeded(n), nil
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length
// of data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, b []byte) int {
	// Add the size of the blob
	binary.BigEndian.PutUint32(b[:bit32Size], uint32(len(data)))
	n := bit32Size

	// Write the data
	n += copy(b[n:], data)

	return n + padBytesNeeded(n)
}

// parsePaddedString reads a padded string from the given slice and returns the string and the number of bytes read.
func parsePaddedString(data []byte) (string, int, error) {
	pos := bytes.IndexByte(data, 0)
	if pos == -1 {
		return "", 0, fmt.Errorf("readPaddedString: %w", io.EOF)
	}

	str := data[:pos]

	return *(*string)(unsafe.Pointer(&str)), pos + 1 + padBytesNeeded(pos+1), nil
}

// writePaddedString writes a string with padding bytes to the buffer.
// Returns, the number of written bytes and an error if any.
func writePaddedString(str string, b []byte) int {
	// Write the string to the buffer
	n := copy(b, str)
	n++

	return n + padBytesNeeded(n)
}

// writeTypeTags writes a typetag string to b.
func writeTypeTags(elems []interface{}, b []byte) (int, error) {
	b[0] = ','
	n := 1
	for _, elem := range elems {
		s := ToTypeTag(elem)
		if s == TypeInvalid {
			return n, fmt.Errorf("writeTypeTags: unsupported type: %T", elem)
		}
		b[n] = byte(s)
		n++
	}
	n++

	return n + padBytesNeeded(n), nil
}

// padBytesNeeded determines how many bytes are needed to fill up to the next 4
// byte length.
func padBytesNeeded(elementLen int) int {
	return (4 - (elementLen % 4)) % 4
}
