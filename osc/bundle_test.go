package osc

import (
	"reflect"
	"testing"
)

func TestBundle_MarshalBinary(t *testing.T) {
	for _, tt := range bundleTestCases {
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

func TestBundle_UnmarshalBinary(t *testing.T) {
	for _, tt := range bundleTestCases {
		t.Run(tt.name, func(t *testing.T) {
			m := new(Bundle)
			if err := m.UnmarshalBinary(tt.raw); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, tt.obj) {
				t.Errorf("MarshalBinary() got = %v, want %v", m, tt.obj)
			}
		})
	}
}
