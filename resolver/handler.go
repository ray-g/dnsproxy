package resolver

import (
	"net"
	"time"

	"github.com/miekg/dns"

	c "github.com/ray-g/dnsproxy/cache"
	r "github.com/ray-g/dnsproxy/cache/record"
	conf "github.com/ray-g/dnsproxy/config"
	h "github.com/ray-g/dnsproxy/hosts"
	"github.com/ray-g/dnsproxy/logger"
	"github.com/ray-g/dnsproxy/stats"
	"github.com/ray-g/dnsproxy/utils"
)

var (
	nullroute   = net.ParseIP("0.0.0.0")
	nullroutev6 = net.ParseIP("0:0:0:0:0:0:0:0")
)

// Question type
type Question struct {
	Qname  string `json:"name"`
	Qtype  string `json:"type"`
	Qclass string `json:"class"`
}

// String formats a question
func (q *Question) String() string {
	return q.Qname + " " + q.Qclass + " " + q.Qtype
}

// DNSHandler type
type DNSHandler struct {
	config   *conf.DNSResolverConfig
	resolver *Resolver
	cache    c.Cache
	hosts    *h.Hosts
}

// DNSOperationData type
type DNSOperationData struct {
	Net string
	w   dns.ResponseWriter
	req *dns.Msg
}

// NewHandler returns a new DNSHandler
func NewHandler(config *conf.DNSResolverConfig, cache c.Cache) *DNSHandler {
	var (
		clientConfig *dns.ClientConfig
		resolver     *Resolver
	)

	resolver = &Resolver{clientConfig}

	handler := &DNSHandler{
		resolver: resolver,
		cache:    cache,
		config:   config,
	}

	if config.Hosts.Enable {
		handler.hosts = h.NewHosts(&config.Hosts)
	}

	return handler
}

func (h *DNSHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	stats.AddQuery()

	q := req.Question[0]
	Q := Question{utils.UnFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}
	key := Q.Qname

	IPQuery := utils.IsIPQuery(q)

	// Only query cache when qtype == 'A'|'AAAA' , qclass == 'IN'
	if stats.Active() {
		if IPQuery > 0 {
			record, err := h.cache.Get(key)
			if err != nil {
				logger.Debugf("%s didn't hit cache", Q.String())
			} else {
				blocked := record.Blocked
				if !blocked {
					logger.Debugf("%s hit cache", Q.String())

					// we need this copy against concurrent modification of Id
					msg := *record.Msg
					msg.Id = req.Id
					h.WriteReplyMsg(w, &msg)
					return
				} else {
					logger.Debugf("%s hit cache and was blocked: forwarding request", Q.String())

					m := new(dns.Msg)
					m.SetReply(req)

					if h.config.NXDomainOnBlock {
						m.SetRcode(req, dns.RcodeNameError)
					} else {
						switch IPQuery {
						case utils.IPv4Query:
							rrHeader := dns.RR_Header{
								Name:   q.Name,
								Rrtype: dns.TypeA,
								Class:  dns.ClassINET,
								Ttl:    h.config.TTL,
							}
							a := &dns.A{Hdr: rrHeader, A: nullroute}
							m.Answer = append(m.Answer, a)
						case utils.IPv6Query:
							rrHeader := dns.RR_Header{
								Name:   q.Name,
								Rrtype: dns.TypeAAAA,
								Class:  dns.ClassINET,
								Ttl:    h.config.TTL,
							}
							a := &dns.AAAA{Hdr: rrHeader, AAAA: nullroutev6}
							m.Answer = append(m.Answer, a)
						}
					}

					h.WriteReplyMsg(w, m)

					stats.AddQueryBlocked()
					logger.Noticef("%s found in blocklist", Q.Qname)
				}
			}
		}

		// Query hosts
		if h.config.Hosts.Enable && IPQuery > 0 {
			if ips, ok := h.hosts.Get(Q.Qname, IPQuery); ok {
				mesg := new(dns.Msg)
				mesg.SetReply(req)

				switch IPQuery {
				case utils.IPv4Query:
					rr_header := dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    h.config.TTL,
					}
					for _, ip := range ips {
						a := &dns.A{
							Hdr: rr_header,
							A:   ip,
						}
						mesg.Answer = append(mesg.Answer, a)
					}
				case utils.IPv6Query:
					rr_header := dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    h.config.TTL,
					}
					for _, ip := range ips {
						aaaa := &dns.AAAA{
							Hdr:  rr_header,
							AAAA: ip,
						}
						mesg.Answer = append(mesg.Answer, aaaa)
					}
				}

				w.WriteMsg(mesg)

				ttl := time.Duration(h.config.TTL) * time.Second
				h.cache.Set(key, r.NewCustomRecord(mesg, ttl))
				logger.Debug("%s found in hosts file", Q.Qname)
				stats.AddCustomDomain()
				return
			} else {
				logger.Debug("%s didn't found in hosts file", Q.Qname)
			}
		}
	}

	// Resolve from upstream DNS servers
	mesg, err := h.resolver.Lookup(Net, req, h.config.Timeout, h.config.Interval, h.config.Nameservers, h.config.DoH.Enable, h.config.DoH.Endpoint)

	if err != nil {
		logger.Errorf("resolve query error %v", err)
		h.HandleFailed(w, req)

		return
	}

	if mesg.Truncated && Net == "udp" {
		mesg, err = h.resolver.Lookup("tcp", req, h.config.Timeout, h.config.Interval, h.config.Nameservers, h.config.DoH.Enable, h.config.DoH.Endpoint)
		if err != nil {
			logger.Errorf("resolve tcp query error %v", err)
			h.HandleFailed(w, req)

			return
		}
	}

	//find the smallest ttl
	ttl := time.Duration(h.config.TTL) * time.Second
	var candidateTTL time.Duration

	for _, answer := range mesg.Answer {
		candidateTTL = time.Duration(answer.Header().Ttl) * time.Second

		if candidateTTL > 0 && candidateTTL < ttl {
			ttl = candidateTTL
		}
	}

	h.WriteReplyMsg(w, mesg)

	if IPQuery > 0 && len(mesg.Answer) > 0 {
		err = h.cache.Set(key, r.NewResolvedRecord(mesg, ttl))
		if err != nil {
			logger.Errorf("set %s cache failed: %v", Q.String(), err)
		}
		logger.Debugf("insert %s into cache with ttl %ds", Q.String(), ttl/time.Second)
		stats.AddNormalDomain()
	}
}

// DoTCP begins a tcp query
func (h *DNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

// DoUDP begins a udp query
func (h *DNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}

// HandleFailed handles dns failures
func (h *DNSHandler) HandleFailed(w dns.ResponseWriter, message *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(message, dns.RcodeServerFailure)
	h.WriteReplyMsg(w, m)
}

// WriteReplyMsg writes the dns reply
func (h *DNSHandler) WriteReplyMsg(w dns.ResponseWriter, message *dns.Msg) {
	defer func() {
		if r := recover(); r != nil {
			logger.Noticef("Recovered in WriteReplyMsg: %s", r)
		}
	}()

	err := w.WriteMsg(message)
	if err != nil {
		logger.Error(err.Error())
	}
}
