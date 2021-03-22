package osc

import (
	"bytes"
	"io"
	"testing"
)

func TestReadPaddedString(t *testing.T) {
	for _, tt := range []struct {
		buf   []byte // buffer
		want  int    // bytes needed
		want1 string // resulting string
		err   error
	}{
		{[]byte{'t', 'e', 's', 't', 's', 't', 'r', 'i', 'n', 'g', 0, 0}, 12, "teststring", nil},
		{[]byte{'t', 'e', 's', 't', 'e', 'r', 's', 0}, 8, "testers", nil},
		{[]byte{'t', 'e', 's', 't', 's', 0, 0, 0}, 8, "tests", nil},
		{[]byte{'t', 'e', 's', 0, 0, 0, 0, 0}, 4, "tes", nil}, // OSC uses null terminated strings
		{[]byte{'t', 'e', 's', 't'}, 0, "", io.EOF},           // if there is no null byte at the end, it doesn't work.

	} {
		got, got1, err := readPaddedString(bytes.NewBuffer(tt.buf))
		if err != tt.err {
			t.Errorf("%s: Error reading padded string: %s", tt.want1, err)
		}
		if got1 != tt.want {
			t.Errorf("%s: Bytes needed don't match; got = %d, want = %d", tt.want1, got1, tt.want)
		}
		if got != tt.want1 {
			t.Errorf("%s: Strings don't match; got = %b, want = %b", tt.want1, []byte(got), []byte(tt.want1))
		}
	}
}

func TestWritePaddedString(t *testing.T) {
	buf := []byte{}
	bytesBuffer := bytes.NewBuffer(buf)
	testString := "testString"
	expectedNumberOfWrittenBytes := len(testString) + padBytesNeeded(len(testString))

	if n := writePaddedString(testString, bytesBuffer); n != expectedNumberOfWrittenBytes {
		t.Errorf("Expected number of written bytes should be \"%d\" and is \"%d\"", expectedNumberOfWrittenBytes, n)
	}
}

func TestPadBytesNeeded(t *testing.T) {
	var n int
	n = padBytesNeeded(4)
	if n != 0 {
		t.Errorf("Number of pad bytes should be 0 and is: %d", n)
	}

	n = padBytesNeeded(3)
	if n != 1 {
		t.Errorf("Number of pad bytes should be 1 and is: %d", n)
	}

	n = padBytesNeeded(1)
	if n != 3 {
		t.Errorf("Number of pad bytes should be 3 and is: %d", n)
	}

	n = padBytesNeeded(0)
	if n != 0 {
		t.Errorf("Number of pad bytes should be 0 and is: %d", n)
	}

	n = padBytesNeeded(32)
	if n != 0 {
		t.Errorf("Number of pad bytes should be 0 and is: %d", n)
	}

	n = padBytesNeeded(63)
	if n != 1 {
		t.Errorf("Number of pad bytes should be 1 and is: %d", n)
	}

	n = padBytesNeeded(10)
	if n != 2 {
		t.Errorf("Number of pad bytes should be 2 and is: %d", n)
	}
}
