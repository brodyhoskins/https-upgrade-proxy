package proxy

import (
	"strings"
)

var captivePortalHosts = []string{
	"captive.apple.com",
	"connectivitycheck.gstatic.com",
	"detectportal.firefox.com",
	"nmcheck.gnome.org",
	"www.msftconnecttest.com",
}

func CaptivePortal(host string) bool {
	host = strings.ToLower(host)

	for _, h := range captivePortalHosts {
		if host == h {
			return true
		}
	}

	return false
}
