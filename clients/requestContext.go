package clients

import (
	"net/http"
	"sync"
)

const (
	metadataHeader    = "X-Vtex-Meta"
	enableTraceHeader = "X-Vtex-Trace-Enable"
	traceHeader       = "X-Call-Trace"
)

type RequestContext interface {
	Parse(h http.Header)
	Write(w http.ResponseWriter)
	UpdateS(header string, update func(current string) string)
	getCache() *CacheConfig
	isTraceEnabled() bool
}

func NewRequestContext(cache *CacheConfig, parent *http.Request) RequestContext {
	enableTrace := parent.Header.Get(enableTraceHeader) == "true"
	headers := map[string][]string{}
	if enableTrace {
		headers[enableTraceHeader] = []string{"true"}
	}
	return &requestContext{
		headers:           headers,
		cache:             cache,
		enableTraceHeader: enableTrace,
	}
}

type requestContext struct {
	sync.RWMutex

	cache             *CacheConfig
	headers           http.Header
	enableTraceHeader bool
}

// Parse parses an incoming response in order to accumulate headers
func (c *requestContext) Parse(h http.Header) {
	c.Lock()
	defer c.Unlock()

	for _, h := range h[metadataHeader] {
		c.headers.Add(metadataHeader, h)
	}
}

// Write writes accumulated headers to an outgoing response
func (c *requestContext) Write(w http.ResponseWriter) {
	c.RLock()
	defer c.RUnlock()

	headers := w.Header()
	for h, v := range c.headers {
		headers[h] = v
	}
}

// UpdateS updates the value of a header for the outgoing response
func (c *requestContext) UpdateS(header string, update func(current string) string) {
	c.Lock()
	defer c.Unlock()

	c.headers.Set(header, update(c.headers.Get(header)))
}

func (c *requestContext) getCache() *CacheConfig {
	return c.cache
}

func (c *requestContext) isTraceEnabled() bool {
	return c.enableTraceHeader
}
