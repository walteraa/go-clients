package vbase

import (
	"bytes"
	"fmt"
	"io"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
	"gopkg.in/h2non/gentleman.v1/plugins/headers"
	"net/http"
	"io/ioutil"
)

// Workspaces is an interface for interacting with workspaces
type VBase interface {
	GetBucket(account, workspace, bucket string) (*BucketResponse, string, error)
	SetBucketState(account, workspace, bucket, state string) (string, error)
	GetFile(account, workspace, bucket, path string) (io.ReadCloser, string, error)
	GetFileB(account, workspace, bucket, path string) ([]byte, http.Header, error)
	GetFileConflict(account, workspace, bucket, path string) (io.ReadCloser, *Conflict, string, error)
	GetFileConflictB(account, workspace, bucket, path string) ([]byte, *Conflict, string, error)
	SaveFile(account, workspace, bucket, path string, body io.Reader) (string, error)
	SaveFileB(account, workspace, bucket, path string, content []byte, contentType string, unzip bool) (string, error)
	ListFiles(account, workspace, bucket, prefix, marker string, size int) (*FileListResponse, string, error)
	ListAllFiles(account, workspace, bucket, prefix string, size int) (*FileListResponse, string, error)
	DeleteFile(account, workspace, bucket, path string) error
}

// Client is a struct that provides interaction with workspaces
type Client struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

// NewClient creates a new Workspaces client
func NewClient(endpoint, authToken, userAgent string, ttl int, reqCtx clients.RequestContext) VBase {
	cl, vc := clients.CreateClient(endpoint, authToken, userAgent, reqCtx, ttl)
	return &Client{cl, vc}
}

const (
	pathToBucket      = "/%v/%v/buckets/%v"
	pathToBucketState = "/%v/%v/buckets/%v/state"
	pathToFileList    = "/%v/%v/buckets/%v/files"
	pathToFile        = "/%v/%v/buckets/%v/files/%v"
)

// GetBucket describes the current state of a bucket
func (cl *Client) GetBucket(account, workspace, bucket string) (*BucketResponse, string, error) {
	const kind = "bucket"
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToBucket, account, workspace, bucket)).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*BucketResponse), res.Header.Get(clients.HeaderETag), nil
	}

	var bucketResponse BucketResponse
	if err := res.JSON(&bucketResponse); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &bucketResponse)

	return &bucketResponse, res.Header.Get(clients.HeaderETag), nil
}

// SetBucketState sets the current state of a bucket
func (cl *Client) SetBucketState(account, workspace, bucket, state string) (string, error) {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToBucketState, account, workspace, bucket)).
		JSON(state).Send()
	if err != nil {
		return "", err
	}

	return "", nil
}

func (cl *Client) getFile(account, workspace, bucket, path string) *gentleman.Request {
	return cl.http.Get().AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path))
}

// GetFile gets a file's content as a read closer
func (cl *Client) GetFile(account, workspace, bucket, path string) (io.ReadCloser, string, error) {
	res, err := cl.getFile(account, workspace, bucket, path).Send()
	if err != nil {
		return nil, res.Header.Get(clients.HeaderETag), err
	}

	return res, res.Header.Get(clients.HeaderETag), nil
}

// GetFileB gets a file's content as bytes
func (cl *Client) GetFileB(account, workspace, bucket, path string) ([]byte, http.Header, error) {
	const kind = "file-bytes"
	res, err := cl.getFile(account, workspace, bucket, path).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, nil, err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, nil, err
	} else if ok {
		return cached.([]byte), res.Header, nil
	}

	bytes, err := ioutil.ReadAll(res.RawResponse.Body)
	if err != nil {
		return nil, nil, err
	}

	cl.cache.SetFor(kind, res, bytes)

	return bytes, res.Header, nil
}

func (cl *Client) getFileConflict(account, workspace, bucket, path string, useCache bool) (io.ReadCloser, *Conflict, string, error) {
	req := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).
		Use(headers.Set("x-conflict-resolution", "merge"))
	if useCache {
		req.UseRequest(clients.Cache)
	}

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

// GetFileConflict gets a file's content as a byte slice, or conflict
func (cl *Client) GetFileConflict(account, workspace, bucket, path string) (io.ReadCloser, *Conflict, string, error) {
	return cl.getFileConflict(account, workspace, bucket, path, false)
}

// GetFileConflictB gets a file's content as bytes or conflict
func (cl *Client) GetFileConflictB(account, workspace, bucket, path string) ([]byte, *Conflict, string, error) {
	const kind = "file-conf-bytes"
	res, conflict, eTag, err := cl.getFileConflict(account, workspace, bucket, path, true)
	if err != nil {
		return nil, nil, eTag, err
	}

	if conflict != nil {
		return nil, conflict, eTag, nil
	}

	gentRes := res.(*gentleman.Response)
	if cached, ok, err := cl.cache.GetFor(kind, gentRes); err != nil {
		return nil, nil, "", err
	} else if ok {
		return cached.([]byte), nil, gentRes.Header.Get(clients.HeaderETag), nil
	}

	bytes := gentRes.Bytes()
	cl.cache.SetFor(kind, gentRes, bytes)

	return bytes, nil, gentRes.Header.Get(clients.HeaderETag), nil
}

// SaveFile saves a file to a workspace
func (cl *Client) SaveFile(account, workspace, bucket, path string, body io.Reader) (string, error) {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).
		Body(body).Send()

	return "", err
}

// SaveFileB saves a file to a workspace
func (cl *Client) SaveFileB(account, workspace, bucket, path string, body []byte, contentType string, unzip bool) (string, error) {
	res, err := cl.http.Put().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).
		SetQuery("unzip", fmt.Sprintf("%v", unzip)).
		Body(bytes.NewReader(body)).Send()

	if err != nil {
		return "", err
	}

	return res.Header.Get(clients.HeaderETag), nil
}

// ListFiles returns a list of files, given a prefix
func (cl *Client) ListFiles(account, workspace, bucket, prefix, marker string, size int) (*FileListResponse, string, error) {
	const kind = "file-list"
	if size <= 0 {
		size = 100
	}
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(pathToFileList, account, workspace, bucket)).
		UseRequest(clients.Cache).
		SetQueryParams(map[string]string{
			"prefix": prefix,
			"_next":  marker,
			"_limit": fmt.Sprintf("%d", size),
		}).Send()

	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*FileListResponse), res.Header.Get(clients.HeaderETag), nil
	}

	var fileListResponse FileListResponse
	if err := res.JSON(&fileListResponse); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &fileListResponse)

	return &fileListResponse, res.Header.Get(clients.HeaderETag), nil
}

// ListAllFiles returns a complete list of files, given a prefix
func (cl *Client) ListAllFiles(account, workspace, bucket, prefix string, size int) (*FileListResponse, string, error) {
	partialList, eTag, err := cl.ListFiles(account, workspace, bucket, prefix, "", size)
	if err != nil {
		return nil, "", err
	}

	for {
		if partialList.NextMarker == "" {
			break
		}

		newPartialList, newETag, err := cl.ListFiles(account, workspace, bucket, prefix, partialList.NextMarker, size)
		if err != nil {
			return nil, "", err
		}

		for _, v := range newPartialList.Files {
			partialList.Files = append(partialList.Files, v)
		}

		partialList.NextMarker = newPartialList.NextMarker
		eTag = newETag
	}

	return partialList, eTag, nil
}

// DeleteFile deletes a file from the workspace
func (cl *Client) DeleteFile(account, workspace, bucket, path string) error {
	_, err := cl.http.Delete().
		AddPath(fmt.Sprintf(pathToFile, account, workspace, bucket, path)).Send()

	return err
}
