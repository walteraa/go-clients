package metadata

import (
	"fmt"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
	"strconv"
)

type Metadata interface {
	GetBucket(bucket string) (*BucketResponse, string, error)
	SetBucketState(bucket, state string) error
	List(bucket string, includeValue bool, limit int) (*MetadataListResponse, string, error)
	Get(bucket, key string, data interface{}) (string, error)
	Save(bucket, key string, data interface{}) (string, error)
	Delete(bucket, key string) (bool, error)
}

type Client struct {
	http *gentleman.Client
}

func NewClient(config *clients.Config) Metadata {
	cl := clients.CreateClient("kube-router", config, true)
	return &Client{cl}
}

const (
	bucketPath      = "/buckets/%v"
	bucketStatePath = "/buckets/%v/state"
	metadataPath    = "/buckets/%v/metadata"
	metadataKeyPath = "/buckets/%v/metadata/%v"
)

func (cl *Client) GetBucket(bucket string) (*BucketResponse, string, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(bucketPath, bucket)).Send()
	if err != nil {
		return nil, "", err
	}

	var bucketResponse BucketResponse
	if err := res.JSON(&bucketResponse); err != nil {
		return nil, "", err
	}

	return &bucketResponse, res.Header.Get("ETag"), nil
}

func (cl *Client) SetBucketState(bucket, state string) error {
	_, err := cl.http.Put().
		AddPath(fmt.Sprintf(bucketStatePath, bucket)).
		JSON(state).Send()
	if err != nil {
		return err
	}
	return nil
}

func (cl *Client) List(bucket string, includeValue bool, limit int) (*MetadataListResponse, string, error) {
	req := cl.http.Get().AddPath(fmt.Sprintf(metadataPath, bucket))
	req = req.SetQuery("_limit", strconv.Itoa(limit))
	if includeValue {
		req = req.SetQuery("value", "true")
	}

	res, err := req.Send()
	if err != nil {
		return nil, "", err
	}

	var metadata MetadataListResponse
	if err := res.JSON(&metadata); err != nil {
		return nil, "", err
	}

	return &metadata, res.Header.Get("ETag"), nil
}

func (cl *Client) Get(bucket, key string, data interface{}) (string, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataKeyPath, bucket, key)).Send()
	if err != nil {
		return "", err
	}

	if err := res.JSON(data); err != nil {
		return "", err
	}

	return res.Header.Get("ETag"), nil
}

func (cl *Client) Save(bucket, key string, data interface{}) (string, error) {
	res, err := cl.http.Put().
		AddPath(fmt.Sprintf(metadataKeyPath, bucket, key)).
		JSON(data).Send()

	if err != nil {
		return "", err
	}

	return res.Header.Get("ETag"), nil
}

func (cl *Client) Delete(bucket, key string) (bool, error) {
	_, err := cl.http.Delete().
		AddPath(fmt.Sprintf(metadataKeyPath, bucket, key)).Send()

	if err != nil {
		if err, ok := err.(clients.ResponseError); ok && err.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
