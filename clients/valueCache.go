package clients

import (
	"fmt"
	"time"

	gentleman "gopkg.in/h2non/gentleman.v1"
)

type ValueCache interface {
	GetFor(kind string, res *gentleman.Response) (interface{}, bool, error)
	SetFor(kind string, res *gentleman.Response, value interface{})
}

type valueCache struct {
	storage CacheStorage
	ttl     time.Duration
}

func (c *valueCache) GetFor(kind string, res *gentleman.Response) (interface{}, bool, error) {
	if res.StatusCode != 304 {
		return nil, false, nil
	}

	eTag := res.Header.Get("ETag")
	if eTag == "" {
		return nil, false, fmt.Errorf("(get) ETag header not found in response for " + res.RawRequest.URL.String())
	}

	fromCache, ok := c.storage.Get(fmt.Sprintf("cached-response:%v:%v:%v", kind, res.RawRequest.URL.String(), eTag))
	if !ok {
		return nil, false, fmt.Errorf("Value not found in cache: " + res.RawRequest.URL.String())
	}

	return fromCache, true, nil
}

func (c *valueCache) SetFor(kind string, res *gentleman.Response, value interface{}) {
	eTag := res.Header.Get("ETag")
	if eTag != "" {
		c.storage.Set(fmt.Sprintf("cached-response:%v:%v:%v", kind, res.RawRequest.URL.String(), eTag), value, c.ttl)
		c.storage.Set(eTagKey(res.RawRequest), eTag, c.ttl)
	}

}

type noOpValueCache struct{}

func (c *noOpValueCache) GetFor(kind string, res *gentleman.Response) (interface{}, bool, error) {
	if res.StatusCode == 304 {
		return nil, false, fmt.Errorf("Unable to handle 304 response. No cache storage.")
	}

	return nil, false, nil
}

func (c *noOpValueCache) SetFor(kind string, res *gentleman.Response, value interface{}) {
	return
}
