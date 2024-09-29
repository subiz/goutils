package loop

import (
	"fmt"
	"time"
)

func Loop(f func()) {
	for {
		err := func() (e error) {
			defer func() {
				if r := recover(); r != nil {
					if er, _ := r.(error); er != nil {
						e = er
						return
					}
					e = fmt.Errorf("%v", r)
					return
				}
			}()

			f()
			return nil
		}()
		if err == nil {
			break
		}
		fmt.Println("will retries in 3 sec")
		time.Sleep(3 * time.Second)
	}
}

func LoopErr(f func() error, maxtime int) error {
	if maxtime <= 0 {
		maxtime = 1000_000_000
	}
	i := 0
	var lasterr error
	Loop(func() {
		i++
		if i > maxtime {
			return
		}

		if err := f(); err != nil {
			lasterr = err
			panic(err)
		}
		lasterr = nil
	})
	return lasterr
}
