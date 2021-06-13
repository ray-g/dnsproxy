package utils

import (
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/miekg/dns"
)

const (
	NotIPQuery = 0
	IPv4Query  = 4
	IPv6Query  = 6
)

func IsIPQuery(q dns.Question) int {
	if q.Qclass != dns.ClassINET {
		return NotIPQuery
	}

	switch q.Qtype {
	case dns.TypeA:
		return IPv4Query
	case dns.TypeAAAA:
		return IPv6Query
	default:
		return NotIPQuery
	}
}

// UnFqdn function
func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}

func IsDomain(domain string) bool {
	if IsIP(domain) {
		return false
	}
	match, _ := regexp.MatchString(`^([a-zA-Z0-9\*]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$`, domain)
	return match
}

func IsIP(ip string) bool {
	return (net.ParseIP(ip) != nil)
}

func EnsureDirectory(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if os.MkdirAll(path, os.ModePerm) != nil {
			return fmt.Errorf("failed to create folders: %s", path)
		}
	}

	if err == nil && !info.IsDir() {
		return fmt.Errorf("%s exists but not a folder", path)
	}

	return nil
}
