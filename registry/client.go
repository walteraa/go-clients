package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/vtex/apps-utils/appidentifier"
	"github.com/vtex/apps-utils/metadata"
	"github.com/vtex/go-clients/errors"
)

var hcli = &http.Client{
	Timeout: time.Second * 10,
}

// Registry is an interface for interacting with the registry
type Registry interface {
	GetAppMetadata(account string, id appidentifier.ComposedIdentifier) (*metadata.AppMetadata, error)
}

// Client is a struct that provides interaction with apps
type Client struct {
	Endpoint  string
	AuthToken string
	UserAgent string
}

// NewClient creates a new Registry client
func NewClient(endpoint, authToken, userAgent string) Registry {
	return &Client{Endpoint: endpoint, AuthToken: authToken, UserAgent: userAgent}
}

const (
	pathToAppMetadata = "/%v/master/registry/%v/%v"
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

// GetAppMetadata returns the app metadata
func (cl *Client) GetAppMetadata(account string, id appidentifier.ComposedIdentifier) (*metadata.AppMetadata, error) {
	req := cl.createRequest("GET", nil, pathToAppMetadata, account, id.Prefix(), id.Suffix())
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return nil, reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return nil, err
	}

	var m metadata.AppMetadata
	buf, buferr := ioutil.ReadAll(res.Body)
	if buferr != nil {
		return nil, buferr
	}

	jsonerr := json.Unmarshal(buf, &m)
	if jsonerr != nil {
		return nil, jsonerr
	}

	return &m, nil
}
