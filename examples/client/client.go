package main

import (
	"bufio"
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"io"
	"os"
	"strings"
)

// TODO: Revise the client!
func main() {
	addr := "localhost:8765"
	client, err := osc.Dial(addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("### Welcome to go-osc transmitter demo")
	fmt.Println("Please, select the OSC packet type you would like to send:")
	fmt.Println("\tm: OSCMessage")
	fmt.Println("\tb: OSCBundle")
	fmt.Println("\tPress \"q\" to exit")
	fmt.Printf("# ")

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		sline := strings.TrimRight(string(line), "\n")
		if sline == "m" {
			message := osc.NewMessage("/message/address", int32(12345), "teststring", true, false)
			fmt.Println(client.Send(message))
		} else if sline == "b" {
			message1 := osc.NewMessage("/bundle/message/1", int32(12345), "teststring", true, false)
			message2 := osc.NewMessage("/bundle/message/2", int32(3344), float32(101.9), "string1", "string2", true)

			bundle := osc.NewBundle(message1, message2)
			fmt.Println(client.Send(bundle))
		} else if sline == "q" {
			fmt.Println("Exit!")
			os.Exit(0)
		}

		fmt.Printf("# ")
	}
}
