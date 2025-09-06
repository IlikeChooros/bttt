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

func GetRequestIPAddress(req *http.Request) (string, bool) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		slog.Error(err.Error())
		return "unknown", false
	}
	return ip, true
}
