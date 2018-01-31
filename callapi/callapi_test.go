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
	callsend chan bool
	c int
}

func (hc *AsyncBackoffHC) Send(method, url string, header map[string]string, body []byte) (map[string][]string, []byte, int, error) {
	if hc.c == 2 {
		return nil, nil, 200, nil
	}
	if hc.c == 1 {
		hc.callsend <- true
		hc.callsend <- true
	}
	hc.c++
	return nil, nil, 500, nil
}

func TestAsyncBackoff(t *testing.T) {
	hc := &AsyncBackoffHC{}
	hc.callsend = make(chan bool)
	clo := clockwork.NewFakeClock()
	handler := callapi.NewHandler(hc, clo)
	handler.Post("https://sutu.shop", nil, nil)
	go func() {
		for {
			clo.Advance(10 * time.Second)
			clo.Advance(10 * time.Second)
			clo.Advance(10 * time.Second)
		}
	}()
	<- hc.callsend
	if handler.GetLastState().State != callapi.S_BACKINGOFF {
		t.Fatalf("should be backing off, got %s", handler.GetLastState().State)
	}
	if handler.GetStatusCode() != 500 {
		t.Fatalf("should be 500, got %d", handler.GetStatusCode())
	}
	<- hc.callsend
	clo.Advance(10 * time.Second)
	<- handler.Wait()
	if handler.GetLastState().State != callapi.S_STOPPED {
		t.Fatalf("should be stoped, got %s", handler.GetLastState().State)
	}
	if handler.GetStatusCode() != 200 {
		t.Fatalf("should be 200, got %d", handler.GetStatusCode())
	}
}

func TestAsyncStop(t *testing.T) {
	hc := &AsyncBackoffHC{}
	hc.callsend = make(chan bool)
	clo := clockwork.NewFakeClock()
	handler := callapi.NewHandler(hc, clo)
	handler.Post("https://sutu.shop", nil, nil)
	go func() {
		for {
			clo.Advance(10 * time.Second)
			clo.Advance(10 * time.Second)
			clo.Advance(10 * time.Second)
		}
	}()
	<- hc.callsend
	if handler.GetLastState().State != callapi.S_BACKINGOFF {
		t.Fatalf("should be backing off, got %s", handler.GetLastState().State)
	}
	go func() {<- hc.callsend}()
	handler.Cancel()
	if handler.GetLastState().State != callapi.S_CANCELLED {
		t.Fatalf("should be cancelled, got %s", handler.GetLastState().State)
	}
}
