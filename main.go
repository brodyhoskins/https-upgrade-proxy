package main

import (
	"flag"
	"fmt"
	"https-upgrade-proxy/proxy"
	"log"
	"net/http"
	"os"
)

func main() {
	egressIPv4 := flag.String("egress-ipv4", "", "Local IPv4 address for outbound connections")
	egressIPv6 := flag.String("egress-ipv6", "", "Local IPv6 address for outbound connections")

	listenIPv4 := flag.String("listen-ipv4", "", "IPv4 address to listen on (optional)")
	listenIPv6 := flag.String("listen-ipv6", "", "IPv6 address to listen on (optional)")

	port := flag.Int("port", 0, "Port to listen on (required)")
	flag.Parse()

	if *port == 0 {
		fmt.Fprintln(os.Stderr, "Port is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	proxy.SetOutboundBindIPs(*egressIPv4, *egressIPv6)

	listeners := []string{}

	if *listenIPv6 != "" {
		listeners = append(listeners, fmt.Sprintf("[%s]:%d", *listenIPv6, *port))
	} else if *listenIPv4 == "" {
		listeners = append(listeners, fmt.Sprintf("[::]:%d", *port))
	}

	if *listenIPv4 != "" {
		listeners = append(listeners, fmt.Sprintf("%s:%d", *listenIPv4, *port))
	}

	if len(listeners) == 0 {
		log.Fatal("No listener address could be determined")
	}

	displayEgressIPv4 := *egressIPv4
	displayEgressIPv6 := *egressIPv6
	if displayEgressIPv4 == "" {
		displayEgressIPv4 = "(default)"
	}
	if displayEgressIPv6 == "" {
		displayEgressIPv6 = "(default)"
	}

	for _, addr := range listeners {
		server := &http.Server{
			Addr:    addr,
			Handler: http.HandlerFunc(proxy.ProxyHandler),
		}

		go func(a string) {
			log.Printf("Listening on %s (egress IPv4: %s, egress IPv6: %s)", a, displayEgressIPv4, displayEgressIPv6)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server on %s: %v", a, err)
			}
		}(addr)
	}

	select {}
}
