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
	"gopkg.in/h2non/gentleman.v1/plugins/auth"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"gopkg.in/h2non/gentleman.v1/plugins/timeout"
	"gopkg.in/h2non/gentleman.v1/plugins/transport"
)

const HeaderETag = "ETag"

type Config struct {
	Account        string
	Workspace      string
	Region         string
	Endpoint       string
	AuthToken      string
	AuthFunc       func() string
	UserAgent      string
	RequestContext RequestContext
	Timeout        time.Duration
	Transport      http.RoundTripper
}

func CreateClient(service string, config *Config, workspaceBound bool) *gentleman.Client {
	if config == nil {
		panic("config cannot be <nil>")
	}

	if config.RequestContext == nil {
		panic("config.RequestContext cannot be <nil>")
	}

	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}

	cl := gentleman.New().
		Use(timeout.Request(config.Timeout)).
		Use(headers.Set("User-Agent", config.UserAgent)).
		Use(responseErrors()).
		Use(recordHeaders(config.RequestContext)).
		Use(traceRequest(config.RequestContext))

	if url := endpoint(service, config); url != "" {
		cl = cl.BaseURL(url)
	}

	if path := basePath(config, workspaceBound); path != "" {
		cl = cl.Path(path)
	}

	if config.AuthToken != "" {
		cl = cl.Use(auth.Bearer(config.AuthToken))
	} else if config.AuthFunc != nil {
		cl = cl.UseRequest(func(ctx *context.Context, h context.Handler) {
			ctx.Request.Header.Set("Authorization", "Bearer "+config.AuthFunc())
			h.Next(ctx)
		})
	}

	if config.Transport != nil {
		cl = cl.Use(transport.Set(config.Transport))
	}

	return cl
}

func responseErrors() plugin.Plugin {
	return plugin.NewResponsePlugin(func(c *context.Context, h context.Handler) {
		if http.StatusOK <= c.Response.StatusCode && c.Response.StatusCode < http.StatusNotFound {
			h.Next(c)
			return
		}

		var descr ErrorDescriptor
		var buf []byte
		var err error

		if buf, err = ioutil.ReadAll(c.Response.Body); err != nil {
			descr = ErrorDescriptor{Code: "undefined"}
		} else if err = json.Unmarshal(buf, &descr); err != nil || descr.Code == "" || descr.Message == "" {
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
	Cache    string      `json:"cache"`
	Time     int64       `json:"time"`
	Children []*CallTree `json:"children,omitempty"`
}

func newCallTree(req *http.Request, res *http.Response, start time.Time) *CallTree {
	resh := res.Header.Get(traceHeader)
	var children []*CallTree
	if err := json.Unmarshal([]byte(resh), &children); err != nil && resh != "" {
		logrus.WithError(err).Error("Failed to unmarshal call trace")
	}

	cache := "miss"
	if _, ok := res.Header["X-From-Cache"]; ok {
		cache = "hit"
	}

	return &CallTree{
		Call:     req.Method + " " + req.URL.String(),
		Time:     time.Now().Sub(start).Nanoseconds() / int64(time.Millisecond),
		Status:   res.StatusCode,
		Cache:    cache,
		Children: children,
	}
}

func endpoint(service string, config *Config) string {
	if config.Endpoint != "" {
		return "http://" + strings.TrimRight(config.Endpoint, "/")
	} else if service != "" {
		return fmt.Sprintf("http://%s.%s.vtex.io", service, config.Region)
	} else {
		return ""
	}
}

func basePath(config *Config, workspaceBound bool) string {
	if workspaceBound {
		return "/" + config.Account + "/" + config.Workspace
	}

	return ""
}
