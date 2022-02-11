package osc

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// Method is an interface for OSC Methods.
type Method interface {
	HandleMessage(msg *Message)
}

// MethodFunc implements the Method interface. Type definition for an OSC Method function.
type MethodFunc func(msg *Message)

// HandleMessage calls itself with the given OSC Message. Implements the Method interface.
func (f MethodFunc) HandleMessage(msg *Message) {
	f(msg)
}

// Dispatcher handles the dispatching of received OSC Packets to Methods for their given Address.
type Dispatcher struct {
	methods map[string]Method
}

// AddMethod adds a new OSC Method for the given OSC Address.
func (d *Dispatcher) AddMethod(addr string, method Method) error {
	if d.methods == nil {
		d.methods = make(map[string]Method)
	}

	if strings.ContainsAny(addr, "*?,[]{}# ") {
		return fmt.Errorf("AddMsgMethod: OSC Method may not contain any characters in \"*?,[]{}# \"")
	}

	if _, ok := d.methods[addr]; ok {
		return fmt.Errorf("AddMsgMethod: OSC Method exists already")
	}

	d.methods[addr] = method
	return nil
}

// AddMethodFunc allows you to just pass a MethodFunc.
func (d *Dispatcher) AddMethodFunc(addr string, method MethodFunc) error {
	return d.AddMethod(addr, method)
}

// Dispatch dispatches OSC Packets.
func (d *Dispatcher) Dispatch(packet Packet, a net.Addr) {
	switch p := packet.(type) {
	default:
		panic(fmt.Errorf("dispatch: invalid Packet: %v", p))

	case *Message:
		r, err := getRegEx(p.Address)
		if err != nil {
			panic(fmt.Errorf("dispatch: invalid Packet: %v: %w", p, err))
		}
		// The OSC Spec mentions that each address is divided into parts, so we could use a radix tree here.
		// For now, I'm gonna hope that being clever is enough
		r.Longest()
		aParts := len(strings.Split(p.Address, "/"))
		for addr, method := range d.methods {
			if aParts == len(strings.Split(addr, "/")) && r.FindString(addr) == addr {
				// This is going to stay blocking until I can figure out how to deal with what appears to be a race condition
				//go func() {
				//	defer recoverer(a)
				method.HandleMessage(p)
				//}()
			}
		}
	case *Bundle:
		time.AfterFunc(p.Timetag.ExpiresIn(), func() {
			defer recoverer(a)
			for _, elem := range p.Elements {
				d.Dispatch(elem, a)
			}
		})
	}
}

// getRegEx returns a regexp.Regexp for the given address.
func getRegEx(pattern string) (*regexp.Regexp, error) {
	r := strings.NewReplacer(
		".", `\.`,
		"(", `\(`,
		")", `\)`,
		"*", "[^/]*",
		"{", "(",
		",", "|",
		"}", ")",
		"?", "[^/]",
		"!", "^",
	)
	pattern = r.Replace(pattern)

	return regexp.Compile(pattern)
}
