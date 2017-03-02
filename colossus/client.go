package colossus

import (
	"bytes"
	"fmt"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

type Colossus interface {
	SendEventJ(sender, subject, key string, body interface{}) error
	SendEventB(sender, subject, key string, body []byte) error
	SendLogJ(sender, subject, level string, body interface{}) error
	SendLogB(sender, subject, level string, body []byte) error
}

type Client struct {
	http *gentleman.Client
}

func NewClient(config *clients.Config) Colossus {
	cl := clients.CreateClient("colossus", config, true)
	return &Client{cl}
}

const (
	eventPath = "/events/%v/%v/%v"
	logPath   = "/logs/%v/%v/%v"
)

func (cl *Client) SendEventJ(sender, subject, key string, body interface{}) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(eventPath, sender, subject, key)).
		JSON(body).Send()

	return err
}

func (cl *Client) SendEventB(sender, subject, key string, body []byte) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(eventPath, sender, subject, key)).
		Body(bytes.NewReader(body)).Send()

	return err
}

func (cl *Client) SendLogJ(sender, subject, level string, body interface{}) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(logPath, sender, subject, level)).
		JSON(body).Send()

	return err
}

func (cl *Client) SendLogB(sender, subject, level string, body []byte) error {
	_, err := cl.http.Post().
		AddPath(fmt.Sprintf(logPath, sender, subject, level)).
		Body(bytes.NewReader(body)).Send()

	return err
}
