package geoip

import (
	"sync"
)

type CachedFetcher struct {
	mu sync.Mutex

	fetcher *Fetcher
	cache   map[string]Info
}

func NewCachedFetcher(p FetcherArgs) (*CachedFetcher, error) {
	fetcher, err := NewFetcher(p)
	if err != nil {
		return nil, err
	}

	return &CachedFetcher{
		fetcher: fetcher,
		cache:   make(map[string]Info),
	}, nil
}

func (f *CachedFetcher) Fetch(ipAddr string) (Info, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	info, ok := f.cache[ipAddr]
	if ok {
		return info, nil
	}

	info, err := f.fetcher.Fetch(ipAddr)
	if err != nil {
		return info, err
	}

	f.cache[ipAddr] = info

	return info, nil
}
