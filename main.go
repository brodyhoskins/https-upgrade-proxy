package main

import (
	"https-upgrade-proxy/proxy"
	"log"
	"net/http"
)

func main() {
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(proxy.ProxyHandler),
	}
	log.Println("Starting proxy on :8080")
	log.Fatal(server.ListenAndServe())
}
