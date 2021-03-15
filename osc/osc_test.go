package osc

const zero = string(byte(0))

// nulls returns a string of `i` nulls.
func nulls(i int) string {
	s := ""
	for j := 0; j < i; j++ {
		s += zero
	}
	return s
}

// makePacket creates a fake Message Packet.
func makePacket(addr string, args []string) Packet {
	msg := NewMessage(addr)
	for _, arg := range args {
		msg.Append(arg)
	}
	return msg
}
