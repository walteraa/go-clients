package vbase

import (
	"bytes"
	"fmt"
	"io"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
)

// Workspaces is an interface for interacting with workspaces
type VBase interface {
	GetBucket(account, workspace, bucket string) (*BucketResponse, error)
	SetBucketState(account, workspace, bucket, state string) error
	GetFile(account, workspace, bucket, path string) (io.ReadCloser, error)
	GetFileB(account, workspace, bucket, path string) ([]byte, error)
	GetFileConflict(account, workspace, bucket, path string) (io.ReadCloser, *Conflict, error)
	GetFileConflictB(account, workspace, bucket, path string) ([]byte, *Conflict, error)
	SaveFile(account, workspace, bucket, path string, body io.Reader) error
	SaveFileB(account, workspace, bucket, path string, content []byte, contentType string, unzip bool) error
	ListFiles(account, workspace, bucket, prefix, marker string, size int) (*FileListResponse, error)
	ListAllFiles(account, workspace, bucket, prefix string, size int) (*FileListResponse, error)
	DeleteFile(account, workspace, bucket, path string) error
}

// Client is a struct that provides interaction with workspaces
type Client struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

// NewClient creates a new Workspaces client
func NewClient(endpoint, authToken, userAgent string, cacheConfig *clients.CacheConfig) VBase {
	cl, vc := clients.CreateClient(endpoint, authToken, userAgent, cacheConfig)
	return &Client{cl, vc}
}

const (
	pathToBucket      = "/%v/%v/buckets/%v"
	pathToBucketState = "/%v/%v/buckets/%v/state"
	pathToFileList    = "/%v/%v/buckets/%v/files"
	pathToFile        = "/%v/%v/buckets/%v/files/%v"
)

// GetBucket describes the current state of a bucket
func (cl *Client) GetBucket(account, workspace, bucket string) (*BucketResponse, error) {
	const kind = "bucket"
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToBucket, account, workspace, bucket)).Send()
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 304 {
		cached, err := cl.cache.GetFor(kind, res)
		if err != nil {
			return nil, err
		}
		return cached.(*BucketResponse), nil
	}

	var bucketResponse BucketResponse
	if err := res.JSON(&bucketResponse); err != nil {
		return nil, err
	}

	cl.cache.SetFor(kind, res, &bucketResponse)

	return &bucketResponse, nil
}

// SetBucketState sets the current state of a bucket
func (cl *Client) SetBucketState(account, workspace, bucket, state string) error {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToBucketState, account, workspace, bucket)).
		JSON(state).Send()
	if err != nil {
		return err
	}

	return nil
}

// GetFile gets a file's content as a read closer
func (cl *Client) GetFile(account, workspace, bucket, path string) (io.ReadCloser, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).Send()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetFileB gets a file's content as bytes
func (cl *Client) GetFileB(account, workspace, bucket, path string) ([]byte, error) {
	const kind = "file-bytes"
	res, err := cl.GetFile(account, workspace, bucket, path)
	if err != nil {
		return nil, err
	}

	gentRes := res.(*gentleman.Response)
	if gentRes.StatusCode == 304 {
		cached, err := cl.cache.GetFor(kind, gentRes)
		if err != nil {
			return nil, err
		}
		return cached.([]byte), nil
	}

	bytes := gentRes.Bytes()
	cl.cache.SetFor(kind, gentRes, bytes)

	return bytes, nil
}

// GetFileConflict gets a file's content as a byte slice, or conflict
func (cl *Client) GetFileConflict(account, workspace, bucket, path string) (io.ReadCloser, *Conflict, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).
		Use(headers.Set("x-conflict-resolution", "merge")).Send()

	if err != nil {
		if err, ok := err.(clients.ResponseError); ok && err.StatusCode == 409 {
			var conflict Conflict
			if err := res.JSON(&conflict); err != nil {
				return nil, nil, err
			}
			return nil, &conflict, nil
		}
		return nil, nil, err
	}

	return res, nil, nil
}

// GetFileConflictB gets a file's content as bytes or conflict
func (cl *Client) GetFileConflictB(account, workspace, bucket, path string) ([]byte, *Conflict, error) {
	const kind = "file-conf-bytes"
	res, conflict, err := cl.GetFileConflict(account, workspace, bucket, path)
	if err != nil {
		return nil, nil, err
	}

	if conflict != nil {
		return nil, conflict, nil
	}

	gentRes := res.(*gentleman.Response)
	if gentRes.StatusCode == 304 {
		cached, err := cl.cache.GetFor(kind, gentRes)
		if err != nil {
			return nil, nil, err
		}
		return cached.([]byte), nil, nil
	}

	bytes := gentRes.Bytes()
	cl.cache.SetFor(kind, gentRes, bytes)

	return bytes, nil, nil
}

// SaveFile saves a file to a workspace
func (cl *Client) SaveFile(account, workspace, bucket, path string, body io.Reader) error {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).
		Body(body).Send()

	return err
}

// SaveFileB saves a file to a workspace
func (cl *Client) SaveFileB(account, workspace, bucket, path string, body []byte, contentType string, unzip bool) error {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).
		SetQuery("unzip", fmt.Sprintf("%v", unzip)).
		Body(bytes.NewReader(body)).Send()

	return err
}

// ListFiles returns a list of files, given a prefix
func (cl *Client) ListFiles(account, workspace, bucket, prefix, marker string, size int) (*FileListResponse, error) {
	const kind = "file-list"
	if size <= 0 {
		size = 100
	}
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFileList, account, workspace, bucket)).
		SetQueryParams(map[string]string{
			"prefix": prefix,
			"_next":  marker,
			"_limit": fmt.Sprintf("%d", size),
		}).Send()

	if err != nil {
		return nil, err
	}

	if res.StatusCode == 304 {
		cached, err := cl.cache.GetFor(kind, res)
		if err != nil {
			return nil, err
		}
		return cached.(*FileListResponse), nil
	}

	var fileListResponse FileListResponse
	if err := res.JSON(&fileListResponse); err != nil {
		return nil, err
	}

	cl.cache.SetFor(kind, res, &fileListResponse)

	return &fileListResponse, nil
}

// ListAllFiles returns a complete list of files, given a prefix
func (cl *Client) ListAllFiles(account, workspace, bucket, prefix string, size int) (*FileListResponse, error) {
	partialList, err := cl.ListFiles(account, workspace, bucket, prefix, "", size)
	if err != nil {
		return nil, err
	}

	for {
		if partialList.NextMarker == "" {
			break
		}

		newPartialList, err := cl.ListFiles(account, workspace, bucket, prefix, partialList.NextMarker, size)
		if err != nil {
			return nil, err
		}

		for _, v := range newPartialList.Files {
			partialList.Files = append(partialList.Files, v)
		}

		partialList.NextMarker = newPartialList.NextMarker
	}

	return partialList, nil
}

// DeleteFile deletes a file from the workspace
func (cl *Client) DeleteFile(account, workspace, bucket, path string) error {
	_, err := cl.http.Delete().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).Send()

	return err
}
