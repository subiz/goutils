package log

import (
	"testing"
)

func TestPrint(t *testing.T) {
	Info("haha", "he he", 4, 20)
}

func TestTimmer(t *testing.T) {
	Time("haivan")
	TimeCheck("haivan", "mot")
	TimeCheck("haivan", "mot", "hai")
	TimeCheck("haivan", "ba")
	TimeEnd("haivan")
	TimeCheck("haivan", "ba")
}
