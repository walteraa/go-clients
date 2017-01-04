package clients

import "time"

type CacheConfig struct {
	Storage        CacheStorage
	RequestContext RequestContext
	TTL            time.Duration
}

type CacheStorage interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}
