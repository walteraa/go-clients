package clients

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
)

func TestCache(t *testing.T) {
	Convey("Given a context and a handler", t, func() {
		c := gentleman.NewContext()
		req, _ := http.NewRequest("GET", "http://my.site/foo", nil)
		c.SetRequest(req)

		h := new(FakeHandler)

		Convey("When I invoke the cache plugin", func() {
			Cache(c, h)

			Convey("Then Next() should be called", func() {
				So(h.InvokedNext, ShouldBeTrue)
			})

			Convey("And If-None-Match should not be set", func() {
				So(req.Header.Get("If-None-Match"), ShouldBeEmpty)
			})
		})

		Convey("With a cache storage", func() {
			storage := &FakeCache{Values: map[string]interface{}{}}
			c.Set(cacheStorageKey, storage)

			Convey("When I invoke the cache plugin", func() {
				Cache(c, h)

				Convey("Then If-None-Match should not be set", func() {
					So(req.Header.Get("If-None-Match"), ShouldBeEmpty)
				})
			})

			Convey("With a stored ETag", func() {
				storage.Values[eTagKey(req)] = "my-e-tag"

				Convey("When I invoke the cache plugin", func() {
					Cache(c, h)

					Convey("Then If-None-Match should be set with the ETag", func() {
						So(req.Header.Get("If-None-Match"), ShouldEqual, "my-e-tag")
					})
				})
			})
		})
	})
}

type FakeHandler struct {
	InvokedNext bool
}

func (h *FakeHandler) Next(*context.Context) {
	h.InvokedNext = true
}

func (h *FakeHandler) Stop(*context.Context)         {}
func (h *FakeHandler) Error(*context.Context, error) {}

type FakeCache struct {
	Values map[string]interface{}
}

func (c *FakeCache) Get(key string) (interface{}, bool) {
	value, ok := c.Values[key]
	return value, ok
}

func (c *FakeCache) Set(key string, value interface{}, ttl time.Duration) {
	c.Values[key] = value
}
