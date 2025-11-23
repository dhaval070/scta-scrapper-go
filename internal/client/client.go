package client

import (
	"compress/gzip"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

type transport struct {
	t                  http.RoundTripper
	domainSems         map[string]chan struct{}
	domainSemsMu       sync.Mutex
	maxRequestsPerHost int
	jar                http.CookieJar
}

func (adt *transport) getDomainSemaphore(host string) chan struct{} {
	if adt.maxRequestsPerHost <= 0 {
		return nil
	}

	adt.domainSemsMu.Lock()
	defer adt.domainSemsMu.Unlock()

	sem, exists := adt.domainSems[host]
	if !exists {
		sem = make(chan struct{}, adt.maxRequestsPerHost)
		adt.domainSems[host] = sem
	}
	return sem
}

func (adt *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rate limit by domain
	if adt.maxRequestsPerHost > 0 && req.URL.Host != "" {
		sem := adt.getDomainSemaphore(req.URL.Host)
		if sem != nil {
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release
		}
	}

	// Set MBSW_IS_HUMAN cookie for any domain if not already set
	if adt.jar != nil && req.URL.Host != "" {
		existingCookies := adt.jar.Cookies(req.URL)
		hasCookie := false
		for _, c := range existingCookies {
			if c.Name == "MBSW_IS_HUMAN" {
				hasCookie = true
				break
			}
		}

		t := time.Now()
		t = t.Add(time.Hour * 48 * -1)
		val := t.Format("Jan-02 3:4 PM")

		if !hasCookie {
			cookie := &http.Cookie{
				Name:  "MBSW_IS_HUMAN",
				Value: "Passed " + val,
				// Value:    "Passed Nov-22 1:45 PM",
				Path:     "/",
				Domain:   req.URL.Hostname(),
				SameSite: http.SameSiteStrictMode,
			}
			adt.jar.SetCookies(req.URL, []*http.Cookie{cookie})
		}
	}

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

func GetClient(proxy string, maxRequestsPerHost int) *http.Client {
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
	t.MaxIdleConns = 200
	t.MaxConnsPerHost = maxRequestsPerHost
	t.MaxIdleConnsPerHost = 10
	t.IdleConnTimeout = 3 * time.Second
	t.ResponseHeaderTimeout = time.Second * 20
	t.TLSHandshakeTimeout = 20 * time.Second

	jar, _ := cookiejar.New(nil)

	var client = &http.Client{
		Transport: &transport{
			t:                  t,
			domainSems:         make(map[string]chan struct{}),
			maxRequestsPerHost: maxRequestsPerHost,
			jar:                jar,
		},
		Timeout: time.Second * 25,
		Jar:     jar,
	}

	return client
}
