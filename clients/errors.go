package clients

import (
	"fmt"
	"net/http"
)

type ResponseError struct {
	Response   *http.Response
	StatusCode int
	Code       string
	Message    string
}

func (err ResponseError) Error() string {
	return fmt.Sprintf("(%d %v at %v) %v", err.StatusCode, err.Code, err.Response.Request.URL, err.Message)
}
