package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugin"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"gopkg.in/h2non/gentleman.v1/plugins/timeout"
)

const (
	cacheStorageKey = "cache-storage"

	HeaderETag = "ETag"
)

type Config struct {
	Account        string
	Workspace      string
	Region         string
	Endpoint       string
	AuthToken      string
	UserAgent      string
	RequestContext RequestContext
	TTL            int
}

func CreateClient(service string, config *Config) (*gentleman.Client, ValueCache) {
	if config == nil {
		panic("config cannot be <nil>")
	}

	if config.RequestContext == nil {
		panic("config.RequestContext cannot be <nil>")
	}

	ttl := config.TTL
	if ttl <= 0 {
		ttl = 5
	}

	cl := gentleman.New().
		BaseURL(baseURL(service, config)).
		Use(timeout.Request(time.Duration(ttl) * time.Second)).
		Use(headers.Set("Authorization", "token "+config.AuthToken)).
		Use(headers.Set("User-Agent", config.UserAgent)).
		Use(responseErrors()).
		Use(recordHeaders(config.RequestContext)).
		Use(traceRequest(config.RequestContext))

	var vc ValueCache
	if cache := config.RequestContext.getCache(); cache != nil {
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
		p.SetHandler("before dial", func(c *context.Context, h context.Handler) {
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

func baseURL(service string, config *Config) string {
	endpoint := config.Endpoint
	if endpoint != "" {
		endpoint = "http://" + strings.TrimRight(endpoint, "/")
	} else {
		endpoint = fmt.Sprintf("http://%s.%s.vtex.io", service, config.Region)
	}

	if config.Account != "" && config.Workspace != "" {
		return endpoint + "/" + config.Account + "/" + config.Workspace
	}

	return endpoint
}
