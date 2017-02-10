package colossus

import (
	"bytes"
	"fmt"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

type Colossus interface {
	SendEventJ(account, workspace, sender, subject, key string, body interface{}) error
	SendEventB(account, workspace, sender, subject, key string, body []byte) error
	SendLogJ(account, workspace, sender, subject, level string, body interface{}) error
	SendLogB(account, workspace, sender, subject, level string, body []byte) error
}

type Client struct {
	http *gentleman.Client
}

func NewClient(endpoint, authToken, userAgent string, ttl int, reqCtx clients.RequestContext) Colossus {
	cl, _ := clients.CreateClient(endpoint, authToken, userAgent, reqCtx, ttl)
	return &Client{cl}
}

const (
	eventPath = "/%v/%v/events/%v/%v/%v"
	logPath   = "/%v/%v/logs/%v/%v/%v"
)

func (cl *Client) SendEventJ(account, workspace, sender, subject, key string, body interface{}) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(eventPath, account, workspace, sender, subject, key)).
		JSON(body).Send()

	return err
}

func (cl *Client) SendEventB(account, workspace, sender, subject, key string, body []byte) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(eventPath, account, workspace, sender, subject, key)).
		Body(bytes.NewReader(body)).Send()

	return err
}

func (cl *Client) SendLogJ(account, workspace, sender, subject, level string, body interface{}) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(logPath, account, workspace, sender, subject, level)).
		JSON(body).Send()

	return err
}

func (cl *Client) SendLogB(account, workspace, sender, subject, level string, body []byte) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(logPath, account, workspace, sender, subject, level)).
		Body(bytes.NewReader(body)).Send()

	return err
}
