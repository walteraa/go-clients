package registry

import (
	"fmt"

	"github.com/vtex/apps-utils/appidentifier"
	"github.com/vtex/apps-utils/metadata"
	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

// Registry is an interface for interacting with the registry
type Registry interface {
	GetAppMetadata(account string, id appidentifier.ComposedIdentifier) (*metadata.AppMetadata, error)
	ListIdentities(account string, id appidentifier.ComposedIdentifier, acceptRange bool) (*IdentityListResponse, error)
	ListAppFiles(account string, id appidentifier.ComposedIdentifier) (*FileListResponse, error)
	GetAppFileB(account string, id appidentifier.ComposedIdentifier, path string) ([]byte, error)
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

// GetAppMetadata returns the app metadata
func (cl *Client) GetAppMetadata(account string, id appidentifier.ComposedIdentifier) (*metadata.AppMetadata, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataPath, account, id.Prefix(), id.Suffix())).Send()
	if err != nil {
		return nil, err
	}

	var m metadata.AppMetadata
	if err := res.JSON(m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (cl *Client) ListIdentities(account string, id appidentifier.ComposedIdentifier, acceptRange bool) (*IdentityListResponse, error) {
	req := cl.http.Get().
		AddPath(fmt.Sprintf(identityPath, account, id.Prefix(), id.Suffix()))
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

func (cl *Client) ListAppFiles(account string, id appidentifier.ComposedIdentifier) (*FileListResponse, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileListPath, account, id.Prefix(), id.Suffix())).Send()
	if err != nil {
		return nil, err
	}

	var l FileListResponse
	if err := res.JSON(&l); err != nil {
		return nil, err
	}

	return &l, nil
}

func (cl *Client) GetAppFileB(account string, id appidentifier.ComposedIdentifier, path string) ([]byte, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(fileContentPath, account, id.Prefix(), id.Suffix(), path)).Send()
	if err != nil {
		return nil, err
	}

	return res.Bytes(), nil
}
