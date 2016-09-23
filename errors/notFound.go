package errors

import (
	"fmt"
	"net/url"
)

type NotFoundError struct {
	URL       *url.URL
	ErrorCode string
	Message   string
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("(%v at %v) %v", err.ErrorCode, err.URL, err.Message)
}
