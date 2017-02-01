package clients

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugin"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"gopkg.in/h2non/gentleman.v1/plugins/timeout"
	"gopkg.in/h2non/gentleman.v1"
)

const cacheStorageKey = "cache-storage"

func CreateClient(endpoint, authToken, userAgent string, reqCtx RequestContext) (*gentleman.Client, ValueCache) {
	if reqCtx == nil {
		panic("reqCtx cannot be <nil>")
	}

	cl := gentleman.New().
		BaseURL(strings.TrimRight(endpoint, "/")).
		Use(timeout.Request(5 * time.Second)).
		Use(headers.Set("Authorization", "token "+authToken)).
		Use(headers.Set("User-Agent", userAgent)).
		Use(responseErrors()).
		Use(recordHeaders(reqCtx)).
		Use(traceRequest(reqCtx))

	var vc ValueCache
	if cache := reqCtx.getCache(); cache != nil {
		if cache.TTL <= 0 {
			panic("Cache TTL should be greater than zero")
		}
		if cache.Storage == nil {
			panic("Cache storage should not be <nil>")
		}

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

func recordHeaders(reqCtx RequestContext) plugin.Plugin {
	return plugin.NewResponsePlugin(func(c *context.Context, h context.Handler) {
		reqCtx.Parse(c.Response.Header)
		h.Next(c)
	})
}

func traceRequest(reqCtx RequestContext) plugin.Plugin {
	const startTime = "startTime"

	p := plugin.New()
	if reqCtx.isTraceEnabled() {
		p.SetHandler("request", func(c *context.Context, h context.Handler) {
			c.Request.Header.Set(enableTraceHeader, "true")
			c.Set(startTime, time.Now())
			h.Next(c)
		})
		p.SetHandler("response", func(c *context.Context, h context.Handler) {
			tree := newCallTree(c.Request, c.Response, c.Get(startTime).(time.Time))
			reqCtx.UpdateS(traceHeader, func(current string) string {
				var traces []*CallTree
				if err := json.Unmarshal([]byte(current), &traces); err != nil || current == "" {
					traces = []*CallTree{}
				}
				traces = append(traces, tree)

				js, _ := json.Marshal(traces)
				return string(js)
			})
			h.Next(c)
		})
	}
	return p
}

type CallTree struct {
	Call     string      `json:"call"`
	Status   int         `json:"status"`
	Time     int64       `json:"time"`
	Children []*CallTree `json:"children,omitempty"`
}

func newCallTree(req *http.Request, res *http.Response, start time.Time) *CallTree {
	resh := res.Header.Get(traceHeader)
	var children []*CallTree
	if err := json.Unmarshal([]byte(resh), &children); err != nil && resh != "" {
		logrus.WithError(err).Error("Failed to unmarshal call trace")
	}
	return &CallTree{
		Call:     req.Method + " " + req.URL.String(),
		Time:     time.Now().Sub(start).Nanoseconds() / int64(time.Millisecond),
		Status:   res.StatusCode,
		Children: children,
	}
}
