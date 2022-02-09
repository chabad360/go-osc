package main

import (
	"fmt"
	"github.com/chabad360/go-osc/osc"
)

func main() {
	d := &osc.Dispatcher{}
	d.AddMethodFunc("/message/address", func(msg *osc.Message) {
		fmt.Println(msg.Address, msg.Arguments)
	})
	osc.ListenAndServe("127.0.0.1:8765", d.Dispatch)
}
