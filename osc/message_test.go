package osc

import (
	"reflect"
	"testing"
)

func TestMessage_Append(t *testing.T) {
	oscAddress := "/address"
	message := NewMessage(oscAddress)

	message.Append("string argument")
	message.Append(int32(123456789))
	message.Append(true)

	if len(message.Arguments) != 3 {
		t.Errorf("Number of arguments should be %d and is %d", 3, len(message.Arguments))
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

func TestMessage_MarshalBinary(t *testing.T) {
	for _, tt := range messageTestCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.obj.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.raw) {
				t.Errorf("MarshalBinary() got = %s, want %s", got, tt.raw)
			}
		})
	}
}

func TestMessage_UnmarshalBinary(t *testing.T) {
	for _, tt := range messageTestCases {
		t.Run(tt.name, func(t *testing.T) {
			m := new(Message)
			if err := m.UnmarshalBinary(tt.raw); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, tt.obj) {
				t.Errorf("MarshalBinary() got = %v, want %v", m, tt.obj)
			}
		})
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
