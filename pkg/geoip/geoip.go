package geoip

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type (
	Fetcher struct {
		timeout   time.Duration
		endpoint  string
		accessKey string
	}

	FetcherArgs struct {
		Timeout   time.Duration
		Endpoint  string
		AccessKey string
	}

	Info struct {
		IP          string  `json:"ip"`
		Hostname    string  `json:"hostname"`
		Country     string  `json:"country_name"`
		CountryCode string  `json:"country_code"`
		City        string  `json:"City"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	}
)

var (
	ErrNoEndpoint = errors.New("endpoint has not been provided")
)

func NewFetcher(p FetcherArgs) (*Fetcher, error) {
	switch {
	case p.AccessKey == "":
		log.Print("access key has not been provided, geo-ip metrics won't be announced")
	case p.Endpoint == "":
		return nil, ErrNoEndpoint
	}

	return &Fetcher{
		timeout:   p.Timeout,
		endpoint:  p.Endpoint,
		accessKey: p.AccessKey,
	}, nil
}

func (f *Fetcher) Fetch(ipAddr string) (Info, error) {
	var (
		result Info

		url = fmt.Sprintf("%s/%s?access_key=%s", f.endpoint, ipAddr, f.accessKey)
		cli = http.Client{
			Timeout: f.timeout,
		}
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, fmt.Errorf("can't prepare geo ip request: %w", err)
	}

	res, err := cli.Do(req)
	if err != nil {
		return result, fmt.Errorf("can't get geo ip response: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, fmt.Errorf("can't read geo ip response body: %w", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("can't parse JSON of geo ip response: %w", err)
	}

	return result, nil
}
