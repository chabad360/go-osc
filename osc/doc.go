// Copyright 2013 - 2015 Sebastian Ruml <sebastian.ruml@gmail.com>
// Copyright 2021 - 2022 Mendel Greenberg <mendel@chabad360.me>

//Package osc provides a client and server for sending and receiving OpenSoundControl messages.
//
//This implementation is based on the Open Sound Control 1.0 Specification (http://opensoundcontrol.org/spec-1_0.html).
//
//Open Sound Control (OSC) is an open, transport-independent, message-based protocol developed for communication among computers,
//sound synthesizers, and other multimedia devices.
//
//Features
//
//- Supports OSC messages with the following TypeTags:
//
//	'i' (int32)
//	'f' (float32)
//	's' (string)
//	'b' ([]byte)
//	't' (TimeTag)
//	'h' (int64)
//	'd' (float64)
//	'T' (true)
//	'F' (false)
//	'N' (nil)
//
//- Supports OSC bundles, including TimeTags
//
//- Full support for OSC Address matching and dispatching.
//
//Packets
//
//The unit of transmission of OSC is an OSC Packet. Any application that sends OSC Packets is an OSC Client;
//any application that receives OSC Packets is an OSC Server.
//
//An OSC packet consists of its contents, a contiguous block of binary data.
//The size of an OSC packet is always 32-bit aligned.
//
//OSC packets come in two flavors:
//
//OSC Messages: An OSC message consists of an OSC address pattern and  zero or more OSC arguments.
//
//OSC Bundles: An OSC Bundle consists of an OSC Timetag, followed by zero or more OSC bundle elements.
//Each bundle element can be another OSC bundle (note this recursive definition: a bundle may contain bundles) or OSC message.
//
//Usage
//
//OSC client example:
//  client := osc.NewClient("localhost", 8765)
//  msg := osc.NewMessage("/osc/address")
//  msg.Append(int32(111))
//  msg.Append(true)
//  msg.Append("hello")
//  client.Send(msg)
//
//OSC server example:
//  d := osc.NewStandardDispatcher()
//  d.AddMsgMethod("/message/address", func(msg *osc.Message) {
//      osc.PrintMessage(msg)
//  })
//
//  server := &osc.Server{
//      Addr: "127.0.0.1:8765",
//      Dispatcher: d,
//  }
//  server.ListenAndServe()
package osc
