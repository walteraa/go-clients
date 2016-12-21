package clients

import (
	"fmt"
	"net/http"
)

type StatusCodeError struct {
	Response *http.Response
	Code     string
	Message  string
}

func (err StatusCodeError) Error() string {
	return fmt.Sprintf("(%d %v at %v) %v", err.Response.StatusCode, err.Code, err.Response.Request.URL, err.Message)
}
