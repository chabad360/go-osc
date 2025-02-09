package osc

import (
	"bytes"
	"fmt"
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

// GetTypeTag returns the OSC type tag for the given argument.
func GetTypeTag(arg interface{}) (string, error) {
	switch t := arg.(type) {
	case bool:
		if t {
			return "T", nil
		}
		return "F", nil
	case nil:
		return "N", nil
	case int32:
		return "i", nil
	case float32:
		return "f", nil
	case string:
		return "s", nil
	case []byte:
		return "b", nil
	case int64:
		return "h", nil
	case float64:
		return "d", nil
	case Timetag:
		return "t", nil
	default:
		return "", fmt.Errorf("unsupported type: %T", t)
	}
}
