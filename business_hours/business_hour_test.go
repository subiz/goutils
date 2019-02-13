package business_hours

import (
	"github.com/subiz/goutils/conv"
	pb "github.com/subiz/header/account"
	"testing"
	"time"
)

func TestIsHoliday(t *testing.T) {
	tcs := []struct {
		holidays  []*pb.BusinessHours_Holiday
		now       string
		tz        string
		isholiday bool
	}{
		{
			[]*pb.BusinessHours_Holiday{{
				Year:  conv.PI32(2018),
				Month: conv.PI32(8),
				Day:   conv.PI32(21),
			}}, "2018-08-21T15:04:00Z", "+07:00", true,
		}, {
			[]*pb.BusinessHours_Holiday{{
				Year:  conv.PI32(2018),
				Month: conv.PI32(8),
				Day:   conv.PI32(21),
			}}, "2018-08-21T23:04:00Z", "+07:00", false,
		}, {
			[]*pb.BusinessHours_Holiday{{
				Year:  conv.PI32(2018),
				Month: conv.PI32(8),
				Day:   conv.PI32(22),
			}, {
				Year:  conv.PI32(2018),
				Month: conv.PI32(8),
				Day:   conv.PI32(21),
			}}, "2018-08-21T23:04:00Z", "+07:00", true,
		},
	}

	for _, tc := range tcs {
		tim, err := time.Parse(time.RFC3339, tc.now)
		if err != nil {
			t.Fatalf("%s: %v", tc.now, err)
		}

		isholiday, err := IsHoliday(&pb.BusinessHours{Holidays: tc.holidays}, tim, tc.tz)
		if err != nil {
			t.Fatalf("%s: %v", tc.now, err)
		}

		if isholiday != tc.isholiday {
			t.Errorf("%s: should be %v, got %v", tc.now, tc.isholiday, isholiday)
		}
	}
}

func TestDuringBusinessHour(t *testing.T) {
	tcs := []struct {
		holidays    []*pb.BusinessHours_Holiday
		workingdays []*pb.BusinessHours_WorkingDay
		now         string
		tz          string
		in          bool
	}{
		{
			[]*pb.BusinessHours_Holiday{{
				Year:  conv.PI32(2018),
				Month: conv.PI32(8),
				Day:   conv.PI32(21),
			}}, []*pb.BusinessHours_WorkingDay{{}}, "2018-08-21T15:04:00Z", "+07:00", false,
		},
		{
			nil, []*pb.BusinessHours_WorkingDay{{
				Weekday:   conv.S("Wednesday"),
				StartTime: conv.S("08:30"),
				EndTime:   conv.S("23:30"),
			}}, "2018-08-21T05:04:00Z", "+07:00", false, // Tuesday
		},
		{
			nil, []*pb.BusinessHours_WorkingDay{{
				Weekday:   conv.S("Tuesday"),
				StartTime: conv.S("08:30"),
				EndTime:   conv.S("23:30"),
			}}, "2018-08-21T05:04:00Z", "+07:00", true, // Tuesday
		},
		{
			nil, []*pb.BusinessHours_WorkingDay{{
				Weekday:   conv.S("Tuesday"),
				StartTime: conv.S("08:30"),
				EndTime:   conv.S("23:30"),
			}}, "2018-08-21T23:04:00Z", "+07:00", false, // Tuesday
		},
	}

	for _, tc := range tcs {
		tim, err := time.Parse(time.RFC3339, tc.now)
		if err != nil {
			t.Fatalf("%s: %v", tc.now, err)
		}

		duringbusiness, err := DuringBusinessHour(&pb.BusinessHours{
			WorkingDays: tc.workingdays,
			Holidays:    tc.holidays,
		}, tim, tc.tz)
		if err != nil {
			t.Fatalf("%s: %v", tc.now, err)
		}

		if duringbusiness != tc.in {
			t.Errorf("%s: should be %v, got %v", tc.now, tc.in, duringbusiness)
		}
	}
}
