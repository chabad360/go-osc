# GoOSC
[![GoDoc](https://godoc.org/github.com/chabad360/go-osc?status.svg)](https://godoc.org/github.com/chabad360/go-osc)
[![Go](https://github.com/chabad360/go-osc/actions/workflows/go.yml/badge.svg)](https://github.com/chabad360/go-osc/actions/workflows/go.yml)

[Open Sound Control (OSC)](http://opensoundcontrol.org/introduction-osc) library for Golang. Implemented in pure Go.

## Features

- OSC Bundles
- OSC Messages
- OSC Client
- OSC Server
- Support for OSC address pattern matching
- Supports the following OSC argument types:
  - `i` (Int32)
  - `f` (Float32)
  - `s` (string)
  - `b` (blob / binary data)
  - `h` (Int64)
  - `t` (OSC timetag)
  - `d` (Double/int64)
  - `T` (True)
  - `F` (False)
  - `N` (Nil)
  
## Usage

### Client

```go
package main

import "github.com/chabad360/go-osc/osc"

func main() {
    client, _ := osc.Dial("localhost:8765")
    msg := osc.NewMessage("/osc/address", int32(111), true, "hello")
    client.Send(msg)
}
```

### Server

```go
package main

import (
    "fmt"
    "github.com/chabad360/go-osc/osc"
)

func main() {
    addr := "127.0.0.1:8765"
    d := &osc.Dispatcher{}
    d.AddMethod("/message/address", func(msg *osc.Message) {
        fmt.Println(msg)
    })

    server := &osc.Server{
        Addr:       addr,
        Handler:    d.Dispatch,
    }
    server.ListenAndServe()
}
```
