package clock

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetHourOfDay(t int64) int {
	return time.Unix(0, UnixNano(t)).Hour()
}

// UnixSec returns number of years that have elapsed since 00:00:00 1/1/1970 UTC
// to current time t.
// Parameter t could be nanosecond, millisecond, microsecond or second.
func UnixYear(t int64) int64 { return UnixNano(t) / 24 / int64(time.Hour) / 365 }

// UnixSec returns number of months that have elapsed since 00:00:00 1/1/1970 UTC
// to current time t.
// Parameter t could be nanosecond, millisecond, microsecond or second.
func UnixMonth(t int64) int64 { return UnixNano(t) / 24 / int64(time.Hour) / 30 }

// UnixSec returns number of days that have elapsed since 00:00:00 1/1/1970 UTC
// to current time t.
// Parameter t could be nanosecond, millisecond, microsecond or second.
func UnixDay(t int64) int64 { return UnixNano(t) / 24 / int64(time.Hour) }

// UnixSec returns number of minutes that have elapsed since 00:00:00 1/1/1970 UTC
// to current time t.
// Parameter t could be nanosecond, millisecond, microsecond or second.
func UnixMin(t int64) int64 { return UnixNano(t) / int64(time.Minute) }

// UnixSec returns number of seconds that have elapsed since 00:00:00 1/1/1970 UTC
// to current time t.
// Parameter t could be nanosecond, millisecond, microsecond or second.
func UnixSec(t int64) int64 { return UnixNano(t) / int64(time.Second) }

// UnixHour returns number of hours that have elapsed since 00:00:00 1/1/1970 UTC
// to current time t.
// Parameter t could be nanosecond, millisecond, microsecond or second.
func UnixHour(t int64) int64 { return UnixNano(t) / int64(time.Hour) }

// IsNano tells whether an integer t is nano seconds
func IsNano(t int64) bool {
	return t > 1e+18
}

// Midnight returns number of nano seconds elapsed since 0h0m0s 1/1/1970 UTC to
// the midnight of the current day in timezone tzoffset
// E.g: Midnight("+07:00"), Midnight("00:00")
func Midnight(tzoffset string) int64 {
	h, _, _ := SplitTzOffset(tzoffset)
	year, month, day := time.Now().UTC().Date()
	curmidnight := time.Date(year, month, day, 23, 59, 59, 0, time.UTC)
	curmidnight_inzone := curmidnight.Add(-time.Duration(h) * time.Hour)
	return curmidnight_inzone.UnixNano()
}

const OneMonth = 31 * 24 * time.Hour

// ToMili converts t (nanosecond, millisecond, microsecond or second) into
// millisecond integer
func UnixMili(t int64) int64 { return UnixNano(t) / 1e+6 }

func RoundSecNano(t int64) int64 { return UnixSec(t) * int64(time.Second) }

// UnixNano convert t (nanosecond, millisecond, microsecond or second) into
// nanosecond integer
func UnixNano(t int64) int64 {
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

// tzMap map timezone name to UTC timezone, used internally in TimezoneToUTC,
// since call to function time.LoadLocation take very long time, this variable
// is used to cache function responses
var tzMap = &sync.Map{}

// TimezoneToUTC convert timezone name to UTC timezone
// The name should be taken in the IANA Time Zone database, for examples:
//
//	"America/New_York", "Asia/Ho_Chi_Minh"
//
// examples:
//
//	TimezoneToUTC("Asia/Ho_Chi_Minh") -> +07:00
//
// This function use tzMap global variable as cache
// CAUTION: in order to run, OS must have tzdata package (use 'apk add tzdata'
// to install)
func TimezoneToUTC(tzName string) string {
	// predefined cache value, for extreme fast lookup
	switch tzName {
	case "":
		return "+00:00"
	case "Asia/Ho_Chi_Minh":
		return "+07:00"
	}

	// look up in cache
	if tz, ok := tzMap.Load(tzName); ok {
		return tz.(string)
	}

	// cache miss, look up in database
	utc, err := time.LoadLocation(tzName)
	if err != nil {
		return "+00:00"
	}
	_, z := time.Now().In(utc).Zone()
	sign := "+"
	if z < 0 {
		sign = "-"
		z = -z
	}

	h := z / 3600
	m := z % 3600
	hh := strconv.Itoa(h)
	mm := strconv.Itoa(m)
	if len(hh) < 2 {
		hh = "0" + hh
	}

	if len(mm) < 2 {
		mm = "0" + mm
	}
	tz := sign + hh + ":" + mm
	tzMap.Store(tzName, tz)
	return tz
}

// tz: 07:00
func ConvertTimezone(t time.Time, tz string) (year, mon, day, hour, min int, weekday string, err error) {
	tzhour, tzmin, err := SplitTzOffset(tz)
	if err != nil {
		return 0, 0, 0, 0, 0, "", err
	}

	t = t.UTC().Add(time.Hour*time.Duration(tzhour) +
		time.Minute*time.Duration(tzmin))
	return t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(),
		t.Weekday().String(), nil
}

// SplitTzOffset splits timezone offset (e.g: +07:00) to hour and minute
// pair (e.g: 7, 0)
// timezone offset must follow +hh:mm or -hh:mm, otherwise the function
// will return an invalid timezone offset error
// examples:
//
//	SplitTzOffset("+07:00") => 7, 0
//	SplitTzOffset("-07:30") => -7, -30
//	SplitTzOffset("-00:30") => -7, -30
func SplitTzOffset(offset string) (int, int, error) {
	offset = strings.TrimSpace(offset)
	if offset == "" || offset == "0" || offset == "00:00" || offset == "Z" {
		return 0, 0, nil
	}
	sign := 1
	if offset[0] == '-' {
		sign = -1
	} else if offset[0] == '+' {
	} else {
		offset = "+" + offset
	}
	if len(offset) != 6 {
		return 0, 0, fmt.Errorf("invalid timezone offset %s", offset)
	}

	if offset[3] != ':' {
		return 0, 0, fmt.Errorf("invalid timezone offset %s", offset)
	}

	if !isNumeric(offset[1]) || !isNumeric(offset[2]) ||
		!isNumeric(offset[4]) || !isNumeric(offset[5]) {
		return 0, 0, fmt.Errorf("invalid timezone offset %s", offset)
	}
	hour := (offset[1]-'0')*10 + offset[2] - '0'
	min := (offset[4]-'0')*10 + offset[5] - '0'

	// why not 24?, see https://en.wikipedia.org/wiki/List_of_UTC_time_offsets
	// for the list of all possible timezone
	if hour > 14 {
		return 0, 0, fmt.Errorf("invalid timezone offset %s", offset)
	}

	if min > 60 {
		return 0, 0, fmt.Errorf("invalid timezone offset %s", offset)
	}

	if sign == -1 {
		return -int(hour), -int(min), nil
	}
	return int(hour), int(min), nil
}

// GetTimezoneOffset returns the difference, in minutes, between this date as evaluated in the UTC time zone.
// eg: +07:00 -> -420
func GetTimezoneOffset(tz string) int {
	h, m, _ := SplitTzOffset(tz)
	return -(h*60 + m)
}

func SubDays(a, b int64, tz string) int {
	offmin := int64(GetTimezoneOffset(tz))
	atime := time.Unix(a/1000-offmin*60, 0)
	btime := time.Unix(b/1000-offmin*60, 0)
	ahour, amin, _ := atime.UTC().Clock()
	bhour, bmin, _ := btime.UTC().Clock()
	tza := int(atime.Unix()) - ahour*3600 - amin*60
	tzb := int(btime.Unix()) - bhour*3600 - bmin*60

	return (tzb - tza) / 86400
}

func isNumeric(r byte) bool { return '0' <= r && r <= '9' }
