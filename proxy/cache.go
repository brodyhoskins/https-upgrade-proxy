package proxy

import (
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	expires time.Time
	host    string
	https   bool
}

var (
	cache   = make(map[string]*cacheEntry)
	cacheMu sync.RWMutex
)

func FetchFromCache(host string) *cacheEntry {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	if entry, ok := cache[strings.ToLower(host)]; ok && entry.expires.After(time.Now()) {
		return entry
	}

	return nil
}

func PushToCache(host string, https bool, expires time.Time) bool {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	cache[strings.ToLower(host)] = &cacheEntry{
		expires: expires,
		host:    host,
		https:   https,
	}
	return true
}
