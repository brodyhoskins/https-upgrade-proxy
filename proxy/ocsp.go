package proxy

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ocspHosts contains OCSP hosts that aren't ocsp.<TLD>
var ocspHosts = []string{
	"ocsp.int-x3.letsencrypt.org",
}

func OCSP(host string) bool {
	if strings.HasPrefix(host, "ocsp.") {
		ps, ok := publicsuffix.PublicSuffix(host)
		if ok && strings.HasSuffix(host, ps) {
			return true
		}
	}

	for _, h := range ocspHosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return true
		}
	}

	return false
}
