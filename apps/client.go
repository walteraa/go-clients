package apps

import (
	"fmt"
	"io"
	"strings"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

// Apps is an interface for interacting with apps
type Apps interface {
	GetApp(account, workspace, app string, context []string) (*Manifest, error)
	ListFiles(account, workspace, app string, context []string) (*FileList, error)
	GetFile(account, workspace, app string, context []string, path string) (io.ReadCloser, error)
	GetFileB(account, workspace, app string, context []string, path string) ([]byte, error)
	GetFileJ(account, workspace, app string, context []string, path string, dest interface{}) error
	GetDependencies(account, workspace string) (map[string][]string, error)
}

// Client is a struct that provides interaction with apps
type Client struct {
	http *gentleman.Client
}

// NewClient creates a new Apps client
func NewClient(endpoint, authToken, userAgent string) Apps {
	return &Client{clients.CreateClient(endpoint, authToken, userAgent)}
}

const (
	pathToDependencies = "/%v/%v/dependencies"
	pathToApp          = "/%v/%v/apps/%v"
	pathToFiles        = "/%v/%v/apps/%v/files"
	pathToFile         = "/%v/%v/apps/%v/files/%v"
)

// GetApp describes an installed app's manifest
func (cl *Client) GetApp(account, workspace, app string, context []string) (*Manifest, error) {
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToApp, account, workspace, app)).
		SetQuery("context", strings.Join(context, "/")).Send()
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := res.JSON(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (cl *Client) ListFiles(account, workspace, app string, context []string) (*FileList, error) {
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToFiles, account, workspace, app)).
		SetQuery("context", strings.Join(context, "/")).Send()
	if err != nil {
		return nil, err
	}

	var files FileList
	if err := res.JSON(&files); err != nil {
		return nil, err
	}
	return &files, nil
}

// GetFile gets an installed app's file as read closer
func (cl *Client) GetFile(account, workspace, app string, context []string, path string) (io.ReadCloser, error) {
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToFile, account, workspace, app, path)).
		SetQuery("context", strings.Join(context, "/")).Send()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetFileB gets an installed app's file as bytes
func (cl *Client) GetFileB(account, workspace, app string, context []string, path string) ([]byte, error) {
	res, err := cl.GetFile(account, workspace, app, context, path)
	if err != nil {
		return nil, err
	}

	return res.(*gentleman.Response).Bytes(), nil
}

// GetFileJ gets an installed app's file as deserialized JSON object
func (cl *Client) GetFileJ(account, workspace, app string, context []string, path string, dest interface{}) error {
	res, err := cl.GetFile(account, workspace, app, context, path)
	if err != nil {
		return err
	}

	if err := res.(*gentleman.Response).JSON(&dest); err != nil {
		return err
	}

	return nil
}

func (cl *Client) GetDependencies(account, workspace string) (map[string][]string, error) {
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToDependencies, account, workspace)).Send()
	if err != nil {
		return nil, err
	}

	var dependencies map[string][]string
	err = res.JSON(&dependencies)
	return dependencies, err
}
