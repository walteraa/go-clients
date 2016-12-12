package colossus

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vtex/go-clients/errors"
)

var hcli = &http.Client{
	Timeout: time.Second * 10,
}

type Colossus interface {
	SendEventB(account, workspace, sender, destination string, body []byte) error
}

type Client struct {
	Endpoint  string
	AuthToken string
	UserAgent string
}

func NewClient(endpoint, authToken, userAgent string) Colossus {
	return &Client{Endpoint: endpoint, AuthToken: authToken, UserAgent: userAgent}
}

const (
	pathToEventSending = "/%v/%v/events/%v/%v"
)

func (cl *Client) createRequestB(method string, content []byte, pathFormat string, a ...interface{}) *http.Request {
	var body io.Reader
	if content != nil {
		body = bytes.NewBuffer(content)
	}
	return cl.createRequest(method, body, pathFormat, a...)
}

func (cl *Client) createRequest(method string, body io.Reader, pathFormat string, a ...interface{}) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf(cl.Endpoint+pathFormat, a...), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "token "+cl.AuthToken)
	req.Header.Set("User-Agent", cl.UserAgent)
	return req
}

func (cl *Client) SendEventB(account, workspace, sender, destination string, body []byte) error {
	req := cl.createRequestB("POST", body, pathToEventSending, account, workspace, sender, destination)
	req.Header.Set("Content-Type", "application/json")

	res, reserr := hcli.Do(req)
	if reserr != nil {
		return reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return err
	}

	return nil
}
