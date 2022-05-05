package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

////
// De/Encoding functions
////

type RangeError struct {
	pos int
	len int
}

func (r *RangeError) Error() string {
	return fmt.Sprintf("data length %d out of range %d: %s", r.pos, r.len, io.EOF.Error())
}

func (r *RangeError) Unwrap() error {
	return io.EOF
}

func NewRangeError(pos, len int) *RangeError {
	return &RangeError{pos, len}
}

var AlignmentError = errors.New("data isn't 32bit aligned")

// parseBlob parses an OSC blob from the blob byte array. Padding bytes are removed from the reader and not returned.
func parseBlob(data []byte) ([]byte, int, error) {
	if len(data) < bit64Size {
		return nil, 0, fmt.Errorf("parseblob: %w", NewRangeError(bit64Size, len(data)))
	}
	// First, get the length
	blobLen := int(binary.BigEndian.Uint32(data[:bit32Size]))
	n := bit32Size + blobLen
	data = data[bit32Size:]

	if blobLen < 1 || blobLen > len(data) {
		return nil, 0, fmt.Errorf("parseBlob: %w", NewRangeError(blobLen, len(data)))
	}

	return data[:blobLen], n + padBytesNeeded(n), nil
}

// writeBlob writes a byte array as an OSC blob into b.
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
	if len(data) < bit32Size {
		return "", 0, fmt.Errorf("parsePaddedString: %w", NewRangeError(bit32Size, len(data)))
	}
	pos := bytes.IndexByte(data, 0)
	if pos == -1 {
		return "", 0, fmt.Errorf("parsePaddedString: %w", io.EOF)
	}

	str := data[:pos]

	return *(*string)(unsafe.Pointer(&str)), pos + 1 + padBytesNeeded(pos+1), nil
}

// writePaddedString writes a string with padding bytes to the buffer.
func writePaddedString(str string, b []byte) int {
	// Write the string to the buffer
	n := copy(b, str)
	n++

	return n + padBytesNeeded(n)
}

// writeTypeTags writes a TypeTag string to b.
func writeTypeTags(elems []interface{}, b []byte) (int, error) {
	b[0] = ','
	n := 1
	for _, elem := range elems {
		s := ToTypeTag(elem)
		if s == TypeInvalid {
			return n, fmt.Errorf("writeTypeTags: %w", NewTypeError(elem))
		}
		b[n] = byte(s)
		n++
	}
	n++

	return n + padBytesNeeded(n), nil
}

// padBytesNeeded determines how many bytes are needed to fill up to the next 4 byte length.
func padBytesNeeded(elementLen int) int {
	return (4 - (elementLen % 4)) % 4
}
