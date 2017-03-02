package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

// Registry is an interface for interacting with the registry
type Registry interface {
	GetApp(id string) (*PublishedApp, string, error)
	ListFiles(id string) (*FileList, string, error)
	GetFile(id string, path string) (io.ReadCloser, string, error)
	GetFileB(id string, path string) ([]byte, string, error)
	GetFileJ(id string, path string, data interface{}) (string, error)
}

// Client is a struct that provides interaction with apps
type RegistryClient struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

// NewClient creates a new Registry client
func NewRegistryClient(config *clients.Config) Registry {
	config.Workspace = "master"
	cl, vc := clients.CreateClient("apps", config, true)
	return &RegistryClient{cl, vc}
}

const (
	metadataPath    = "/registry/%v/%v"
	fileListPath    = "/registry/%v/%v/files"
	fileContentPath = "/registry/%v/%v/files/%v"
)

// GetApp returns the app metadata
func (cl *RegistryClient) GetApp(id string) (*PublishedApp, string, error) {
	const kind = "app"

	segments, err := getSegments(id)
	if err != nil {
		return nil, "", err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataPath, segments[0], segments[1])).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*PublishedApp), res.Header.Get(clients.HeaderETag), nil
	}

	var m PublishedApp
	if err := res.JSON(&m); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &m)

	return &m, res.Header.Get(clients.HeaderETag), nil
}

func (cl *RegistryClient) ListFiles(id string) (*FileList, string, error) {
	const kind = "files"

	segments, err := getSegments(id)
	if err != nil {
		return nil, "", err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileListPath, segments[0], segments[1])).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*FileList), res.Header.Get(clients.HeaderETag), nil
	}

	var l FileList
	if err := res.JSON(&l); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &l)

	return &l, res.Header.Get(clients.HeaderETag), nil
}

func (cl *RegistryClient) getFile(id string, path string, useCache bool) (io.ReadCloser, error) {
	segments, err := getSegments(id)
	if err != nil {
		return nil, err
	}

	req := cl.http.Get().
		AddPath(fmt.Sprintf(fileContentPath, segments[0], segments[1], path))
	if useCache {
		req.UseRequest(clients.Cache)
	}

	return req.Send()
}

func (cl *RegistryClient) GetFile(id string, path string) (io.ReadCloser, string, error) {
	res, err := cl.getFile(id, path, false)
	if err != nil {
		return nil, "", err
	}

	return res, res.(*gentleman.Response).Header.Get(clients.HeaderETag), nil
}

func (cl *RegistryClient) GetFileB(id string, path string) ([]byte, string, error) {
	const kind = "file-bytes"
	res, err := cl.getFile(id, path, true)
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

func (cl *RegistryClient) GetFileJ(id string, path string, data interface{}) (string, error) {
	b, eTag, err := cl.GetFileB(id, path)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, data)
	return eTag, err
}

func getSegments(id string) ([]string, error) {
	segments := strings.SplitN(id, "@", 2)
	if len(segments) != 2 {
		return nil, fmt.Errorf("Not a composed app identifier: %s", id)
	}
	return segments, nil
}
