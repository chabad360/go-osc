package osc

import (
	"time"
)

const (
	secondsFrom1900To1970 = 2208988800
)

// Timetag represents an OSC Time Tag.
type Timetag uint64

// NewTimetag returns a new OSC time tag object with the time set to now.
func NewTimetag() Timetag {
	return timeToTimetag(time.Now().UTC())
}

// NewTimetagFromTime returns a new OSC time tag object from a time.Time.
func NewTimetagFromTime(t time.Time) Timetag {
	return timeToTimetag(t)
}

// NewImmediateTimetag creates an OSC Time Tag with only the least significant bit set.
// Compliant implementations should always instantly execute this bundle.
func NewImmediateTimetag() Timetag {
	return Timetag(1)
}

// Time returns the time.
func (t Timetag) Time() time.Time {
	return timetagToTime(t)
}

// FractionalSecond returns the last 32 bits of the OSC time tag. Specifies the fractional part of a second.
func (t Timetag) FractionalSecond() uint32 {
	return uint32(t << 32)
}

// SecondsSinceEpoch returns the first 32 bits (the number of seconds since the midnight 1900) from the OSC time tag.
func (t Timetag) SecondsSinceEpoch() uint32 {
	return uint32(t >> 32)
}

// SetTime sets the value of the OSC time tag.
func (t *Timetag) SetTime(tt time.Time) {
	*t = timeToTimetag(tt)
}

// ExpiresIn calculates the number of seconds until the current time is the same as the value of the time tag.
// It returns zero if the value of the time tag is in the past.
func (t Timetag) ExpiresIn() time.Duration {
	if t <= 1 {
		return 0
	}
	if d := time.Until(timetagToTime(t)); d > 0 {
		return d
	}

	return 0
}

// timeToTimetag converts the given time to an OSC time tag.
func timeToTimetag(t time.Time) (timetag Timetag) {
	return (Timetag(secondsFrom1900To1970+t.Unix()) << 32) + Timetag(t.Nanosecond())
}

// timetagToTime converts the given timetag to a time object.
func timetagToTime(timetag Timetag) (t time.Time) {
	return time.Unix(int64((timetag>>32)-secondsFrom1900To1970), int64(timetag&0xffffffff))
}
