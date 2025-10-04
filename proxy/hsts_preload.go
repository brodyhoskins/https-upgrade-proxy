package proxy

import (
	"log"
	"time"

	"github.com/chromium/hstspreload/chromium/preloadlist"
)

func HSTSPreload(host string) (bool, time.Time) {
	pl, err := preloadlist.NewFromLatest()
	if err != nil {
		log.Fatalf("Failed to load HSTS preload list: %v", err)
	}
	idx := pl.Index()

	entry, _ := idx.Get(host)

	return entry.Mode == "force-https", time.Now().AddDate(1, 0, 0)
}
