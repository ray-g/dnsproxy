package record

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

// Record
type Record struct {
	Msg      *dns.Msg
	Blocked  bool      `json:"blocked"`
	NoExpire bool      `json:"no_expire"`
	UpdateAt time.Time `json:"update_at"`
	ExpireAt time.Time `json:"expire_at"`
}

func (r *Record) Expired() bool {
	if r.NoExpire {
		return false
	}

	if r.ExpireAt.Before(time.Now()) {
		return true
	}

	return false
}

func NewRecord(msg *dns.Msg, blocked bool, noexpire bool, ttl time.Duration) *Record {
	now := time.Now()
	return &Record{
		Msg:      msg,
		Blocked:  blocked,
		NoExpire: noexpire,
		UpdateAt: now,
		ExpireAt: now.Add(ttl),
	}
}

func NewResolvedRecord(msg *dns.Msg, ttl time.Duration) *Record {
	return NewRecord(msg, false, false, ttl)
}

func NewCustomRecord(msg *dns.Msg, ttl time.Duration) *Record {
	return NewRecord(msg, false, false, ttl)
}

func NewBlockedRecord() *Record {
	return NewRecord(bMesg, true, true, 0)
}

func blockedMesg() *dns.Msg {
	m := new(dns.Msg)
	rrAHeader := dns.RR_Header{
		Name:   "domain.blocked",
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    600,
	}
	a := &dns.A{Hdr: rrAHeader, A: net.ParseIP("0.0.0.0")}
	m.Answer = append(m.Answer, a)

	return m
}

var bMesg *dns.Msg

func init() {
	bMesg = blockedMesg()
}
