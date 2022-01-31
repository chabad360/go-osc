package osc

import (
	"reflect"
	"testing"
)

func TestMessage_Append(t *testing.T) {
	oscAddress := "/address"
	message := NewMessage(oscAddress)
	if message.Address != oscAddress {
		t.Errorf("OSC address should be \"%s\" and is \"%s\"", oscAddress, message.Address)
	}

	message.Append("string argument")
	message.Append(int32(123456789))
	message.Append(true)

	if len(message.Arguments) != 3 {
		t.Errorf("Number of arguments should be %d and is %d", 3, len(message.Arguments))
	}
}

func TestMessage_Equals(t *testing.T) {
	msg1 := NewMessage("/address")
	msg2 := NewMessage("/address")
	msg1.Append(int32(1234))
	msg2.Append(int32(1234))
	msg1.Append("test string")
	msg2.Append("test string")

	if !reflect.DeepEqual(msg1, msg2) {
		t.Error("Messages should be equal")
	}
}

func TestOscMessageMatch(t *testing.T) {
	tc := []struct {
		desc        string
		addr        string
		addrPattern string
		want        bool
	}{
		{
			"match everything",
			"*",
			"/a/b",
			true,
		},
		{
			"don't match",
			"/a/b",
			"/a",
			false,
		},
		{
			"match alternatives",
			"/a/{foo,bar}",
			"/a/foo",
			true,
		},
		{
			"don't match if address is not part of the alternatives",
			"/a/{foo,bar}",
			"/a/bob",
			false,
		},
	}

	for _, tt := range tc {
		msg := NewMessage(tt.addr)

		got := msg.Match(tt.addrPattern)
		if got != tt.want {
			t.Errorf("%s: msg.Match('%s') = '%t', want = '%t'", tt.desc, tt.addrPattern, got, tt.want)
		}
	}
}

var result interface{}

func BenchmarkMessageMarshalBinary(b *testing.B) {
	var buf []byte
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		buf, _ = temp.MarshalBinary()
	}
	result = buf
}

func TestMessage_MarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		message *Message
		want    []byte
		wantErr bool
	}{
		{"sample1", sample1Msg, sample1, false},
		{"sample2", sample2Msg, sample2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.message.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalBinary() got = %s, want %s", got, tt.want)
			}
		})
	}
}

// https://opensoundcontrol.stanford.edu/spec-1_0-examples.html#osc-message-examples
// /oscillator/4/frequency ,f 440.0
var sample1 = []byte{47, 111, 115, 99, 105, 108, 108, 97, 116, 111, 114, 47, 52, 47, 102, 114, 101, 113, 117, 101, 110, 99, 121, 0, 44, 102, 0, 0, 67, 220, 0, 0}
var sample1Msg = NewMessage("/oscillator/4/frequency", float32(440.0))

// /foo ,iisff 1000 -1 hello 1.234 5.678
var sample2 = []byte{47, 102, 111, 111, 0, 0, 0, 0, 44, 105, 105, 115, 102, 102, 0, 0, 0, 0, 3, 232, 255, 255, 255, 255, 104, 101, 108, 108, 111, 0, 0, 0, 63, 157, 243, 182, 64, 181, 178, 45}
var sample2Msg = NewMessage("/foo", int32(1000), int32(-1), "hello", float32(1.234), float32(5.678))
