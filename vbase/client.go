package vbase

import (
	"bytes"
	"fmt"
	"io"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"strconv"
)

type Options struct {
	Prefix string
	Marker string
	Limit  int
}

// Workspaces is an interface for interacting with workspaces
type VBase interface {
	GetBucket(bucket string) (*BucketResponse, string, error)
	SetBucketState(bucket, state string) (string, error)
	GetFile(bucket, path string) (*gentleman.Response, string, error)
	GetFileConflict(bucket, path string) (*gentleman.Response, *Conflict, string, error)
	SaveFile(bucket, path string, body io.Reader) (string, error)
	SaveFileB(bucket, path string, content []byte, contentType string, unzip bool) (string, error)
	ListFiles(bucket string, options *Options) (*FileListResponse, string, error)
	ListAllFiles(bucket, prefix string) (*FileListResponse, string, error)
	DeleteFile(bucket, path string) error
}

// Client is a struct that provides interaction with workspaces
type Client struct {
	http *gentleman.Client
}

// NewClient creates a new Workspaces client
func NewClient(config *clients.Config) VBase {
	cl := clients.CreateClient("vbase", config, true)
	return &Client{cl}
}

const (
	pathToBucket      = "/buckets/%v"
	pathToBucketState = "/buckets/%v/state"
	pathToFileList    = "/buckets/%v/files"
	pathToFile        = "/buckets/%v/files/%v"
)

// GetBucket describes the current state of a bucket
func (cl *Client) GetBucket(bucket string) (*BucketResponse, string, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToBucket, bucket)).Send()
	if err != nil {
		return nil, "", err
	}

	var bucketResponse BucketResponse
	if err := res.JSON(&bucketResponse); err != nil {
		return nil, "", err
	}

	return &bucketResponse, res.Header.Get(clients.HeaderETag), nil
}

// SetBucketState sets the current state of a bucket
func (cl *Client) SetBucketState(bucket, state string) (string, error) {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToBucketState, bucket)).
		JSON(state).Send()
	if err != nil {
		return "", err
	}

	return "", nil
}

// GetFile gets a file's content as a read closer
func (cl *Client) GetFile(bucket, path string) (*gentleman.Response, string, error) {
	res, err := cl.http.Get().AddPath(fmt.Sprintf(pathToFile, bucket, path)).Send()
	if err != nil {
		return nil, res.Header.Get(clients.HeaderETag), err
	}

	return res, res.Header.Get(clients.HeaderETag), nil
}

// GetFileConflict gets a file's content as a byte slice, or conflict
func (cl *Client) GetFileConflict(bucket, path string) (*gentleman.Response, *Conflict, string, error) {
	req := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFile, bucket, path)).
		Use(headers.Set("x-conflict-resolution", "merge"))

	res, err := req.Send()
	if err != nil {
		if err, ok := err.(clients.ResponseError); ok && err.StatusCode == 409 {
			var conflict Conflict
			if err := res.JSON(&conflict); err != nil {
				return nil, nil, "", err
			}
			return nil, &conflict, res.Header.Get(clients.HeaderETag), nil
		}
		return nil, nil, "", err
	}

	return res, nil, res.Header.Get(clients.HeaderETag), nil
}

// SaveFile saves a file to a workspace
func (cl *Client) SaveFile(bucket, path string, body io.Reader) (string, error) {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToFile, bucket, path)).
		Body(body).Send()

	return "", err
}

// SaveFileB saves a file to a workspace
func (cl *Client) SaveFileB(bucket, path string, body []byte, contentType string, unzip bool) (string, error) {
	res, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToFile, bucket, path)).
		SetQuery("unzip", fmt.Sprintf("%v", unzip)).
		Body(bytes.NewReader(body)).Send()

	if err != nil {
		return "", err
	}

	return res.Header.Get(clients.HeaderETag), nil
}

// ListFiles returns a list of files, given a prefix
func (cl *Client) ListFiles(bucket string, options *Options) (*FileListResponse, string, error) {
	if options.Limit <= 0 {
		options.Limit = 10
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFileList, bucket)).
		SetQueryParams(map[string]string{
			"prefix": options.Prefix,
			"_next":  options.Marker,
			"_limit": strconv.Itoa(options.Limit),
		}).Send()

	if err != nil {
		return nil, "", err
	}

	var fileListResponse FileListResponse
	if err := res.JSON(&fileListResponse); err != nil {
		return nil, "", err
	}

	return &fileListResponse, res.Header.Get(clients.HeaderETag), nil
}

// ListAllFiles returns a complete list of files, given a prefix
func (cl *Client) ListAllFiles(bucket, prefix string) (*FileListResponse, string, error) {
	options := &Options{
		Limit:  100,
		Prefix: prefix,
	}

	list, eTag, err := cl.ListFiles(bucket, options)
	if err != nil {
		return nil, "", err
	}

	for {
		if list.NextMarker == "" {
			break
		}
		options.Marker = list.NextMarker

		partialList, newETag, err := cl.ListFiles(bucket, options)
		if err != nil {
			return nil, "", err
		}

		list.Files = append(list.Files, partialList.Files...)
		list.NextMarker = partialList.NextMarker
		eTag = newETag
	}
	return list, eTag, nil
}

// DeleteFile deletes a file from the workspace
func (cl *Client) DeleteFile(bucket, path string) error {
	_, err := cl.http.Delete().
		AddPath(fmt.Sprintf(pathToFile, bucket, path)).Send()

	return err
}
