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

// ReadPacket parses an OSC packet.
func ReadPacket(data []byte) (Packet, error) {
	if !(len(data) > 0) {
		return nil, fmt.Errorf("ReadPacket: invalid packet")
	}

	switch data[0] {
	case '/': // An OSC Message starts with a '/'
		return NewMessageFromData(data)
	case '#': // An OSC bundle starts with a '#'
		return NewBundleFromData(data)
	default:
		return nil, fmt.Errorf("ReadPacket: invalid packet")
	}
}
