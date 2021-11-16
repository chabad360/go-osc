package osc

import (
	"reflect"
	"testing"
)

func TestReadPacket(t *testing.T) {
	for _, tt := range []struct {
		desc string
		msg  string
		pkt  Packet
		ok   bool
	}{
		{"no_args",
			"/a/b/c" + nulls(2) + "," + nulls(3),
			makePacket("/a/b/c", nil),
			true},
		{"string_arg",
			"/d/e/f" + nulls(2) + ",s" + nulls(2) + "foo" + nulls(1),
			makePacket("/d/e/f", []string{"foo"}),
			true},
		{"empty", "", nil, false},
		{"designed",
			string(msg), temp, true},
	} {
		pkt, err := ReadPacket([]byte(tt.msg))
		if err != nil && tt.ok {
			t.Fatalf("%s: ReadPacket() returned unexpected error; %s", tt.desc, err)
		}
		if err == nil && !tt.ok {
			t.Errorf("%s: ReadPacket() expected error", tt.desc)
		}
		if !tt.ok {
			continue
		}

		pktBytes, err := pkt.MarshalBinary()
		if err != nil {
			t.Errorf("%s: failure converting pkt to byte array; %s", tt.desc, err)
			continue
		}
		ttpktBytes, err := tt.pkt.MarshalBinary()
		if err != nil {
			t.Errorf("%s: failure converting tt.pkt to byte array; %s", tt.desc, err)
			continue
		}
		if !reflect.DeepEqual(pktBytes, ttpktBytes) {
			t.Errorf("%s: ReadPacket() as bytes = '%s', want = '%s'", tt.desc, pktBytes, ttpktBytes)
			continue
		}
	}
}

var temp = &Message{Address: "/composition/layers/1/clips/1/transport/position", Arguments: []interface{}{0.123456789, "hello world"}}
var msg, _ = temp.MarshalBinary()

func BenchmarkReadPacket(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	var p Packet
	for n := 0; n < b.N; n++ {
		p, _ = ReadPacket(msg)
	}
	result = p
}
