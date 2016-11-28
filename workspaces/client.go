package workspaces

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"net/url"

	"github.com/vtex/go-clients/errors"
)

var hcli = &http.Client{
	Timeout: time.Second * 10,
}

// Workspaces is an interface for interacting with workspaces
type Workspaces interface {
	GetBucket(account, workspace, bucket string) (*BucketResponse, error)
	SetBucketState(account, workspace, bucket, state string) error
	GetFile(account, workspace, bucket, path string) (io.ReadCloser, error)
	GetFileB(account, workspace, bucket, path string) ([]byte, error)
	SaveFile(account, workspace, bucket, path string, body io.Reader) error
	SaveFileB(account, workspace, bucket, path string, content []byte, unzip bool) error
	ListFiles(account, workspace, bucket, prefix, marker string, size int) (*FileListResponse, error)
	ListAllFiles(account, workspace, bucket, prefix string, size int) (*FileListResponse, error)
}

// Client is a struct that provides interaction with workspaces
type Client struct {
	Endpoint  string
	AuthToken string
	UserAgent string
}

// NewClient creates a new Workspaces client
func NewClient(endpoint, authToken, userAgent string) Workspaces {
	return &Client{Endpoint: endpoint, AuthToken: authToken, UserAgent: userAgent}
}

const (
	pathToBucket      = "%v/%v/buckets/%v"
	pathToBucketState = "%v/%v/buckets/%v/state"
	pathToFile        = "%v/%v/buckets/%v/files/%v?unzip=%t"
	pathToFileList    = "%v/%v/buckets/%v/files?prefix=%v&marker=%v&size=%d"
)

func (cl *Client) createRequestB(method string, content []byte, pathFormat string, a ...interface{}) *http.Request {
	var body io.Reader
	if content != nil {
		body = bytes.NewBuffer(content)
	}
	return cl.createRequest(method, body, pathFormat, a...)
}

func (cl *Client) createRequest(method string, body io.Reader, pathFormat string, a ...interface{}) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf(cl.Endpoint+pathFormat, a...), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "token "+cl.AuthToken)
	req.Header.Set("User-Agent", cl.UserAgent)
	return req
}

// GetBucket describes the current state of a bucket
func (cl *Client) GetBucket(account, workspace, bucket string) (*BucketResponse, error) {
	req := cl.createRequest("GET", nil, pathToBucket, account, workspace, bucket)
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return nil, reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return nil, err
	}

	var bucketResponse BucketResponse
	buf, buferr := ioutil.ReadAll(res.Body)
	if buferr != nil {
		return nil, buferr
	}
	json.Unmarshal(buf, &bucketResponse)
	return &bucketResponse, nil
}

// SetBucket sets the current state of a bucket
func (cl *Client) SetBucketState(account, workspace, bucket, state string) error {
	req := cl.createRequest("PUT", bytes.NewBufferString(state), pathToBucketState, account, workspace, bucket)
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return err
	}

	return nil
}

// GetFile gets a file's content as a read closer
func (cl *Client) GetFile(account, workspace, bucket, path string) (io.ReadCloser, error) {
	req := cl.createRequest("GET", nil, pathToFile, account, workspace, bucket, path)
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return nil, reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return nil, err
	}

	return res.Body, nil
}

// GetFileB gets a file's content as bytes
func (cl *Client) GetFileB(account, workspace, bucket, path string) ([]byte, error) {
	body, err := cl.GetFile(account, workspace, bucket, path)
	if err != nil {
		return nil, err
	}

	buf, buferr := ioutil.ReadAll(body)
	if buferr != nil {
		return nil, buferr
	}
	return buf, nil
}

// SaveFile saves a file to a workspace
func (cl *Client) SaveFile(account, workspace, bucket, path string, body io.Reader) error {
	req := cl.createRequest("PUT", body, pathToFile, account, workspace, bucket, path)
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return err
	}
	return nil
}

// SaveFileB saves a file to a workspace
func (cl *Client) SaveFileB(account, workspace, bucket, path string, body []byte, unzip bool) error {
	req := cl.createRequestB("PUT", body, pathToFile, account, workspace, bucket, path, unzip)
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return err
	}
	return nil
}

// ListFiles returns a list of files, given a prefix
func (cl *Client) ListFiles(account, workspace, bucket, prefix, marker string, size int) (*FileListResponse, error) {
	if size <= 0 {
		size = 100
	}
	prefix = url.QueryEscape(prefix)
	marker = url.QueryEscape(marker)

	req := cl.createRequest("GET", nil, pathToFile, account, workspace, bucket, prefix, marker, size)
	res, reserr := hcli.Do(req)
	if reserr != nil {
		return nil, reserr
	}
	if err := errors.StatusCode(res); err != nil {
		return nil, err
	}

	var fileListResponse FileListResponse
	buf, buferr := ioutil.ReadAll(res.Body)
	if buferr != nil {
		return nil, buferr
	}
	json.Unmarshal(buf, &fileListResponse)
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
