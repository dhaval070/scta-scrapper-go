package client

import (
	"net/http"
	"net/url"
)

type transport struct {
	t http.RoundTripper
}

func (adt *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/112.0")
	req.Header.Add("Accept", "text/html")
	return adt.t.RoundTrip(req)
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
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	var client = &http.Client{
		Transport: &transport{t},
	}

	return client
}
