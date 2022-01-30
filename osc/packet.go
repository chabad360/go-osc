package osc

import (
	"encoding"
	"fmt"
)

// Packet is the interface for Message and Bundle.
type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// ParsePacket parses an OSC packet.
func ParsePacket(d []byte) (Packet, error) {
	data := make([]byte, len(d))
	copy(data, d)

	return parsePacket(data)
}

// parsePacket assumes that the bytes have already been copied.
func parsePacket(data []byte) (Packet, error) {
	if !(len(data) > 0) {
		return nil, fmt.Errorf("parsePacket: invalid packet")
	}

	switch data[0] {
	case '/': // An OSC Message starts with a '/'
		return newMessageFromData(data)
	case '#': // An OSC bundle starts with a '#'
		return newBundleFromData(data)
	default:
		return nil, fmt.Errorf("parsePacket: invalid packet")
	}
}
