package client

import (
	"compress/gzip"
	"net/http"
	"net/url"
	"time"
)

type transport struct {
	t http.RoundTripper
}

func (adt *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/112.0")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := adt.t.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = gzReader
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
	}
	return resp, err
}

func GetClient(proxy string) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()

	// use proxy only for www.theonedb.com
	if proxy != "" {
		u, err := url.Parse(proxy)

		t.Proxy = func(r *http.Request) (*url.URL, error) {
			if r.URL.Hostname() == "www.theonedb.com" {
				return u, err
			}
			return nil, nil
		}
	} else {
		t.Proxy = nil
	}
	t.MaxIdleConns = 50
	t.MaxConnsPerHost = 5
	t.MaxIdleConnsPerHost = 10
	t.ResponseHeaderTimeout = time.Second * 20
	t.TLSHandshakeTimeout = 20 * time.Second

	var client = &http.Client{
		Transport: &transport{t},
		Timeout:   time.Second * 25,
	}

	return client
}
