package apps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/vtex/go-clients/errors"
)

var hcli = &http.Client{
	Timeout: time.Second * 10,
}

// Apps is an interface for interacting with apps
type Apps interface {
	GetApp(account, workspace, app string, context []string) (*Manifest, error)
	GetFile(account, workspace, app string, context []string, path string) (io.ReadCloser, error)
	GetFileB(account, workspace, app string, context []string, path string) ([]byte, error)
	GetFileJ(account, workspace, app string, context []string, path string, dest interface{}) error
}

// Client is a struct that provides interaction with apps
type Client struct {
	Endpoint  string
	AuthToken string
	UserAgent string
}

// NewClient creates a new Apps client
func NewClient(endpoint, authToken, userAgent string) Apps {
	return &Client{Endpoint: endpoint, AuthToken: authToken, UserAgent: userAgent}
}

const (
	pathToApp  = "/%v/%v/apps/%v?context=%v"
	pathToFile = "/%v/%v/apps/%v/files/%v?context=%v"
)

func (cl *Client) createRequest(method string, content []byte, pathFormat string, a ...interface{}) *http.Request {
	var body io.Reader
	if content != nil {
		body = bytes.NewBuffer(content)
	}
	req, _ := http.NewRequest(method, fmt.Sprintf(cl.Endpoint+pathFormat, a...), body)
	req.Header.Set("Authorization", "token "+cl.AuthToken)
	req.Header.Set("User-Agent", cl.UserAgent)
	return req
}

// GetApp describes an installed app's manifest
func (cl *Client) GetApp(account, workspace, app string, context []string) (*Manifest, error) {
	req := cl.createRequest("GET", nil, pathToApp, account, workspace, app, strings.Join(context, "/"))
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return nil, reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return nil, err
	}

	var manifest Manifest
	buf, buferr := ioutil.ReadAll(res.Body)
	if buferr != nil {
		return nil, buferr
	}

	jsonerr := json.Unmarshal(buf, &manifest)
	if jsonerr != nil {
		return nil, jsonerr
	}

	return &manifest, nil
}

// GetFile gets an installed app's file as read closer
func (cl *Client) GetFile(account, workspace, app string, context []string, path string) (io.ReadCloser, error) {
	req := cl.createRequest("GET", nil, pathToFile, account, workspace, app, path, strings.Join(context, "/"))
	res, resErr := hcli.Do(req)
	if resErr != nil {
		return nil, resErr
	}
	if err := errors.StatusCode(res); err != nil {
		return nil, err
	}
	return res.Body, nil
}

// GetFileB gets an installed app's file as bytes
func (cl *Client) GetFileB(account, workspace, app string, context []string, path string) ([]byte, error) {
	file, err := cl.GetFile(account, workspace, app, context, path)
	if err != nil {
		return nil, err
	}

	buf, bufErr := ioutil.ReadAll(file)
	if bufErr != nil {
		return nil, bufErr
	}
	return buf, nil
}

// GetFileJ gets an installed app's file as deserialized JSON object
func (cl *Client) GetFileJ(account, workspace, app string, context []string, path string, dest interface{}) error {
	buf, err := cl.GetFileB(account, workspace, app, context, path)
	if err != nil {
		return err
	}

	jsonErr := json.Unmarshal(buf, &dest)
	if jsonErr != nil {
		return jsonErr
	}

	return nil
}
