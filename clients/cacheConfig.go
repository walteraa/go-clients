package clients

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	gentleman "gopkg.in/h2non/gentleman.v1"
)

const metadataHeader = "x-vtex-meta"

type CacheConfig struct {
	Storage        CacheStorage
	RequestContext RequestContext
	TTL            time.Duration
}

type CacheStorage interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

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

type ValueCache interface {
	GetFor(res *gentleman.Response) (interface{}, error)
	SetFor(res *gentleman.Response, value interface{}) error
}

type valueCache struct {
	storage CacheStorage
	ttl     time.Duration
}

func (c *valueCache) GetFor(res *gentleman.Response) (interface{}, error) {
	eTag := res.Header.Get("ETag")
	if eTag == "" {
		return nil, fmt.Errorf("ETag header not found in response")
	}

	fromCache, ok := c.storage.Get(fmt.Sprintf("cached-response:%v:%v", res.RawRequest.URL.String(), eTag))
	if !ok {
		return nil, fmt.Errorf("Value not found in cache: " + res.RawRequest.URL.String())
	}

	return fromCache, nil
}

func (c *valueCache) SetFor(res *gentleman.Response, value interface{}) error {
	eTag := res.Header.Get("ETag")
	if eTag == "" {
		return fmt.Errorf("ETag header not found in response")
	}

	c.storage.Set(fmt.Sprintf("cached-response:%v:%v", res.RawRequest.URL.String(), eTag), value, c.ttl)

	return nil
}
