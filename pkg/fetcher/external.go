package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type FetcherHttp struct {
	client     http.Client
	serviceUrl string
}

func NewFetcherHttp(client http.Client, serviceUrl string) *FetcherHttp {
	return &FetcherHttp{
		client:     client,
		serviceUrl: serviceUrl,
	}
}

func (f *FetcherHttp) Fetch(scrapeUrl, class string) (string, error) {
	u, err := url.Parse(f.serviceUrl)
	if err != nil {
		return "", err
	}

	qry := u.Query()
	qry.Add("scrape_url", scrapeUrl)
	qry.Add("class", class)

	u.RawQuery = qry.Encode()

	resp, err := f.client.Get(u.String())

	if err != nil {
		return "", fmt.Errorf("client err %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("resp read err %w", err)
	}
	addr := strings.Trim(string(b), " \n")

	fmt.Printf("address for %s = %s\n", scrapeUrl, addr)
	return addr, nil
}
