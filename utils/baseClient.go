package utils

import (
	"strings"

	"time"

	"github.com/vtex/go-clients/errors"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugin"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"gopkg.in/h2non/gentleman.v1/plugins/timeout"
)

func CreateClient(endpoint, authToken, userAgent string) *gentleman.Client {
	return gentleman.New().
		BaseURL(strings.TrimRight(endpoint, "/")).
		Use(timeout.Request(5 * time.Second)).
		Use(headers.Set("Authorization", "token "+authToken)).
		Use(headers.Set("User-Agent", userAgent)).
		Use(plugin.NewResponsePlugin(func(c *context.Context, h context.Handler) {
			if err := errors.StatusCode(c.Response); err != nil {
				h.Error(c, err)
			} else {
				h.Next(c)
			}
		}))
}
