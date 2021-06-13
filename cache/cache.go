package cache

import (
	"errors"

	r "github.com/ray-g/dnsproxy/cache/record"
)

var (
	ErrorCacheKeyMissed  = errors.New("Key missed")
	ErrorCacheKeyExpired = errors.New("Key expired")
	ErrorCacheFull       = errors.New("Cache full")
)

// Cache interface
type Cache interface {
	Get(key string) (record *r.Record, err error)
	Set(key string, record *r.Record) error
	Exists(key string) bool
	Remove(key string)
	Length() int
	Dump() string
}
