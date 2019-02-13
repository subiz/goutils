package clock

import (
	"testing"
	"time"
)

func TestHourOfDay(t *testing.T) {
	println(GetHourOfDay(time.Now().UnixNano()))
	println(RoundSecNano(time.Now().UnixNano()))
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
		year, mon, day, hour, min, err := ConvertTimezone(tim, tc.tz)
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
