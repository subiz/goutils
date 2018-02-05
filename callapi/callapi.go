package callapi

import (
	"github.com/cenkalti/backoff"
	"github.com/jonboulle/clockwork"
	"sync"
	"time"
)

type Call struct {
	hc  HttpClient
	clo clockwork.Clock

}

type HCState struct {
	BackoffCount int
	State        string
}

type AsyncCall interface {
	Post(url string, header map[string]string, body []byte) AsyncRestAPIHandler
}

type AsyncRestAPIHandler interface {
	Post(url string, header map[string]string, body []byte)
	Cancel()
	Wait() <-chan bool
	GetLastState() HCState
	GetBody() []byte
	GetStatusCode() int
	GetHeader(string) string
}

func (c *Call) Post(url string, header map[string]string, body []byte) AsyncRestAPIHandler {
	handler := NewHandler(c.hc, c.clo)
	handler.Post(url, header, body)
	return handler
}

func NewCall(hc HttpClient, clo clockwork.Clock) *Call {
	if hc == nil {
		hc = NewFHC()
	}

	if clo == nil {
		clo = clockwork.NewRealClock()
	}
	return &Call{
		hc:  hc,
		clo: clo,
	}
}

const (
	S_CALLING    = "calling"
	S_BACKINGOFF = "backingoff"
	S_STOPPED    = "stopped"
	S_CANCELLED  = "cancelled"
)

type Handler struct {
	body       []byte
	statuscode int
	header     map[string][]string
	laststate  HCState
	wc         chan bool
	hc         HttpClient
	clo        clockwork.Clock
	lock       *sync.Mutex
	cancelchan chan bool
	canceldone chan bool
	donechan   chan bool
}

func NewHandler(hc HttpClient, clo clockwork.Clock) *Handler {
	if hc == nil {
		hc = NewFHC()
	}

	if clo == nil {
		clo = clockwork.NewRealClock()
	}
	return &Handler{
		lock:       &sync.Mutex{},
		hc:         hc,
		clo:        clo,
		laststate:  HCState{BackoffCount: 0, State: S_STOPPED},
		wc:         make(chan bool),
		cancelchan: make(chan bool),
		donechan:   make(chan bool),
		canceldone: make(chan bool),
	}
}

type resp struct {
	code   int
	body   []byte
	header map[string][]string
	err    error
}

func (h *Handler) asyncPostHttp(url string, header map[string]string, body []byte, c chan resp) {
	res := resp{}
	res.header, res.body, res.code, res.err = h.hc.Send("POST", url, header, body)
	c <- res
}

func (h *Handler) Post(url string, header map[string]string, body []byte) {
	h.lock.Lock()
	if h.laststate.State != S_STOPPED {
		h.lock.Unlock()
		panic("handler is in middle of something (not stopped), got " + h.laststate.State)
	}
	h.laststate = HCState{BackoffCount: 0, State: S_CALLING}
	go h.SyncPost(url, header, body)
	h.lock.Unlock()
}

func (h *Handler) SyncPost(url string, header map[string]string, body []byte) {
	bf := &backoff.ExponentialBackOff{
		Multiplier:          2,
		RandomizationFactor: 0.1,
		InitialInterval:     1 * time.Second,
		Clock:               h.clo,
	}
	bf.Reset()
	cancelled, c := false, 15
	callchan := make(chan resp)

	for c > 0 {
		if cancelled {
			return
		}
		go h.asyncPostHttp(url, header, body, callchan)
		select {
		case res := <-callchan:
			if cancelled {
				return
			}
			h.lock.Lock()
			h.body, h.header, h.statuscode = res.body, res.header, res.code
			if !Is5xx(res.code) && res.code != 429 { // success or fail
				h.laststate = HCState{BackoffCount: c, State: S_STOPPED}
				h.donechan <- true
				h.lock.Unlock()
			}
			// only retrying on 500 or 429
			h.laststate = HCState{BackoffCount: c, State: S_BACKINGOFF}
			h.lock.Unlock()
			select {
			case <-h.cancelchan:
				if cancelled {
					return
				}
				cancelled = true
				h.lock.Lock()
				h.laststate = HCState{BackoffCount: c, State: S_CANCELLED}
				h.lock.Unlock()
				h.canceldone <- true
				return
			case <-h.clo.After(bf.NextBackOff()):
 			}
		case <-h.cancelchan:
			if cancelled {
				return
			}
			cancelled = true
			h.lock.Lock()
			h.laststate = HCState{BackoffCount: c, State: S_CANCELLED}
			h.lock.Unlock()
			h.canceldone <- true
			return
		}
		c--
	}
}

func (h *Handler) Cancel() {
	h.cancelchan <- true
	<-h.canceldone
}

func (h *Handler) Wait() <-chan bool {
	return h.donechan
}

func (h *Handler) GetLastState() HCState {
	h.lock.Lock()
	s := h.laststate
	h.lock.Unlock()
	return s
}

func (h *Handler) GetBody() []byte {
	h.lock.Lock()
	b := h.body
	h.lock.Unlock()
	return b
}

func (h *Handler) GetStatusCode() int {
	h.lock.Lock()
	c := h.statuscode
	h.lock.Unlock()
	return c
}

func (h *Handler) GetHeader(header string) string {
	h.lock.Lock()
	hea := h.header[header]
	h.lock.Unlock()
	if len(hea) == 0 {
		return ""
	}
	return hea[0]
}

func Is2xx(code int) bool {
	return 199 < code && code < 300
}

func Is4xx(code int) bool {
	return 399 < code && code < 500
}

func Is5xx(code int) bool {
	return 499 < code && code < 600
}
