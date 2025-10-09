// helpers.go
package proxy

import (
	"context"
	"net"
	"net/http"
	"time"
)

var httpClient = &http.Client{}

func SetOutboundBindIPs(ipv4, ipv6 string) {
	var v4Addr, v6Addr *net.TCPAddr

	if ipv4 != "" {
		v4Addr = &net.TCPAddr{IP: net.ParseIP(ipv4)}
	}
	if ipv6 != "" {
		v6Addr = &net.TCPAddr{IP: net.ParseIP(ipv6)}
	}

	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		if network == "tcp6" && v6Addr != nil {
			return (&net.Dialer{
				LocalAddr: v6Addr,
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, network, address)
		}
		if network == "tcp4" && v4Addr != nil {
			return (&net.Dialer{
				LocalAddr: v4Addr,
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, network, address)
		}

		return new(net.Dialer).DialContext(ctx, network, address)
	}

	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext:         dialContext,
			ForceAttemptHTTP2:   true,
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func GetHTTPClient() *http.Client {
	return httpClient
}
