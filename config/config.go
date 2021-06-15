package config

import (
	"github.com/go-srv/configreader"
)

type DNSServerConfig struct {
	BindAddr string `default:"0.0.0.0:53"`
}

type DNSResolverConfig struct {
	Nameservers     []string `default:"[\"1.1.1.1:53\", \"1.0.0.1:53\"]"`
	Interval        int      `default:"200"`
	Timeout         int      `default:"5"`
	TTL             uint32   `default:"600"`
	NXDomainOnBlock bool     `default:"false"`
	DoH             DoHConfig
	Hosts           HostsFileConfig
}

type DoHConfig struct {
	Enable   bool   `default:"false"`
	Endpoint string `default:"https://cloudflare-dns.com/dns-query"`
}

type APIServerConfig struct {
	Enable   bool   `default:"true"`
	BindAddr string `default:"127.0.0.1:8080"`
}

type HostsFileConfig struct {
	Enable          bool   `default:"true"`
	HostsFile       string `default:"/etc/hosts"`
	RefreshInterval uint32 `default:"900"`
}

type BlockerConfig struct {
	Enable     bool `default:"true"`
	SourceURLs []DNSBlockSource
	SourceDir  string `default:"sources"`
	Blocklist  []string
	Whitelist  []string `default:"[\"getsentry.com\",\"www.getsentry.com\"]"`
}

type DNSBlockSource struct {
	Name string
	URL  string
}

type Config struct {
	LogLevel  string `default:"Info"`
	DebugMode bool   `default:"true"`
	DNSServer DNSServerConfig
	Resolver  DNSResolverConfig
	APIServer APIServerConfig
	Blocker   BlockerConfig
	Hosts     HostsFileConfig
}

func LoadConfig(filepath string) (*Config, error) {

	var config Config

	err := configreader.ReadFromFile(filepath, &config)

	return &config, err
}
