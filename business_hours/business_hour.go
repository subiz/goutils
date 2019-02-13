package business_hours

import (
	"fmt"
	pb "github.com/subiz/header/account"
	"strconv"
	"strings"
	"time"
)

type WorkingTime int64

func (t WorkingTime) UnixNano() int64 {
	return int64(time.Duration(t) / time.Hour)
}

type BusinessHours struct {
	data *pb.BusinessHours
}

func NewBusinessHours(bh *pb.BusinessHours) *BusinessHours {
	return &BusinessHours{data: bh}
}

func (bh *BusinessHours) DuringBusinessHour(date time.Time) bool {
	if bh.IsHoliday(date) {
		return false
	}
	if len(bh.data.GetWorkingDays()) == 0 {
		return true
	}
	for _, wd := range bh.data.GetWorkingDays() {
		if wd.GetWeekday() == date.Weekday().String() {
			start, _ := parseTime(wd.GetStartTime())
			end, _ := parseTime(wd.GetEndTime())
			if date.UnixNano() >= start.UnixNano() && date.UnixNano() <= end.UnixNano() {
				return true
			}
		}
	}
	return false
}

func (bh *BusinessHours) IsHoliday(date time.Time) bool {
	for _, h := range bh.data.GetHolidays() {
		if h.GetWeekday() == date.Weekday().String() {
			return true
		}
		if h.GetDay() == int32(date.Day()) && h.GetMonth() == int32(date.Month()) {
			if h.GetYear() == 0 {
				return true
			} else if h.GetYear() == int32(date.Year()) {
				return true
			}
		}
	}
	return false
}

// parseTime from string 10:25 convert to WorkingTime (int64)
func parseTime(t string) (WorkingTime, error) {
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

	return WorkingTime(time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute), nil
}
