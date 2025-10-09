package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type checkResult struct {
	useHTTPS   bool
	expiration time.Time
}

func ProxyHandler(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	host = strings.ToLower(host)
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	fmt.Print(time.Now().Format("2006-01-02 15:04:05"), " ", host)

	if request.Method == http.MethodGet {
		if CaptivePortal(host) || OCSP(host) {
			pass(writer, request)
			return
		}

		ctx, cancel := context.WithTimeout(request.Context(), 5*time.Second)
		defer cancel()

		results := make(chan checkResult, 4)

		go func() {
			cacheEntry := FetchFromCache(host)
			if cacheEntry != nil && cacheEntry.https {
				results <- checkResult{true, cacheEntry.expires} // lowercase
			} else {
				results <- checkResult{false, time.Time{}}
			}
		}()

		go func() {
			hstsPreload, expiresAt := HSTSPreload(host)
			results <- checkResult{hstsPreload, expiresAt}
		}()

		go func() {
			httpsRecord, expiresAt := HTTPSRecord(host)
			results <- checkResult{httpsRecord, expiresAt}
		}()

		go func() {
			hstsHeader, expiresAt := HSTSHeader(host) // Only 2 values
			results <- checkResult{hstsHeader, expiresAt}
		}()

		var chosen checkResult
		timeout := time.After(4 * time.Second)
		for i := 0; i < 4; i++ {
			select {
			case r := <-results:
				if r.useHTTPS {
					chosen = r
					cancel() // cancel other checks
					goto Redirect
				}
			case <-timeout:
				break
			case <-ctx.Done():
				break
			}
		}

		PushToCache(host, false, time.Now().Add(5*time.Minute))
		pass(writer, request)
		return

	Redirect:
		PushToCache(host, true, chosen.expiration)
		redirect(writer, request, chosen.expiration)
		return
	}

	pass(writer, request)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func pass(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(": Use HTTP")

	outReq, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
	if err != nil {
		http.Error(writer, "Proxy error: "+err.Error(), http.StatusBadRequest)
		return
	}
	outReq.Header = request.Header.Clone()

	resp, err := http.DefaultClient.Do(outReq)
	if err != nil {
		http.Error(writer, "Proxy error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(writer.Header(), resp.Header)
	writer.WriteHeader(resp.StatusCode)
	io.Copy(writer, resp.Body)
}

func redirect(writer http.ResponseWriter, request *http.Request, expiresAt time.Time) {
	fmt.Println(": Use HTTPS")

	target := "https://" + request.Host + request.URL.RequestURI()

	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	writer.Header().Set("Expires", expiresAt.UTC().Format(http.TimeFormat))

	http.Redirect(writer, request, target, http.StatusMovedPermanently)
}
