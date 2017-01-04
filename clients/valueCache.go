package clients

import (
	"fmt"
	"time"

	gentleman "gopkg.in/h2non/gentleman.v1"
)

type ValueCache interface {
	GetFor(kind string, res *gentleman.Response) (interface{}, bool, error)
	SetFor(kind string, res *gentleman.Response, value interface{}) error
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
		return nil, false, fmt.Errorf("ETag header not found in response")
	}

	fromCache, ok := c.storage.Get(fmt.Sprintf("cached-response:%v:%v:%v", kind, res.RawRequest.URL.String(), eTag))
	if !ok {
		return nil, false, fmt.Errorf("Value not found in cache: " + res.RawRequest.URL.String())
	}

	return fromCache, true, nil
}

func (c *valueCache) SetFor(kind string, res *gentleman.Response, value interface{}) error {
	eTag := res.Header.Get("ETag")
	if eTag == "" {
		return fmt.Errorf("ETag header not found in response")
	}

	c.storage.Set(fmt.Sprintf("cached-response:%v:%v:%v", kind, res.RawRequest.URL.String(), eTag), value, c.ttl)

	return nil
}

type noOpValueCache struct{}

func (c *noOpValueCache) GetFor(kind string, res *gentleman.Response) (interface{}, bool, error) {
	if res.StatusCode == 304 {
		return nil, false, fmt.Errorf("Unable to handle 304 response. No cache storage.")
	}

	return nil, false, nil
}

func (c *noOpValueCache) SetFor(kind string, res *gentleman.Response, value interface{}) error {
	return nil
}
