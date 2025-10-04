package proxy

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

func HTTPSRecord(host string) (bool, time.Time) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		panic(err)
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), dns.TypeHTTPS)

	client := new(dns.Client)

	https := false
	var expires time.Time

	for _, server := range config.Servers {
		addr := server + ":" + config.Port
		response, _, err := client.Exchange(msg, addr)
		if err != nil {
			fmt.Println("Error querying", addr, ":", err)
			continue
		}

		if len(response.Answer) > 0 {
			for _, ans := range response.Answer {
				ttl := ans.Header().Ttl
				expires = time.Now().Add(time.Duration(ttl) * time.Second)
			}
			https = true
			break
		}
	}

	return https, expires
}
