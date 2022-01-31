package osc

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
)

////
// Utility and helper functions
////
var (
	bufPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, MaxPacketSize))
		},
	}
	bPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, MaxPacketSize)
			return &b
		},
	}
	empty = [MaxPacketSize]byte{}
)

// addressExists returns true if the OSC address `addr` is found in `handlers`.
func addressExists(addr string, handlers map[string]Handler) bool {
	for h := range handlers {
		if h == addr {
			return true
		}
	}
	return false
}

// getRegEx compiles and returns a regular expression object for the given
// address `pattern`.
func getRegEx(pattern string) (*regexp.Regexp, error) {
	for _, trs := range []struct {
		old, new string
	}{
		{".", `\.`}, // Escape all '.' in the pattern
		{"(", `\(`}, // Escape all '(' in the pattern
		{")", `\)`}, // Escape all ')' in the pattern
		{"*", ".*"}, // Replace a '*' with '.*' that matches zero or more chars
		{"{", "("},  // Change a '{' to '('
		{",", "|"},  // Change a ',' to '|'
		{"}", ")"},  // Change a '}' to ')'
		{"?", "."},  // Change a '?' to '.'
	} {
		pattern = strings.ReplaceAll(pattern, trs.old, trs.new)
	}

	return regexp.Compile(pattern)
}
