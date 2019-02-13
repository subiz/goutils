package clock

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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

// IsNano tells whether an integer t is nano seconds
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

// ToMili converts t (nanosecond, millisecond, microsecond or second) into
// millisecond integer
func ToMili(t int64) int64 {
	return ToNano(t) / 1e+6
}

func RoundSecNano(t int64) int64 {
	return ToSec(t) * int64(time.Second)
}

// ToNano convert t (nanosecond, millisecond, microsecond or second) into
// nanosecond integer
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

// tzMap map timezone name to UTC timezone, used internally in TimezoneToUTC,
// since call to function time.LoadLocation take very long time, this variable
// is used to cache function responses
var tzMap = &sync.Map{}

// TimezoneToUTC convert timezone name to UTC timezone
// The name should be taken in the IANA Time Zone database, for examples:
//   "America/New_York", "Asia/Ho_Chi_Minh"
// examples:
//   TimezoneToUTC("Asia/Ho_Chi_Minh") -> +07:00
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
//  SplitTzOffset("+07:00") => 7, 0
//  SplitTzOffset("-07:30") => -7, -30
//  SplitTzOffset("-00:30") => -7, -30
func SplitTzOffset(offset string) (int, int, error) {
	offset = strings.TrimSpace(offset)
	if len(offset) != 6 {
		return 0, 0, fmt.Errorf("invalid timezone offset %s", offset)
	}

	sign := 1
	if offset[0] == '-' {
		sign = -1
	} else if offset[0] == '+' {
	} else {
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

func isNumeric(r byte) bool { return '0' <= r && r <= '9' }
