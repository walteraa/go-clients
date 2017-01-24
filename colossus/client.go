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
}

type Client struct {
	http *gentleman.Client
}

func NewClient(endpoint, authToken, userAgent string, reqCtx clients.RequestContext) Colossus {
	cl, _ := clients.CreateClient(endpoint, authToken, userAgent, reqCtx)
	return &Client{cl}
}

const (
	eventPath = "/%v/%v/events/%v/%v/%v"
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
