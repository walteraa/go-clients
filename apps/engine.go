package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

// Apps is an interface for interacting with apps
type Apps interface {
	GetApp(account, workspace, app string, context []string) (*ActiveApp, string, error)
	ListFiles(account, workspace, app string, context []string) (*FileList, string, error)
	GetFile(account, workspace, app string, context []string, path string) (io.ReadCloser, string, error)
	GetFileB(account, workspace, app string, context []string, path string) ([]byte, string, error)
	GetFileJ(account, workspace, app string, context []string, path string, dest interface{}) (string, error)
	GetDependencies(account, workspace string) (map[string][]string, string, error)
}

// Client is a struct that provides interaction with apps
type AppsClient struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

// NewClient creates a new Apps client
func NewAppsClient(endpoint, authToken, userAgent string, ttl int, reqCtx clients.RequestContext) Apps {
	cl, vc := clients.CreateClient(endpoint, authToken, userAgent, reqCtx, ttl)
	return &AppsClient{cl, vc}
}

const (
	pathToDependencies = "/%v/%v/dependencies"
	pathToApp          = "/%v/%v/apps/%v"
	pathToFiles        = "/%v/%v/apps/%v/files"
	pathToFile         = "/%v/%v/apps/%v/files/%v"
)

// GetApp describes an installed app's manifest
func (cl *AppsClient) GetApp(account, workspace, app string, context []string) (*ActiveApp, string, error) {
	const kind = "manifest"
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToApp, account, workspace, app)).
		SetQuery("context", strings.Join(context, "/")).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*ActiveApp), res.Header.Get(clients.HeaderETag), nil
	}

	var manifest ActiveApp
	if err := res.JSON(&manifest); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &manifest)
	return &manifest, res.Header.Get(clients.HeaderETag), nil
}

func (cl *AppsClient) ListFiles(account, workspace, app string, context []string) (*FileList, string, error) {
	const kind = "file-list"
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToFiles, account, workspace, app)).
		SetQuery("context", strings.Join(context, "/")).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*FileList), res.Header.Get(clients.HeaderETag), nil
	}

	var files FileList
	if err := res.JSON(&files); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &files)

	return &files, res.Header.Get(clients.HeaderETag), nil
}

func (cl *AppsClient) getFile(account, workspace, app string, context []string, path string, useCache bool) (io.ReadCloser, error) {
	req := cl.http.Get().AddPath(fmt.Sprintf(pathToFile, account, workspace, app, path)).
		SetQuery("context", strings.Join(context, "/"))
	if useCache {
		req.UseRequest(clients.Cache)
	}

	return req.Send()
}

// GetFile gets an installed app's file as read closer
func (cl *AppsClient) GetFile(account, workspace, app string, context []string, path string) (io.ReadCloser, string, error) {
	res, err := cl.getFile(account, workspace, app, context, path, false)
	if err != nil {
		return nil, "", err
	}

	return res, res.(*gentleman.Response).Header.Get(clients.HeaderETag), nil
}

// GetFileB gets an installed app's file as bytes
func (cl *AppsClient) GetFileB(account, workspace, app string, context []string, path string) ([]byte, string, error) {
	const kind = "file-bytes"
	res, err := cl.getFile(account, workspace, app, context, path, true)
	if err != nil {
		return nil, "", err
	}

	gentRes := res.(*gentleman.Response)
	if cached, ok, err := cl.cache.GetFor(kind, gentRes); err != nil {
		return nil, "", err
	} else if ok {
		return cached.([]byte), gentRes.Header.Get(clients.HeaderETag), nil
	}

	bytes := gentRes.Bytes()
	cl.cache.SetFor(kind, gentRes, bytes)

	return bytes, gentRes.Header.Get(clients.HeaderETag), nil
}

// GetFileJ gets an installed app's file as deserialized JSON object
func (cl *AppsClient) GetFileJ(account, workspace, app string, context []string, path string, dest interface{}) (string, error) {
	b, eTag, err := cl.GetFileB(account, workspace, app, context, path)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, dest)
	return eTag, err
}

func (cl *AppsClient) GetDependencies(account, workspace string) (map[string][]string, string, error) {
	const kind = "dependencies"
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToDependencies, account, workspace)).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(map[string][]string), res.Header.Get(clients.HeaderETag), nil
	}

	var dependencies map[string][]string
	if err := res.JSON(&dependencies); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, dependencies)
	return dependencies, res.Header.Get(clients.HeaderETag), err
}
