package main

import (
	"bufio"
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"net"
	"os"
)

func main() {
	addr := "127.0.0.1:8765"

	fmt.Println("### Welcome to go-osc receiver demo")
	fmt.Println("Press \"q\" to exit")

	go func() {
		fmt.Println("Start listening on", addr)
		osc.ListenAndServe(addr, func(packet osc.Packet, addr net.Addr) {
			fmt.Printf("Recived packet from %s\n", addr)
			switch p := packet.(type) {
			default:
				fmt.Println("Unknown packet type!")

			case *osc.Message:
				fmt.Printf("-- OSC Message:\n%v", p)

			case *osc.Bundle:
				fmt.Printf("-- OSC Bundle:\n%v", p)
			}
		})
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		c, err := reader.ReadByte()
		if err != nil {
			os.Exit(0)
		}

		if c == 'q' {
			os.Exit(0)
		}
	}
}
