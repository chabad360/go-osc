package osc

import (
	"reflect"
	"testing"
)

var temp = &Message{Address: "/composition/layers/1/clips/1/transport/position", Arguments: []interface{}{0.123456789, "hello world"}}
var msg, _ = temp.MarshalBinary()

func BenchmarkParsePacket(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	var p Packet
	for n := 0; n < b.N; n++ {
		p, _ = parsePacket(msg)
	}
	result = p
}

func TestParsePacket(t *testing.T) {
	tests := []testCase{}
	tests = append(tests, messageTestCases...)
	tests = append(tests, bundleTestCases...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePacket(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.obj) {
				t.Errorf("ParsePacket() got = %v, want %v", got, tt.obj)
			}
		})
	}
}

func FuzzParsePacket(f *testing.F) {
	for _, tc := range bundleTestCases {
		f.Add(tc.raw)
	}
	for _, tc := range messageTestCases {
		f.Add(tc.raw)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		packet, err := ParsePacket(data)
		if err != nil {
			return
		}

		dataNew, err := packet.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary(): err != nil on parsed packet %#v: %v", packet, err)
		}

		packet, err = ParsePacket(dataNew)
		if err != nil {
			t.Fatalf("ParsePacket(): err != nil on marshaled packet %#v: %v", packet, err)
		}

		dataNew2, err := packet.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary(): err != nil on double-parsed packet %#v: %v", packet, err)
		}

		if !reflect.DeepEqual(dataNew, dataNew2) {
			t.Fatalf("dataNew != dataNew2: dataNew: %s %v\ndataNew2: %s %v\npacket: %v\n", dataNew, dataNew, dataNew2, dataNew2, packet)
		}
	})
}
