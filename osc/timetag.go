package osc

import (
	"encoding/binary"
	"time"
)

const (
	// MinValue is the minimum value of an OSC Time Tag.
	MinValue = uint64(1)
)

// Timetag represents an OSC Time Tag.
// An OSC Time Tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
type Timetag uint64

// NewTimetagFromTime returns a new OSC time tag object from a time.Time.
func NewTimetagFromTime(timeStamp time.Time) Timetag {
	return Timetag(timeToTimetag(timeStamp))
}

// Time returns the time.
func (t Timetag) Time() time.Time {
	return timetagToTime(t)
}

// FractionalSecond returns the last 32 bits of the OSC time tag. Specifies the
// fractional part of a second.
func (t Timetag) FractionalSecond() uint32 {
	return uint32(t << 32)
}

// SecondsSinceEpoch returns the first 32 bits (the number of seconds since the
// midnight 1900) from the OSC time tag.
func (t Timetag) SecondsSinceEpoch() uint32 {
	return uint32(t >> 32)
}

// TimeTag returns the time tag value
func (t Timetag) TimeTag() uint64 {
	return uint64(t)
}

// MarshalBinary converts the OSC time tag to a byte array.
func (t Timetag) MarshalBinary() (b []byte, err error) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t))
	return
}

// SetTime sets the value of the OSC time tag.
func (t *Timetag) SetTime(time time.Time) {
	*t = Timetag(timeToTimetag(time))
}

// ExpiresIn calculates the number of seconds until the current time is the
// same as the value of the time tag. It returns zero if the value of the
// time tag is in the past.
func (t Timetag) ExpiresIn() time.Duration {
	if t <= 1 {
		return 0
	}

	tt := timetagToTime(t)
	seconds := tt.Sub(time.Now())

	if seconds <= 0 {
		return 0
	}

	return seconds
}

// timeToTimetag converts the given time to an OSC time tag.
//
// An OSC time tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
//
// The time tag value consisting of 63 zero bits followed by a one in the least
// significant bit is a special case meaning "immediately."
func timeToTimetag(t time.Time) (timetag uint64) {
	timetag = uint64((secondsFrom1900To1970 + t.Unix()) << 32)
	return timetag + uint64(t.Nanosecond())
}

// timetagToTime converts the given timetag to a time object.
func timetagToTime(timetag Timetag) (t time.Time) {
	return time.Unix(int64((timetag>>32)-secondsFrom1900To1970), int64(timetag&0xffffffff))
}
