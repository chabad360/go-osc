package main

import (
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"net"
)

func main() {
	addr := "127.0.0.1:7001"

	fmt.Println("### Welcome to go-osc server demo")

	fmt.Println("Start listening on", addr)
	fmt.Println(osc.ListenAndServe(addr, handle))
}

func handle(packet osc.Packet, addr net.Addr) {
	switch p := packet.(type) {
	default:
		fmt.Println("Unknown packet type!")

	case *osc.Message:
		fmt.Printf("-- OSC Message (%s):\tAddress: \"%s\"\tArguments: %v\n", addr, p.Address, p.Arguments)

	case *osc.Bundle:
		fmt.Printf("-- OSC Bundle (%s):\tTimeTag: %v\tElements: %d\n", addr, p.Timetag.Time(), len(p.Elements))
		for _, p := range p.Elements {
			fmt.Printf("\t")
			handle(p, addr)
		}
	}
}
