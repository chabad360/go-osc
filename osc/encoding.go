package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

////
// De/Encoding functions
////

const (
	MaxPacketSize int = 65535
	bit32Size     int = 4
	bit64Size     int = 8
)

var padBytes = []byte{0, 0, 0, 0}

// readBlob reads an OSC blob from the blob byte array. Padding bytes are
// removed from the reader and not returned.
func readBlob(reader *bytes.Buffer) ([]byte, int, error) {
	// First, get the length
	blobLen := int(binary.BigEndian.Uint32(reader.Next(4)))
	n := 4 + blobLen

	if blobLen < 1 || blobLen > reader.Len() {
		return nil, 0, fmt.Errorf("readBlob: invalid blob length %d", blobLen)
	}

	// Read the data
	blob := make([]byte, blobLen)
	if _, err := reader.Read(blob); err != nil {
		return nil, 0, fmt.Errorf("readBlob: %w", err)
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(blobLen)
	reader.Next(numPadBytes)

	b := blob

	return b, n + numPadBytes, nil
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length
// of data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, buf *bytes.Buffer) (int, error) {
	if len(data) > MaxPacketSize-4 {
		return 0, fmt.Errorf("writeBlob: blob length greater than 65,531 bytes")
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

// readPaddedString reads a padded string from the given reader. The padding
// bytes are removed from the reader.
func readPaddedString(reader *bytes.Buffer) (string, int, error) {
	str, err := reader.ReadString(0)
	if err != nil {
		return "", 0, err
	}

	if str[0] == 0 {
		return "", 0, fmt.Errorf("readPaddedString: empty string")
	}

	n := len(str)
	str = str[:n-1]

	// Remove the padding bytes
	n += len(reader.Next(padBytesNeeded(n)))

	return str, n, nil
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
