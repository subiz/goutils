package clock

import (
	"testing"
	"time"
)

func TestHourOfDay(t *testing.T) {
	println(GetHourOfDay(time.Now().UnixNano()))
	println(RoundSecNano(time.Now().UnixNano()))
}
