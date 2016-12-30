package clients

import (
	"fmt"
	"time"

	gentleman "gopkg.in/h2non/gentleman.v1"
)

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
