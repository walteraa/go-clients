package metadata

import (
	"fmt"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

type Metadata interface {
	GetBucket(account, workspace, bucket string) (*BucketResponse, string, error)
	List(account, workspace, bucket string, includeValue bool) (*MetadataListResponse, string, error)
	Get(account, workspace, bucket, key string, data interface{}) (string, error)
	Save(account, workspace, bucket, key string, data interface{}) (string, error)
	Delete(account, workspace, bucket, key string) (bool, error)
}

type Client struct {
	http  *gentleman.Client
	cache clients.ValueCache
}

func NewClient(endpoint, authToken, userAgent string, cacheConfig *clients.CacheConfig) Metadata {
	cl, vc := clients.CreateClient(endpoint, authToken, userAgent, cacheConfig)
	return &Client{cl, vc}
}

const (
	bucketPath      = "/%v/%v/buckets/%v"
	metadataPath    = "/%v/%v/buckets/%v/metadata"
	metadataKeyPath = "/%v/%v/buckets/%v/metadata/%v"
)

func (cl *Client) GetBucket(account, workspace, bucket string) (*BucketResponse, string, error) {
	const kind = "bucket"
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(bucketPath, account, workspace, bucket)).Send()
	if err != nil {
		return nil, "", err
	}

	if res.StatusCode == 304 {
		cached, err := cl.cache.GetFor(kind, res)
		if err != nil {
			return nil, "", err
		}
		return cached.(*BucketResponse), res.Header.Get("ETag"), nil
	}

	var bucketResponse BucketResponse
	if err := res.JSON(&bucketResponse); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &bucketResponse)

	return &bucketResponse, res.Header.Get("ETag"), nil
}

func (cl *Client) List(account, workspace, bucket string, includeValue bool) (*MetadataListResponse, string, error) {
	const kind = "list"
	req := cl.http.Get().AddPath(fmt.Sprintf(metadataPath, account, workspace, bucket))
	if includeValue {
		req = req.AddQuery("value", "true")
	}

	res, err := req.Send()
	if err != nil {
		return nil, "", err
	}

	if res.StatusCode == 304 {
		cached, err := cl.cache.GetFor(kind, res)
		if err != nil {
			return nil, "", err
		}
		return cached.(*MetadataListResponse), res.Header.Get("ETag"), nil
	}

	var metadata MetadataListResponse
	if err := res.JSON(&metadata); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &metadata)

	return &metadata, res.Header.Get("ETag"), nil
}

func (cl *Client) Get(account, workspace, bucket, key string, data interface{}) (string, error) {
	const kind = "get"
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataKeyPath, account, workspace, bucket, key)).Send()
	if err != nil {
		return "", err
	}

	if res.StatusCode == 304 {
		data, err = cl.cache.GetFor(kind, res)
		if err != nil {
			return "", err
		}
		return res.Header.Get("ETag"), nil
	}

	if err := res.JSON(data); err != nil {
		return "", err
	}

	cl.cache.SetFor(kind, res, data)
	return res.Header.Get("ETag"), nil
}

func (cl *Client) Save(account, workspace, bucket, key string, data interface{}) (string, error) {
	res, err := cl.http.Put().
		AddPath(fmt.Sprintf(metadataKeyPath, account, workspace, bucket, key)).
		JSON(data).Send()

	if err != nil {
		return "", err
	}

	return res.Header.Get("ETag"), nil
}

func (cl *Client) Delete(account, workspace, bucket, key string) (bool, error) {
	_, err := cl.http.Delete().
		AddPath(fmt.Sprintf(metadataKeyPath, account, workspace, bucket, key)).Send()

	if err != nil {
		if err, ok := err.(clients.ResponseError); ok && err.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
