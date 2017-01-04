package clients

import (
	"net/http"
	"sync"
)

type RequestContext interface {
	Parse(h http.Header)
	Write(w http.ResponseWriter)
	getCache() *CacheConfig
}

func NewRequestContext(cache *CacheConfig) RequestContext {
	return &requestContext{
		trackedKeys: []string{"X-Vtex-Meta"},
		headers:     http.Header{},
		cache:       cache,
	}
}

type requestContext struct {
	trackedKeys []string
	headers     http.Header
	lock        sync.RWMutex

	cache *CacheConfig
}

func (c *requestContext) Parse(h http.Header) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, k := range c.trackedKeys {
		for _, v := range h[k] {
			c.headers.Add(k, v)
		}
	}
}

func (c *requestContext) Write(w http.ResponseWriter) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	h := w.Header()
	for k, v := range c.headers {
		h[k] = v
	}
}

func (c *requestContext) getCache() *CacheConfig {
	return c.cache
}
