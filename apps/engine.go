package apps

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/context"
	"gopkg.in/h2non/gentleman.v1/plugin"
)

// Apps is an interface for interacting with apps
type Apps interface {
	GetApp(app, parentID string) (*ActiveApp, string, error)
	ListFiles(app, parentID string) (*FileList, string, error)
	GetFile(app, parentID string, path string) (io.ReadCloser, string, error)
	GetFileB(app, parentID string, path string) ([]byte, string, error)
	GetFileJ(app, parentID string, path string, dest interface{}) (string, error)
	GetDependencies() (map[string][]string, string, error)
}

// Client is a struct that provides interaction with apps
type AppsClient struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

// NewClient creates a new Apps client
func NewAppsClient(config *clients.Config) Apps {
	cl, vc := clients.CreateClient("apps", config)
	return &AppsClient{cl, vc}
}

const (
	pathToDependencies = "/dependencies"
	pathToApp          = "/apps/%v"
	pathToFiles        = "/apps/%v/files"
	pathToFile         = "/apps/%v/files/%v"
)

// GetApp describes an installed app's manifest
func (cl *AppsClient) GetApp(app, parentID string) (*ActiveApp, string, error) {
	const kind = "manifest"
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToApp, app)).
		Use(addParent(parentID)).
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

func (cl *AppsClient) ListFiles(app, parentID string) (*FileList, string, error) {
	const kind = "file-list"
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToFiles, app)).
		Use(addParent(parentID)).
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

func (cl *AppsClient) getFile(app, parentID string, path string, useCache bool) (io.ReadCloser, error) {
	req := cl.http.Get().AddPath(fmt.Sprintf(pathToFile, app, path)).
		Use(addParent(parentID))
	if useCache {
		req.UseRequest(clients.Cache)
	}

	return req.Send()
}

// GetFile gets an installed app's file as read closer
func (cl *AppsClient) GetFile(app, parentID string, path string) (io.ReadCloser, string, error) {
	res, err := cl.getFile(app, parentID, path, false)
	if err != nil {
		return nil, "", err
	}

	return res, res.(*gentleman.Response).Header.Get(clients.HeaderETag), nil
}

// GetFileB gets an installed app's file as bytes
func (cl *AppsClient) GetFileB(app, parentID string, path string) ([]byte, string, error) {
	const kind = "file-bytes"
	res, err := cl.getFile(app, parentID, path, true)
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
func (cl *AppsClient) GetFileJ(app, parentID string, path string, dest interface{}) (string, error) {
	b, eTag, err := cl.GetFileB(app, parentID, path)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, dest)
	return eTag, err
}

func (cl *AppsClient) GetDependencies() (map[string][]string, string, error) {
	const kind = "dependencies"
	res, err := cl.http.Get().AddPath(pathToDependencies).
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

func addParent(parentID string) plugin.Plugin {
	return plugin.NewRequestPlugin(func(ctx *context.Context, h context.Handler) {
		if parentID != "" {
			query := ctx.Request.URL.Query()
			query.Set("parent", parentID)
			ctx.Request.URL.RawQuery = query.Encode()
		}
		h.Next(ctx)
	})
}
