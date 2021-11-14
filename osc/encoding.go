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

var padBytes = []byte{0, 0, 0, 0}

// readBlob reads an OSC blob from the blob byte array. Padding bytes are
// removed from the reader and not returned.
func readBlob(reader *bytes.Buffer) ([]byte, int, error) {
	// First, get the length
	var blobLen int64
	var err error
	if blobLen, err = binary.ReadVarint(reader); err != nil {
		return nil, 0, err
	}
	n := 4 + int(blobLen)

	if blobLen < 1 || blobLen > int64(reader.Len()) {
		return nil, 0, fmt.Errorf("readBlob: invalid blob length %d", blobLen)
	}

	// Read the data
	blob := make([]byte, blobLen)
	if _, err := reader.Read(blob); err != nil {
		return nil, 0, err
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(int(blobLen))
	reader.Next(numPadBytes)

	b := blob

	return b, n + numPadBytes, nil
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length
// of data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, buf *bytes.Buffer) (int, error) {
	if len(data) > len(initBuf)-4 {
		return 0, fmt.Errorf("writeBlob: blob length greater than 65,531 bytes")
	}

	// Add the size of the blob
	//dlen := int32(len(data))
	if err := binary.Write(buf, binary.BigEndian, int32(len(data))); err != nil {
		return 0, err
	}

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
	p := bytes.IndexByte(reader.Bytes(), 0)
	if p < 0 {
		return "", 0, io.EOF
	}

	//Read the string from the reader
	//str, err := reader.ReadString(0)
	//if err != nil {
	//	return "", 0, err
	//}

	str := make([]byte, p)
	reader.Read(str)

	n := len(str)

	//str = str[:n-1]

	// Remove the padding bytes
	reader.Next(padBytesNeeded(n+1) + 1)
	n += padBytesNeeded(n+1) + 1

	return *(*string)(unsafe.Pointer(&str)), n, nil
}

// writePaddedString writes a string with padding bytes to the buffer.
// Returns, the number of written bytes and an error if any.
func writePaddedString(str string, buf *bytes.Buffer) int {
	// Write the string to the buffer
	n, _ := buf.WriteString(str)
	// Write the null terminator to the buffer as well
	buf.WriteByte(0)
	n += 1

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
