package utils

import "net"

func IsPrivateIP(ip string) bool {
	p := net.ParseIP(ip)
	return p != nil && (p.IsLoopback() || p.IsPrivate())
}
