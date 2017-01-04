package clients

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"time"

	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugin"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"gopkg.in/h2non/gentleman.v1/plugins/timeout"
)

const cacheStorageKey = "cache-storage"

func CreateClient(endpoint, authToken, userAgent string, reqCtx RequestContext) (*gentleman.Client, ValueCache) {
	cl := gentleman.New().
		BaseURL(strings.TrimRight(endpoint, "/")).
		Use(timeout.Request(5 * time.Second)).
		Use(headers.Set("Authorization", "token "+authToken)).
		Use(headers.Set("User-Agent", userAgent)).
		Use(responseErrors()).
		Use(recordHeaders(reqCtx))

	var vc ValueCache
	if cache := reqCtx.getCache(); cache != nil {
		if cache.TTL <= 0 {
			panic("Cache TTL should be greater than zero")
		}
		if cache.Storage == nil {
			panic("Cache storage should not be <nil>")
		}

		cl = cl.
			Use(addETag(cache.Storage)).
			Use(storeETag(cache.Storage, cache.TTL))
		cl.Context.Set(cacheStorageKey, cache.Storage)

		vc = &valueCache{
			storage: cache.Storage,
			ttl:     cache.TTL + 30*time.Second, // values should be cached for a little longer than e-tags
		}
	} else {
		vc = &noOpValueCache{}
	}

	return cl, vc
}

func addETag(storage CacheStorage) plugin.Plugin {
	if storage == nil {
		return plugin.New()
	}

	return plugin.NewPhasePlugin("before dial", func(c *context.Context, h context.Handler) {
		if c.Request.Method == "" || c.Request.Method == "GET" {
			if eTag, ok := storage.Get("cached-etag:" + c.Request.URL.String()); ok {
				c.Request.Header.Add("If-None-Match", eTag.(string))
			}
		}
		h.Next(c)
	})
}

func responseErrors() plugin.Plugin {
	return plugin.NewResponsePlugin(func(c *context.Context, h context.Handler) {
		if 200 <= c.Response.StatusCode && c.Response.StatusCode < 400 {
			h.Next(c)
			return
		}

		var descr ErrorDescriptor
		var buf []byte
		var err error

		if buf, err = ioutil.ReadAll(c.Response.Body); err != nil {
			descr = ErrorDescriptor{Code: "undefined"}
		} else if err = json.Unmarshal(buf, &descr); err == nil {
			descr = ErrorDescriptor{Code: "undefined", Message: string(buf)}
		}

		h.Error(c, ResponseError{
			Response:   c.Response,
			StatusCode: c.Response.StatusCode,
			Code:       descr.Code,
			Message:    descr.Message,
		})
	})
}

func storeETag(storage CacheStorage, ttl time.Duration) plugin.Plugin {
	if storage == nil {
		return plugin.New()
	}

	return plugin.NewResponsePlugin(func(c *context.Context, h context.Handler) {
		eTag := c.Response.Header.Get("ETag")
		if eTag != "" {
			storage.Set("cached-etag:"+c.Request.URL.String(), eTag, ttl)
		}
		h.Next(c)
	})
}

func recordHeaders(reqCtx RequestContext) plugin.Plugin {
	if reqCtx == nil {
		return plugin.New()
	}

	return plugin.NewResponsePlugin(func(c *context.Context, h context.Handler) {
		reqCtx.Parse(c.Response.Header)
		h.Next(c)
	})
}
