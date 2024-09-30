package http

import (
	"bytes"
	"errors"
	"io"
	nethttp "net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
)

var clientPool = sync.Pool{
	New: func() any {
		var t = nethttp.DefaultTransport.(*nethttp.Transport).Clone()
		t.MaxIdleConns = 200
		t.MaxConnsPerHost = 200
		t.MaxIdleConnsPerHost = 200
		return &Client{
			HttpClient: &nethttp.Client{
				Timeout:   120 * time.Second,
				Transport: t,
			},
		}
	},
}

// Config used to specific detailed configurations when making http request
type Config struct {
	// map contains HTTP header entries to be injected when make http request
	// contains pair of (header key, header value)
	Header map[string]string

	// maximum amount of time wait for the request to complete, included retry time
	// each call to server only wait for 60 secs
	Timeout time.Duration
}

// which provide simpler syntax and exponential backoff retries.
type Client struct {
	HttpClient *nethttp.Client
}

func NewClient() *Client {
	return &Client{HttpClient: &nethttp.Client{Timeout: 60 * time.Second}}
}

// Request sends http request to url, it retries automatically on
// 429 (rate limit) or 5xx error
// By default, this method will block no longer than 5 minutes, user can change
// the timeout in config paramater. The method forced to return error when
// timeout.
// If success, this method returns raw response body, an ErrNot200 is returned
// if the server don't return 2xx code.
func (me *Client) Request(method, url string, body []byte, config *Config) ([]byte, int, nethttp.Header) {
	var header map[string]string
	timeout := 5 * time.Minute
	if config != nil {
		header = config.Header
		if config.Timeout > 0 {
			timeout = config.Timeout
		}
	}
	var out []byte     // raw response body
	var statuscode int // returned status code, -1 indicates internal error
	var respheader nethttp.Header

	// create backoff utility to do retry
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 60 * time.Second
	bo.MaxElapsedTime = timeout
	bo.Reset()

	err := backoff.Retry(func() error {
		out, statuscode, respheader = sendHTTP(me.HttpClient, method, url, header, body)
		// retry on 429 or 5xx
		if statuscode == 429 || Is5xx(statuscode) {
			return errors.New("retry")
		}

		// we don't retry on other status code (400, 300)
		return nil
	}, bo)
	if err != nil {
		return out, -2, respheader
	}

	return out, statuscode, respheader
}

// sendHTTP make an http request to http endpoint
// method, url must not be empty
// this method returns (response body in []byte, status code, and an error)
func sendHTTP(client *nethttp.Client, method, url string, header map[string]string, body []byte) ([]byte, int, nethttp.Header) {
	var req *nethttp.Request
	var err error
	if body == nil {
		req, err = nethttp.NewRequest(method, url, nil)
	} else {
		req, err = nethttp.NewRequest(method, url, bytes.NewReader(body))
	}
	if err != nil {
		return []byte(err.Error()), -1, nil
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", "Subiz-Gun/4.016")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return []byte(err.Error()), 0, nil
	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte(err.Error()), -5, nil
	}
	return b, res.StatusCode, res.Header
}

// Is2xx return whether code is in range of (200; 299)
func Is2xx(code int) bool { return 199 < code && code < 300 }

// Is4xx tells whether code is in range of (300; 399)
func Is4xx(code int) bool { return 399 < code && code < 500 }

// Is5xx tells whether code is in range of (400; 499)
func Is5xx(code int) bool { return 499 < code && code < 600 }

// Request use default client to sends http request to url, it retries
// automatically on 429 (rate limit) or  5xx error
// By default, this method will block no longer than 5 minutes, user can change
// the timeout in config paramater. The method forced to return error when
// timeout.
// If success, this method returns raw response body, an ErrNot200 is returned if
// the server don't return 2xx code.
func Request(method, url string, body []byte, config *Config) ([]byte, int, nethttp.Header) {
	client := clientPool.Get().(*Client)
	defer func() {
		clientPool.Put(client)
	}()
	return client.Request(method, url, body, config)
}

func Get(url string, header map[string]string) ([]byte, int, nethttp.Header) {
	return Request("GET", url, nil, &Config{
		Header:  header,
		Timeout: 1 * time.Minute,
	})
}

func Head(url string, header map[string]string) ([]byte, int, nethttp.Header) {
	return Request("HEAD", url, nil, &Config{
		Header:  header,
		Timeout: 1 * time.Minute,
	})
}

func Post(url string, body []byte, header map[string]string) ([]byte, int, nethttp.Header) {
	return Request("POST", url, body, &Config{
		Header:  header,
		Timeout: 1 * time.Minute,
	})
}

func Patch(url string, body []byte, header map[string]string) ([]byte, int, nethttp.Header) {
	return Request("PATCH", url, body, &Config{
		Header:  header,
		Timeout: 1 * time.Minute,
	})
}

func Put(url string, body []byte, header map[string]string) ([]byte, int, nethttp.Header) {
	return Request("PUT", url, body, &Config{
		Header:  header,
		Timeout: 1 * time.Minute,
	})
}

func Delete(url string, body []byte, header map[string]string) ([]byte, int, nethttp.Header) {
	return Request("DELETE", url, body, &Config{
		Header:  header,
		Timeout: 1 * time.Minute,
	})
}
