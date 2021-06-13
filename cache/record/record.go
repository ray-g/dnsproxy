package record

import (
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

func NewBlockedRecord(msg *dns.Msg) *Record {
	return NewRecord(msg, true, true, 0)
}
