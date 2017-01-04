package clients

import (
	"net/http"
	"sync"
)

const (
	metadataHeader = "X-Vtex-Meta"
)

type HeaderTracker interface {
	Parse(h http.Header)
	Write(w http.ResponseWriter)
}

type RequestContext interface {
	HeaderTracker
	getCache() *CacheConfig
}

func NewRequestContext(cache *CacheConfig) RequestContext {
	return &requestContext{
		trackers: []HeaderTracker{newMetadataForwarder()},
		cache:    cache,
	}
}

type requestContext struct {
	trackers []HeaderTracker
	cache    *CacheConfig
}

func (c *requestContext) Parse(h http.Header) {
	for _, t := range c.trackers {
		t.Parse(h)
	}
}

func (c *requestContext) Write(w http.ResponseWriter) {
	for _, t := range c.trackers {
		t.Write(w)
	}
}

func (c *requestContext) getCache() *CacheConfig {
	return c.cache
}

type metadataForwarder struct {
	metadata []string
	lock     sync.RWMutex
}

func newMetadataForwarder() *metadataForwarder {
	return &metadataForwarder{[]string{}, sync.RWMutex{}}
}

func (f *metadataForwarder) Parse(h http.Header) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.metadata = append(f.metadata, h[metadataHeader]...)
}

func (f *metadataForwarder) Write(w http.ResponseWriter) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	w.Header()[metadataHeader] = f.metadata
}
