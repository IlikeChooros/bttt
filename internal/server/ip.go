package server

import (
	"log/slog"
	"net"
	"net/http"
)

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func GetRequestIPAddress(req *http.Request) string {
	headers := []string{
		"x-client-ip", // Standard headers used by Amazon EC2, Heroku, and others.
		// "x-forwarded-for",     // Load-balancers (AWS ELB) or proxies.
		"cf-connecting-ip",    // @see https://support.cloudflare.com/hc/en-us/articles/200170986-How-does-Cloudflare-handle-HTTP-Request-headers-
		"fastly-client-ip",    // Fastly and Firebase hosting header (When forwarded to cloud function)
		"true-client-ip",      // Akamai and Cloudflare: True-Client-IP.
		"x-real-ip",           // Default nginx proxy/fcgi; alternative to x-forwarded-for, used by some proxies.
		"x-cluster-client-ip", // (Rackspace LB and Riverbed's Stingray) http://www.rackspace.com/knowledge_center/article/controlling-access-to-linux-cloud-sites-based-on-the-client-ip-address
		"x-forwarded",
		"forwarded-for",
		"forwarded",
	}

	for _, header := range headers {
		if value := req.Header.Get(header); value != "" {
			if ip := net.ParseIP(value); ip != nil {
				return value
			}
		}
	}

	return "unknown"
}
