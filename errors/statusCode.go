package errors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func StatusCode(res *http.Response) error {
	if 200 <= res.StatusCode && res.StatusCode < 300 {
		return nil
	}

	descr := parse(res)

	switch res.StatusCode {
	case 404:
		return NotFoundError{
			URL:       res.Request.URL,
			ErrorCode: descr.Code,
			Message:   descr.Message,
		}
	}

	return StatusCodeError{
		StatusCode: res.StatusCode,
		URL:        res.Request.URL,
		ErrorCode:  descr.Code,
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

type StatusCodeError struct {
	StatusCode int
	URL        *url.URL
	ErrorCode  string
	Message    string
}

func (err StatusCodeError) Error() string {
	return fmt.Sprintf("(%d %v at %v) %v", err.StatusCode, err.ErrorCode, err.URL, err.Message)
}
