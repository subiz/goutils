package clock

import (
	"testing"
	"time"
)

func TestHourOfDay(t *testing.T) {
	println(GetHourOfDay(time.Now().UnixNano()))
	println(RoundSecNano(time.Now().UnixNano()))
}

func TestTimezoneOffset(t *testing.T) {
	tcs := []struct {
		tz  string
		min int
	}{
		{"", 0},
		{"+00:00", 0},
		{"+07:00", -420},
		{"+07:30", -450},
		{"-00:30", 30},
	}

	for _, tc := range tcs {
		min := GetTimezoneOffset(tc.tz)
		if tc.min != min {
			t.Errorf("%s: should be %d got %d", tc.tz, tc.min, min)
		}
	}
}

func TestSubDay(t *testing.T) {
	tcs := []struct {
		a   string
		b   string
		tz  string
		day int
	}{
		{"2019-10-12T16:20:50.52Z", "2019-10-12T18:21:50.52Z", "00:00", 0},
		{"2019-10-12T16:20:50Z", "2019-10-12T18:20:50Z", "+07:00", 1},
		{"2019-10-12T18:20:50Z", "2019-10-12T16:20:50Z", "+07:00", -1},
	}

	for i, tc := range tcs {
		aunix, err := time.Parse(time.RFC3339, tc.a)
		if err != nil {
			panic(err)
		}
		bunix, _ := time.Parse(time.RFC3339, tc.b)
		day := SubDays(aunix.UnixMilli(), bunix.UnixMilli(), tc.tz)
		if tc.day != day {
			t.Errorf("%d: should be %d got %d", i, tc.day, day)
		}
	}
}

func TestConvertTimezone(t *testing.T) {
	tcs := []struct {
		intime string
		tz     string
		year   int
		month  int
		day    int
		hour   int
		min    int
	}{
		{"2018-01-02T15:04:00Z", "+07:00", 2018, 1, 2, 22, 4},
		{"2018-01-02T15:04:00Z", "+07:30", 2018, 1, 2, 22, 34},
		{"2018-01-02T00:00:00Z", "-00:30", 2018, 1, 1, 23, 30},
	}

	for _, tc := range tcs {
		tim, err := time.Parse(time.RFC3339, tc.intime)
		if err != nil {
			t.Fatalf("%s: %v", tc.intime, err)

		}
		year, mon, day, hour, min, _, err := ConvertTimezone(tim, tc.tz)
		if err != nil {
			t.Fatalf("%s: %v", tc.intime, err)
		}

		if year != tc.year || mon != tc.month || day != tc.day || hour != tc.hour ||
			min != tc.min {
			t.Errorf("%s: should be %d, %d, %d, %d, %d, got %d, %d, %d, %d, %d", tc.intime,
				tc.year, tc.month, tc.day, tc.hour, tc.min, year, mon, day, hour, min)
		}
	}
}
