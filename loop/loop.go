package loop

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	"os"
)

type Error struct {
	Description string
	Stack       string
}

func Loop(f func()) {
	for {
		err := func() (e *Error) {
			defer func() {
				if r := recover(); r != nil {
					e = &Error{Description: fmt.Sprintf("%v", r), Stack: getMinifiedStack()}
					return
				}
			}()

			f()
			return nil
		}()
		if err == nil {
			break
		}
		fmt.Println(err.Description)
		fmt.Println(err.Stack)
		fmt.Println("will retries in 3 sec")
		time.Sleep(3 * time.Second)
	}
}

func LoopErr(f func() error) {
	Loop(func() {
		if err := f(); err != nil {
			panic(err)
		}
	})
}

func getMinifiedStack() string {
	stack := ""
	for i := 3; i < 90; i++ {
		_, fn, line, _ := runtime.Caller(i)
		if fn == "" {
			break
		}
		hl := false // highlight
		if strings.Contains(fn, "bitbucket.org/subiz") {
			hl = true
		}
		var split = strings.Split(fn, string(os.PathSeparator))
		var n int

		if len(split) >= 2 {
			n = len(split) - 2
		} else {
			n = len(split)
		}
		fn = strings.Join(split[n:], string(os.PathSeparator))
		if hl {
			stack += fmt.Sprintf("\n→ %s:%d", fn, line)
		} else {
			stack += fmt.Sprintf("\n→ %s:%d", fn, line)
		}
	}
	return stack
}
