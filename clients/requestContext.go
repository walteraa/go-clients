package clients

import (
	"net/http"
	"sync"
)

type RequestContext interface {
	AddMetadata(keys []string)
	AddHeadersTo(w http.ResponseWriter)
}

func NewRequestContext() RequestContext {
	return &requestContext{metadata: []string{}}
}

type requestContext struct {
	metadata []string
	lock     sync.RWMutex
}

func (c *requestContext) AddMetadata(keys []string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.metadata = append(c.metadata, keys...)
}

func (c *requestContext) AddHeadersTo(w http.ResponseWriter) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	h := w.Header()
	for _, k := range c.metadata {
		h.Add(metadataHeader, k)
	}
}
