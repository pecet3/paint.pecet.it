package wardrouter

import (
	"net"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	var rawIP string
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		rawIP = strings.TrimSpace(ips[0])
	}

	if rawIP == "" {
		rawIP = strings.TrimSpace(r.Header.Get("X-Real-IP"))
	}

	if rawIP == "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			rawIP = strings.TrimSpace(r.RemoteAddr)
		} else {
			rawIP = ip
		}
	}

	parsedIP := net.ParseIP(rawIP)
	if parsedIP != nil {

		if ip4 := parsedIP.To4(); ip4 != nil {
			return ip4.String()
		}
		return parsedIP.String()
	}

	return rawIP
}
