package clock

import (
	"time"
)

func GetMonthFromNano(created int64) int64 {
	return ToNano(created) / int64(time.Hour) / 24 / 30
}

func GetHourOfDay(t int64) int {
	return time.Unix(0, ToNano(t)).Hour()
}

func ToDay(t int64) int64 {
	return ToNano(t) / 24 / int64(time.Hour)
}

func ToMin(t int64) int64 {
	return ToNano(t) / int64(time.Minute)
}

func ToSec(t int64) int64 {
	return ToNano(t) / int64(time.Second)
}

func ToHour(t int64) int64 {
	return ToNano(t) / int64(time.Hour)
}

func IsNano(t int64) bool {
	return t > 1e+18
}

func GetThisYear() int64 {
	return GetMonthFromNano(time.Now().UnixNano()) / 12
}

func Now() time.Time {
	return time.Now()
}

const OneMonth = 31 * 24 * time.Hour

func ToMili(t int64) int64 {
	return ToNano(t) / 1e+6
}

func RoundSecNano(t int64) int64 {
	return ToSec(t) * int64(time.Second)
}

func ToNano(t int64) int64 {
	if t > 1e+18 { // nanoseconds
		return t
	}
	if t > 1e+15 { // microseconds
		return t * 1e+3
	}
	if t > 1e+12 { // milliseconds
		return t * 1e+6
	}
	return t * 1e+9 // seconds
}
