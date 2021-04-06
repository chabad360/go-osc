package osc

import (
	"bytes"
	"encoding/binary"
)

////
// De/Encoding functions
////

var padBytes = make([]byte, 4)

// readBlob reads an OSC blob from the blob byte array. Padding bytes are
// removed from the reader and not returned.
func readBlob(reader *bytes.Buffer) ([]byte, int, error) {
	// First, get the length
	var blobLen int32
	if err := binary.Read(reader, binary.BigEndian, &blobLen); err != nil {
		return nil, 0, err
	}
	n := 4 + int(blobLen)

	// Read the data
	blob := make([]byte, blobLen)
	if _, err := reader.Read(blob); err != nil {
		return nil, 0, err
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(int(blobLen))
	reader.Next(numPadBytes)

	return blob, n + numPadBytes, nil
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length
// of data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, buf *bytes.Buffer) (int, error) {
	// Add the size of the blob
	dlen := int32(len(data))
	if err := binary.Write(buf, binary.BigEndian, dlen); err != nil {
		return 0, err
	}

	// Write the data
	n, _ := buf.Write(data)

	// Add padding bytes if necessary
	numPadBytes := padBytesNeeded(n)
	buf.Write(padBytes[:numPadBytes])

	return 4 + n + numPadBytes, nil
}

var str []byte
var err error

// readPaddedString reads a padded string from the given reader. The padding
// bytes are removed from the reader.
func readPaddedString(reader *bytes.Buffer) (string, int, error) {
	//Read the string from the reader
	str, err := reader.ReadString(0)
	if err != nil {
		return "", 0, err
	}

	n := len(str)

	//str = str[:reader.Len()]
	//_, err := reader.Read(str)
	//if err != nil {
	//	return "", 0, err
	//}
	//
	//n := bytes.IndexByte(str, 0)
	//if n < 0 {
	//	return "", 0, io.EOF
	//}
	//n++
	//reader.Reset()
	//reader.Write(str[n:])

	str = str[:n-1]

	// Remove the padding bytes
	reader.Next(padBytesNeeded(n))
	n += padBytesNeeded(n)

	return str, n, nil
	//return *(*string)(unsafe.Pointer(&str)), n, nil
}

// writePaddedString writes a string with padding bytes to the a buffer.
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
