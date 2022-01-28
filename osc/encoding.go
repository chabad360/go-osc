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

const (
	MaxPacketSize int = 65507
	bit32Size     int = 4
	bit64Size     int = 8
)

var padBytes = []byte{0, 0, 0, 0}

// readBlob reads an OSC blob from the blob byte array. Padding bytes are
// removed from the reader and not returned.
func readBlob(data []byte) ([]byte, int, error) {
	// First, get the length
	blobLen := int(binary.BigEndian.Uint32(data[:bit32Size]))
	n := bit32Size + blobLen
	data = data[bit32Size:]

	if blobLen < 1 || blobLen > len(data) {
		return nil, 0, fmt.Errorf("readBlob: invalid blob length %d", blobLen)
	}

	return data[:blobLen], n + padBytesNeeded(n), nil
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length
// of data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, buf *bytes.Buffer) (int, error) {
	if len(data) > MaxPacketSize-4 { // TODO: properly compute packet size
		return 0, fmt.Errorf("writeBlob: blob length greater than 65,503 bytes")
	}

	// Add the size of the blob
	b := make([]byte, bit32Size)
	binary.BigEndian.PutUint32(b, uint32(len(data)))
	buf.Write(b)

	// Write the data
	n, _ := buf.Write(data)

	// Add padding bytes if necessary
	numPadBytes := padBytesNeeded(n)
	buf.Write(padBytes[:numPadBytes])

	return 4 + n + numPadBytes, nil
}

// readPaddedString reads a padded string from the given slice and returns the string and the number of bytes read.
func readPaddedString(data []byte) (string, int, error) {
	pos := bytes.IndexByte(data, 0)
	if pos == -1 {
		return "", 0, fmt.Errorf("readPaddedString: %w", io.EOF)
	}

	str := data[:pos]

	return *(*string)(unsafe.Pointer(&str)), pos + 1 + padBytesNeeded(pos+1), nil
}

// writePaddedString writes a string with padding bytes to the buffer.
// Returns, the number of written bytes and an error if any.
func writePaddedString(str string, buf *bytes.Buffer) int {
	// Write the string to the buffer
	n, _ := buf.WriteString(str)
	// Write the null terminator to the buffer as well
	buf.WriteByte(0)
	n++

	// Calculate the padding bytes needed and create a buffer for the padding bytes
	numPadBytes := padBytesNeeded(n)
	// Add the padding bytes to the buffer
	buf.Write(padBytes[:numPadBytes])

	return n + numPadBytes
}

// padBytesNeeded determines how many bytes are needed to fill up to the next 4
// byte length.
func padBytesNeeded(elementLen int) int {
	return (4 - (elementLen % 4)) % 4
}
