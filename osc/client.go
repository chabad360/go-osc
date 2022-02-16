package osc

import (
	"net"
)

// Client enables you to send OSC Packets to a specified server.
type Client struct {
	conn *net.UDPConn
}

// Dial creates a new OSC Client with a connection to the specified server.
func Dial(addr string) (*Client, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, a)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

// Send sends an OSC Packet to the server.
func (c *Client) Send(packet Packet) error {
	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = c.conn.Write(data)
	return err
}

// Close closes the connection to the server.
func (c *Client) Close() error {
	return c.conn.Close()
}
