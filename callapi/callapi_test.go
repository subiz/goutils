package callapi_test

import (
	"testing"
	"bitbucket.org/subiz/goutils/callapi"
	"github.com/jonboulle/clockwork"
	"time"
)

type AsyncSuccessHC struct {
	callsend chan bool
}

func (hc *AsyncSuccessHC) Send(method, url string, header map[string]string, body []byte) (map[string][]string, []byte, int, error) {
	hc.callsend <- true
	return nil, nil, 200, nil
}

func TestAsyncSuccess(t *testing.T) {
	hc := &AsyncSuccessHC{}
	hc.callsend = make(chan bool)
	handler := callapi.NewHandler(hc, nil)
	handler.Post("https://sutu.shop", nil, nil)
	if handler.GetLastState().State != callapi.S_CALLING {
		t.Fatalf("should be calling, got %s", handler.GetLastState().State)
	}
	<-hc.callsend // must call callsend first
	<-handler.Wait()
	if handler.GetLastState().State != callapi.S_STOPPED {
		t.Fatalf("should be STOPPD, got %s", handler.GetLastState().State)
	}
	if handler.GetStatusCode() != 200 {
		t.Fatalf("should be 200, got %d", handler.GetStatusCode())
	}
}

type AsyncBackoffHC struct {
	c int
}

func (hc *AsyncBackoffHC) Send(method, url string, header map[string]string, body []byte) (map[string][]string, []byte, int, error) {
	if hc.c == 3 {
		return nil, nil, 200, nil
	}
	hc.c++
	return nil, nil, 500, nil
}

func TestAsyncBackoff(t *testing.T) {
	hc := &AsyncBackoffHC{}
	clo := clockwork.NewFakeClock()
	handler := callapi.NewHandler(hc, clo)
	handler.Post("https://sutu.shop", nil, nil) // will fail and sleep for 1 sec
	clo.BlockUntil(1) // wait for one sleep
	clo.Advance(1 * time.Second) // should wake goroutine and let it die again
	clo.BlockUntil(1)
	if handler.GetLastState().State != callapi.S_BACKINGOFF {
		t.Fatalf("should be backing off, got %s", handler.GetLastState().State)
	}
	if handler.GetStatusCode() != 500 {
		t.Fatalf("should be 500, got %d", handler.GetStatusCode())
	}

	clo.Advance(1 * time.Second)
	clo.BlockUntil(1)
	if hc.c != 2 {
		t.Fatalf("should equal 2, got %d", hc.c)
	}
	clo.Advance(1 * time.Second)
	clo.BlockUntil(1)
	if hc.c != 3 {
		t.Fatalf("should equal 3, got %d", hc.c)
	}
	then := clo.Now()
	cont := true
	go func() {
		for cont {
			clo.Advance(1 * time.Second)
			clo.BlockUntil(1)
		}
	}()
	<-handler.Wait()
	cont = false
	sub := clo.Now().Sub(then) / time.Second
	if 4 != sub {
		t.Fatalf("should wait for 4 sec, got %d", sub)
	}
	if handler.GetLastState().State != callapi.S_STOPPED {
		t.Fatalf("should be stoped, got %s", handler.GetLastState().State)
	}
	if handler.GetStatusCode() != 200 {
		t.Fatalf("should be 200, got %d", handler.GetStatusCode())
	}
}

func TestAsyncStop(t *testing.T) {
	hc := &AsyncBackoffHC{}
	clo := clockwork.NewFakeClock()
	handler := callapi.NewHandler(hc, clo)
	handler.Post("https://sutu.shop", nil, nil)
	clo.BlockUntil(1)
	handler.Cancel()
	clo.Advance(3 * time.Second)
	if handler.GetLastState().State != callapi.S_CANCELLED {
		t.Fatalf("should be cancelled, got %s", handler.GetLastState().State)
	}
}
