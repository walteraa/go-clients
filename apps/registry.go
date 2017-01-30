package apps

import (
	"io"
	"gopkg.in/h2non/gentleman.v1"
	"github.com/vtex/go-clients/clients"
	"fmt"
	"encoding/json"
	"strings"
)

// Registry is an interface for interacting with the registry
type Registry interface {
	GetApp(account string, id string) (*PublishedApp, error)
	ListFiles(account string, id string) (*FileList, error)
	GetFile(account string, id string, path string) (io.ReadCloser, error)
	GetFileB(account string, id string, path string) ([]byte, error)
	GetFileJ(account string, id string, path string, data interface{}) error
}

// Client is a struct that provides interaction with apps
type RegistryClient struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

// NewClient creates a new Registry client
func NewRegistryClient(endpoint, authToken, userAgent string, reqCtx clients.RequestContext) Registry {
	cl, vc := clients.CreateClient(endpoint, authToken, userAgent, reqCtx)
	return &RegistryClient{cl, vc}
}

const (
	metadataPath    = "/%v/master/registry/%v/%v"
	fileListPath    = "/%v/master/registry/%v/%v/files"
	fileContentPath = "/%v/master/registry/%v/%v/files/%v"
)

// GetApp returns the app metadata
func (cl *RegistryClient) GetApp(account string, id string) (*PublishedApp, error) {
	const kind = "app"

	segments, err := getSegments(id)
	if err != nil {
		return nil, err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataPath, account, segments[0], segments[1])).Send()
	if err != nil {
		return nil, err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, err
	} else if ok {
		return cached.(*PublishedApp), nil
	}

	var m PublishedApp
	if err := res.JSON(&m); err != nil {
		return nil, err
	}

	cl.cache.SetFor(kind, res, &m)

	return &m, nil
}

func (cl *RegistryClient) ListFiles(account string, id string) (*FileList, error) {
	const kind = "files"

	segments, err := getSegments(id)
	if err != nil {
		return nil, err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileListPath, account, segments[0], segments[1])).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, err
	} else if ok {
		return cached.(*FileList), nil
	}

	var l FileList
	if err := res.JSON(&l); err != nil {
		return nil, err
	}

	cl.cache.SetFor(kind, res, &l)

	return &l, nil
}

func (cl *RegistryClient) getFile(account string, id string, path string, useCache bool) (io.ReadCloser, error) {
	segments, err := getSegments(id)
	if err != nil {
		return nil, err
	}

	req := cl.http.Get().
		AddPath(fmt.Sprintf(fileContentPath, account, segments[0], segments[1], path))
	if useCache {
		req.UseRequest(clients.Cache)
	}

	return req.Send()
}
func (cl *RegistryClient) GetFile(account string, id string, path string) (io.ReadCloser, error) {
	res, err := cl.getFile(account, id, path, false)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cl *RegistryClient) GetFileB(account string, id string, path string) ([]byte, error) {
	const kind = "file-bytes"
	res, err := cl.getFile(account, id, path, true)
	if err != nil {
		return nil, err
	}

	gentRes := res.(*gentleman.Response)
	if cached, ok, err := cl.cache.GetFor(kind, gentRes); err != nil {
		return nil, err
	} else if ok {
		return cached.([]byte), nil
	}

	bytes := gentRes.Bytes()
	cl.cache.SetFor(kind, gentRes, bytes)

	return bytes, nil
}

func (cl *RegistryClient) GetFileJ(account string, id string, path string, data interface{}) error {
	b, err := cl.GetFileB(account, id, path)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, data)
}

func getSegments(id string) ([]string, error) {
	segments := strings.SplitN(id, "@", 2)
	if len(segments) != 2 {
		return nil, fmt.Errorf("Not a composed app identifier: %s", id)
	}
	return segments, nil
}
