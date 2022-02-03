package osc

import (
	"testing"
	"time"
)

func TestNewImmediateTimetag(t *testing.T) {
	tt := NewImmediateTimetag()
	if i := tt.ExpiresIn(); i != 0 {
		t.Errorf("NewImmediateTimetag() = %d, want 1", tt)
	}
}

func TestNewTimetag(t *testing.T) {
	ti := time.Now()
	tt := NewTimetag()
	if i := tt.ExpiresIn(); i != 0 {
		t.Errorf("NewTimetag() = %d, want %d", tt, NewTimetagFromTime(ti))
	}
}

func TestNewTimetagFromTime(t *testing.T) {
	tt := NewTimetagFromTime(time.Now().Add(time.Second))
	if i := tt.ExpiresIn(); i.Round(time.Millisecond) != time.Second {
		t.Errorf("NewTimetag() = %d, want %d", i.Round(time.Second), time.Second)
	}
}

func TestTimetag_ExpiresIn(t *testing.T) {
	tests := []struct {
		name string
		t    Timetag
		want time.Duration
	}{
		{"one_second", NewTimetagFromTime(time.Now().Add(time.Second)), time.Second},
		{"immediate", NewImmediateTimetag(), 0},
		{"late", NewTimetagFromTime(time.Now().Add(-time.Second)), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.ExpiresIn(); got.Round(time.Millisecond) != tt.want {
				t.Errorf("ExpiresIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestTimetag_FractionalSecond(t *testing.T) {
//	tests := []struct {
//		name string
//		t    Timetag
//		want uint32
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := tt.t.FractionalSecond(); got != tt.want {
//				t.Errorf("FractionalSecond() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestTimetag_SecondsSinceEpoch(t *testing.T) {
//	tests := []struct {
//		name string
//		t    Timetag
//		want uint32
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := tt.t.SecondsSinceEpoch(); got != tt.want {
//				t.Errorf("SecondsSinceEpoch() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestTimetag_SetTime(t *testing.T) {
//	tests := []struct {
//		name string
//		t    Timetag
//		arg time.Time
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			tt.t.SetTime(tt.arg)
//		})
//	}
//}
//
//func TestTimetag_Time(t *testing.T) {
//	tests := []struct {
//		name string
//		t    Timetag
//		want time.Time
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := tt.t.Time(); !got.Equal(tt.want) {
//				t.Errorf("Time() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_timeToTimetag(t *testing.T) {
//	type args struct {
//		t time.Time
//	}
//	tests := []struct {
//		name string
//		args args
//		want Timetag
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := timeToTimetag(tt.args.t); got != tt.want {
//				t.Errorf("timeToTimetag() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_timetagToTime(t *testing.T) {
//	type args struct {
//		timetag Timetag
//	}
//	tests := []struct {
//		name string
//		args args
//		want time.Time
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := timetagToTime(tt.args.timetag); !got.Equal(tt.want) {
//				t.Errorf("timetagToTime() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
