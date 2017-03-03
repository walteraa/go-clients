package apps

import (
	"fmt"
	"strings"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

// Registry is an interface for interacting with the registry
type Registry interface {
	GetApp(id string) (*PublishedApp, string, error)
	ListFiles(id string) (*FileList, string, error)
	GetFile(id string, path string) (*gentleman.Response, string, error)
}

// Client is a struct that provides interaction with apps
type RegistryClient struct {
	http *gentleman.Client
}

// NewClient creates a new Registry client
func NewRegistryClient(config *clients.Config) Registry {
	config.Workspace = "master"
	cl := clients.CreateClient("apps", config, true)
	return &RegistryClient{cl}
}

const (
	metadataPath    = "/registry/%v/%v"
	fileListPath    = "/registry/%v/%v/files"
	fileContentPath = "/registry/%v/%v/files/%v"
)

// GetApp returns the app metadata
func (cl *RegistryClient) GetApp(id string) (*PublishedApp, string, error) {

	segments, err := getSegments(id)
	if err != nil {
		return nil, "", err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataPath, segments[0], segments[1])).Send()
	if err != nil {
		return nil, "", err
	}

	var m PublishedApp
	if err := res.JSON(&m); err != nil {
		return nil, "", err
	}

	return &m, res.Header.Get(clients.HeaderETag), nil
}

func (cl *RegistryClient) ListFiles(id string) (*FileList, string, error) {
	segments, err := getSegments(id)
	if err != nil {
		return nil, "", err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileListPath, segments[0], segments[1])).Send()
	if err != nil {
		return nil, "", err
	}

	var l FileList
	if err := res.JSON(&l); err != nil {
		return nil, "", err
	}

	return &l, res.Header.Get(clients.HeaderETag), nil
}

func (cl *RegistryClient) GetFile(id string, path string) (*gentleman.Response, string, error) {
	segments, err := getSegments(id)
	if err != nil {
		return nil, "", err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileContentPath, segments[0], segments[1], path)).Send()
	if err != nil {
		return nil, "", err
	}

	return res, res.Header.Get(clients.HeaderETag), nil
}

func getSegments(id string) ([]string, error) {
	segments := strings.SplitN(id, "@", 2)
	if len(segments) != 2 {
		return nil, fmt.Errorf("Not a composed app identifier: %s", id)
	}
	return segments, nil
}
