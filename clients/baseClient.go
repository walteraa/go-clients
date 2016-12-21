package clients

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"time"

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
			if err := statusCode(c.Response); err != nil {
				h.Error(c, err)
			} else {
				h.Next(c)
			}
		}))
}

func statusCode(res *http.Response) error {
	if 200 <= res.StatusCode && res.StatusCode < 300 {
		return nil
	}

	descr := parse(res)

	return ResponseError{
		Response:   res,
		StatusCode: res.StatusCode,
		Code:       descr.Code,
		Message:    descr.Message,
	}
}

func parse(res *http.Response) *ErrorDescriptor {
	var descr ErrorDescriptor
	var buf []byte
	var err error

	if buf, err = ioutil.ReadAll(res.Body); err != nil {
		descr = ErrorDescriptor{Code: "undefined"}
	} else if err = json.Unmarshal(buf, &descr); err == nil {
		descr = ErrorDescriptor{Code: "undefined", Message: string(buf)}
	}

	return &descr
}
