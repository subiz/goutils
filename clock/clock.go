package clock

import "time"

func GetMonthFromNano(created int64) int64 {
	if IsNano(created) {
		return created / int64(time.Hour) / 24 / 30
	} else {
		return created / 1000 / 60 / 60 / 24 / 30
	}
}

func ToDay(t int64) int64 {
	if IsNano(t) {
		return t / 24 / int64(time.Hour)
	} else { // t is microsecond, convert t to nano
		return t * 1000000 / 24 / int64(time.Hour)
	}
}

func ToMin(t int64) int64 {
	if IsNano(t) {
		return t / int64(time.Minute)
	} else { // t is microsecond, convert t to nano
		return t * 1000000 / int64(time.Minute)
	}
}

func ToSec(t int64) int64 {
	if IsNano(t) {
		return t / int64(time.Second)
	} else { // t is microsecond, convert t to nano
		return t * 1000000 / int64(time.Second)
	}
}

func ToHour(t int64) int64 {
	if IsNano(t) {
		return t / int64(time.Hour)
	} else { // t is microsecond, convert t to nano
		return t * 1000000 / int64(time.Hour)
	}
}

func IsNano(t int64) bool {
	return t > 1000000000000000
}

func GetThisYear() int64 {
	return GetMonthFromNano(time.Now().UnixNano()) / 12
}

func Now() time.Time {
	return time.Now()
}

const OneMonth = 31 * 24 * time.Hour

func ToMili(t int64) int64 {
	if t > 1000000000000000 {
		return t / 1000000
	}
	return t
}
