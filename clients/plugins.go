package clients

import (
	"net/http"

	"gopkg.in/h2non/gentleman.v1/context"
)

// Cache plugin inserts an If-None-Match header if the request URL has a known ETag
func Cache(c *context.Context, h context.Handler) {
	if c.Request.Method != "" && c.Request.Method != "GET" {
	} else if ctxStorage, ok := c.GetOk(cacheStorageKey); !ok {
	} else if storage, ok := ctxStorage.(CacheStorage); !ok {
	} else if eTag, ok := storage.Get(eTagKey(c.Request)); ok {
		c.Request.Header.Add("If-None-Match", eTag.(string))
	}
	h.Next(c)
}

func eTagKey(req *http.Request) string {
	return "cached-etag:" + req.URL.String()
}
