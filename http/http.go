package http

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

var (
	ErrUrlIsEmpty = errors.New("url is empty")
	ErrNot200     = errors.New("not 200")
)

// Config used to specific detailed configurations when making http request
type Config struct {
	// map contains HTTP header entries to be injected when make http request
	// contains pair of (header key, header value)
	Header map[string]string

	// maximum amount of time wait for the request to complete, included retry time
	// each call to server only wait for 60 secs
	Timeout time.Duration
}

// g_fhc used to send http request
var g_fhc = &http.Client{Timeout: 60 * time.Second}

// Request send http request to url, it retries automatically on 429 (rate limit) or
// 5xx error
// By default, this method will block no longer than 5 minutes, user can change the
// timeout in config paramater. The method forced to return error when timeout.
// If success, this method returns raw response body, an ErrNot200 is returned if
// the server don't return 2xx code
func Request(method, url string, body []byte, config *Config) ([]byte, error) {
	if url == "" {
		return nil, ErrUrlIsEmpty
	}

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

	// create backoff utility to do retry
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 60 * time.Second
	bo.MaxElapsedTime = timeout
	bo.Reset()

	err := backoff.Retry(func() error {
		var err error
		out, statuscode, err = sendHTTP(method, url, header, body)
		if err != nil {
			// something wrong with the parameters, return nil since retry won't help
			out = []byte(err.Error())
			return nil
		}
		// retry on 429 or 5xx
		if statuscode == 429 || Is5xx(statuscode) {
			return errors.New(string(out))
		}

		// we don't retry on other status code (400, 300)
		return nil
	}, bo)
	if err != nil {
		return out, err
	}

	if !Is2xx(statuscode) {
		return out, ErrNot200
	}

	return out, nil
}

// sendHTTP make an http request to http endpoint
// method, url must not be empty
// this method returns (response body in []byte, status code, and an error)
func sendHTTP(method, url string, header map[string]string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", "Subiz-Gun/4.014")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	res, err := g_fhc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}
	return b, res.StatusCode, res.Body.Close()
}

// Is2xx return whether code is in range of (200; 299)
func Is2xx(code int) bool { return 199 < code && code < 300 }

// Is4xx tells whether code is in range of (300; 399)
func Is4xx(code int) bool { return 399 < code && code < 500 }

// Is5xx tells whether code is in range of (400; 499)
func Is5xx(code int) bool { return 499 < code && code < 600 }
