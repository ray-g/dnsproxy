package hosts

import (
	"bufio"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"

	conf "github.com/ray-g/dnsproxy/config"
	"github.com/ray-g/dnsproxy/logger"
	"github.com/ray-g/dnsproxy/utils"
)

type Hosts struct {
	fileHosts       *FileHosts
	refreshInterval time.Duration
}

func NewHosts(hs *conf.HostsFileConfig) *Hosts {
	fileHosts := &FileHosts{
		file:  hs.HostsFile,
		hosts: make(map[string]string),
	}

	hosts := Hosts{fileHosts, time.Second * time.Duration(hs.RefreshInterval)}
	hosts.refresh()
	return &hosts
}

func (h *Hosts) Get(domain string, family int) ([]net.IP, bool) {

	var sips []string
	var ip net.IP
	var ips []net.IP

	sips, _ = h.fileHosts.Get(domain)

	if sips == nil {
		return nil, false
	}

	for _, sip := range sips {
		switch family {
		case utils.IPv4Query:
			ip = net.ParseIP(sip).To4()
		case utils.IPv6Query:
			ip = net.ParseIP(sip).To16()
		default:
			continue
		}
		if ip != nil {
			ips = append(ips, ip)
		}
	}

	return ips, (ips != nil)
}

/*
Update hosts records from /etc/hosts file and redis per minute
*/
func (h *Hosts) refresh() {
	ticker := time.NewTicker(h.refreshInterval)
	go func() {
		for {
			h.fileHosts.Refresh()
			<-ticker.C
		}
	}()
}

type FileHosts struct {
	file  string
	hosts map[string]string
	mu    sync.RWMutex
}

func (f *FileHosts) Get(domain string) ([]string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	domain = strings.ToLower(domain)
	ip, ok := f.hosts[domain]
	if ok {
		return []string{ip}, true
	}

	sld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, false
	}

	for host, ip := range f.hosts {
		if strings.HasPrefix(host, "*.") {
			old, err := publicsuffix.EffectiveTLDPlusOne(host)
			if err != nil {
				continue
			}
			if sld == old {
				return []string{ip}, true
			}
		}
	}

	return nil, false
}

func (f *FileHosts) Refresh() {
	buf, err := os.Open(f.file)
	if err != nil {
		logger.Warn("Update hosts records from file failed %s", err)
		return
	}
	defer buf.Close()

	f.mu.Lock()
	defer f.mu.Unlock()

	f.clear()

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "\t", " ", -1)

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		sli := strings.Split(line, " ")

		if len(sli) < 2 {
			continue
		}

		ip := sli[0]
		if !utils.IsIP(ip) {
			continue
		}

		// Would have multiple columns of domain in line.
		// Such as "127.0.0.1  localhost localhost.domain" on linux.
		// The domains may not strict standard, like "local" so don't check with f.isDomain(domain).
		for i := 1; i <= len(sli)-1; i++ {
			domain := strings.TrimSpace(sli[i])
			if domain == "" {
				continue
			}

			f.hosts[strings.ToLower(domain)] = ip
		}
	}
	logger.Debug("update hosts records from %s, total %d records.", f.file, len(f.hosts))
}

func (f *FileHosts) clear() {
	f.hosts = make(map[string]string)
}
