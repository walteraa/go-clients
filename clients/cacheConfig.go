package clients

import "time"

const metadataHeader = "X-Vtex-Meta"

type CacheConfig struct {
	Storage        CacheStorage
	RequestContext RequestContext
	TTL            time.Duration
}

type CacheStorage interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}
