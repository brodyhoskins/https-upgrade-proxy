package proxy

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

func HSTSHeader(host string) (enabled bool, expiration time.Time) {
	parent, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		parent = host
	}

	hosts := []string{host}
	if parent != host {
		hosts = append(hosts, parent)
	}

	methods := []string{"HEAD", "GET"}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	type result struct {
		hsts   string
		host   string
		method string
		err    error
	}

	results := make(chan result, len(hosts)*len(methods))

	for _, h := range hosts {
		url := "https://" + h + "/"
		for _, method := range methods {
			go func(url, host, method string) {
				req, _ := http.NewRequestWithContext(ctx, method, url, nil)
				resp, err := client.Do(req)
				if err != nil {
					results <- result{"", host, method, err}
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode >= 400 {
					results <- result{"", host, method, nil}
					return
				}
				hsts := resp.Header.Get("Strict-Transport-Security")
				results <- result{hsts, host, method, nil}
			}(url, h, method)
		}
	}

	var hsts string

	for i := 0; i < len(hosts)*len(methods); i++ {
		select {
		case r := <-results:
			if r.err == nil && r.hsts != "" {
				hsts = r.hsts
				cancel()
				break
			}
		case <-ctx.Done():
			break
		}
		if hsts != "" {
			break
		}
	}

	if hsts == "" {
		return false, time.Now().AddDate(1, 0, 0)
	}

	var maxAgeSeconds int64
	for _, part := range strings.Split(hsts, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "max-age=") {
			val := strings.TrimPrefix(part, "max-age=")
			n, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				maxAgeSeconds = n
				break
			}
		}
	}

	expirationTime := time.Now().AddDate(1, 0, 0)
	if maxAgeSeconds > 0 {
		expirationTime = time.Now().Add(time.Duration(maxAgeSeconds) * time.Second)
	}

	return true, expirationTime
}
