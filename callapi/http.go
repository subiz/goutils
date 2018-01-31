package callapi

import (
	"time"
	"net/http"
	"bytes"
	"io/ioutil"
)

type HttpClient interface {
	Send(method, url string, header map[string]string, body []byte) (map[string][]string, []byte, int, error)
}

// FHC: fast http client
type FHC struct {
	client *http.Client
}

func NewFHC() *FHC {
	fhc := &FHC{}
	fhc.client = &http.Client{
		Timeout: 1 * time.Minute,
	}
	return fhc
}

func (f FHC) Send(method, url string, header map[string]string, body []byte) (map[string][]string, []byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, 0, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", "Subiz-Gun/4.014")
	res, err := f.client.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return res.Header, b, res.StatusCode, err
}
