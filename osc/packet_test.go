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
