package business_hours

import (
	"fmt"
	"github.com/subiz/goutils/clock"
	pb "github.com/subiz/header/account"
	"strconv"
	"strings"
	"time"
)

func DuringBusinessHour(bh *pb.BusinessHours, date time.Time, tz string) (bool, error) {
	isholiday, err := IsHoliday(bh, date, tz)
	if err != nil {
		return false, err
	}
	if isholiday {
		return false, nil
	}

	if len(bh.GetWorkingDays()) == 0 {
		return true, nil
	}

	_, _, _, h, m, weekday, err := clock.ConvertTimezone(date, tz)
	if err != nil {
		return false, err
	}
	currentmin := h*60 + m

	for _, wd := range bh.GetWorkingDays() {
		if wd.GetWeekday() == weekday {
			start, err := toMinute(wd.GetStartTime())
			if err != nil {
				return false, err
			}
			end, err := toMinute(wd.GetEndTime())
			if err != nil {
				return false, err
			}

			if start <= currentmin && currentmin <= end {
				return true, nil
			}
		}
	}
	return false, nil
}

// IsHoliday tells where ther date (within timezone tz) is in holiday list
func IsHoliday(bh *pb.BusinessHours, date time.Time, tz string) (bool, error) {
	y, m, d, _, _, _, err := clock.ConvertTimezone(date, tz)
	if err != nil {
		return false, err
	}
	for _, h := range bh.GetHolidays() {
		if int32(y) == h.GetYear() && int32(m) == h.GetMonth() &&
			int32(d) == h.GetDay() {
			return true, nil
		}
	}
	return false, nil
}

// toMinute from string 10:25 convert to number of minutes: 675
func toMinute(t string) (int, error) {
	var mins, hours int
	var err error

	parts := strings.SplitN(t, ":", 2)
	switch len(parts) {
	case 1:
		mins, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
	case 2:
		hours, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}

		mins, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("invalid time: %s", t)
	}

	if mins > 59 || mins < 0 || hours > 23 || hours < 0 {
		return 0, fmt.Errorf("invalid time: %s", t)
	}

	return hours*60 + mins, nil
}
