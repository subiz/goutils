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

func TestAssert(t *testing.T) {
	Assert(nil, nil)
	Assert(1, 1)
	Assert("a", "a")
	Assert("a", "b")
}

func TestNotAssert(t *testing.T) {
	NotAssert("a", nil)
	NotAssert(1, 2)
	NotAssert("a", "b")
	NotAssert("a", "a")
}
