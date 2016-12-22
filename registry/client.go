package registry

import (
	"fmt"

	gentleman "gopkg.in/h2non/gentleman.v1"

	"io"

	"github.com/vtex/apps-utils/appidentifier"
	"github.com/vtex/go-clients/clients"
)

// Registry is an interface for interacting with the registry
type Registry interface {
	GetApp(account string, id string) (*Manifest, error)
	ListIdentities(account string, id string, acceptRange bool) (*IdentityListResponse, error)
	ListFiles(account string, id string) (*FileList, error)
	GetFile(account string, id string, path string) (io.ReadCloser, error)
	GetFileB(account string, id string, path string) ([]byte, error)
	GetFileJ(account string, id string, path string, data interface{}) error
}

// Client is a struct that provides interaction with apps
type Client struct {
	http *gentleman.Client
}

// NewClient creates a new Registry client
func NewClient(endpoint, authToken, userAgent string) Registry {
	return &Client{clients.CreateClient(endpoint, authToken, userAgent)}
}

const (
	metadataPath    = "/%v/master/registry/%v/%v"
	identityPath    = "/%v/master/registry/%v/%v/identity"
	fileListPath    = "/%v/master/registry/%v/%v/files"
	fileContentPath = "/%v/master/registry/%v/%v/files/%v"
)

// GetApp returns the app metadata
func (cl *Client) GetApp(account string, id string) (*Manifest, error) {
	compID, err := parseComposedID(id)
	if err != nil {
		return nil, err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataPath, account, compID.Prefix(), compID.Suffix())).Send()
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := res.JSON(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (cl *Client) ListIdentities(account string, id string, acceptRange bool) (*IdentityListResponse, error) {
	compID, err := parseComposedID(id)
	if err != nil {
		return nil, err
	}

	req := cl.http.Get().
		AddPath(fmt.Sprintf(identityPath, account, compID.Prefix(), compID.Suffix()))
	if acceptRange {
		req = req.SetQuery("acceptRange", "true")
	}
	res, err := req.Send()
	if err != nil {
		return nil, err
	}

	var l IdentityListResponse
	if err := res.JSON(&l); err != nil {
		return nil, err
	}

	return &l, nil
}

func (cl *Client) ListFiles(account string, id string) (*FileList, error) {
	compID, err := parseComposedID(id)
	if err != nil {
		return nil, err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileListPath, account, compID.Prefix(), compID.Suffix())).Send()
	if err != nil {
		return nil, err
	}

	var l FileList
	if err := res.JSON(&l); err != nil {
		return nil, err
	}

	return &l, nil
}

func (cl *Client) GetFile(account string, id string, path string) (io.ReadCloser, error) {
	compID, err := parseComposedID(id)
	if err != nil {
		return nil, err
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileContentPath, account, compID.Prefix(), compID.Suffix(), path)).Send()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cl *Client) GetFileB(account string, id string, path string) ([]byte, error) {
	res, err := cl.GetFile(account, id, path)
	if err != nil {
		return nil, err
	}

	return res.(*gentleman.Response).Bytes(), nil
}

func (cl *Client) GetFileJ(account string, id string, path string, data interface{}) error {
	res, err := cl.GetFile(account, id, path)
	if err != nil {
		return err
	}

	return res.(*gentleman.Response).JSON(data)
}

func parseComposedID(id string) (appidentifier.ComposedIdentifier, error) {
	appID, err := appidentifier.ParseAppIdentifier(id)
	if err != nil {
		return nil, err
	}

	if compID, ok := appID.(appidentifier.ComposedIdentifier); ok {
		return compID, nil
	}
	return nil, fmt.Errorf("Not a composed app identifier: " + id)
}
