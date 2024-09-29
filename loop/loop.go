package loop

import (
	"fmt"
	"time"
)

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

func LoopErr(f func() error, maxtime int) error {
	i := 0
	return Loop(func() {
		i++
		if i > maxTime
		if err := f(); err != nil {
			panic(err)
		}
	})
}
