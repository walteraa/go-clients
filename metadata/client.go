package metadata

import (
	"fmt"
	"reflect"

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
	http  *gentleman.Client
	cache clients.ValueCache
}

func NewClient(config *clients.Config) Metadata {
	cl, vc := clients.CreateClient("kube-router", config, true)
	return &Client{cl, vc}
}

const (
	bucketPath      = "/buckets/%v"
	bucketStatePath = "/buckets/%v/state"
	metadataPath    = "/buckets/%v/metadata"
	metadataKeyPath = "/buckets/%v/metadata/%v"
)

func (cl *Client) GetBucket(bucket string) (*BucketResponse, string, error) {
	const kind = "bucket"
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(bucketPath, bucket)).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*BucketResponse), res.Header.Get("ETag"), nil
	}

	var bucketResponse BucketResponse
	if err := res.JSON(&bucketResponse); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &bucketResponse)

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
	const kind = "list"
	req := cl.http.Get().AddPath(fmt.Sprintf(metadataPath, bucket))
	req = req.SetQuery("_limit", strconv.Itoa(limit))
	if includeValue {
		req = req.SetQuery("value", "true")
	}
	req = req.UseRequest(clients.Cache)

	res, err := req.Send()
	if err != nil {
		return nil, "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return nil, "", err
	} else if ok {
		return cached.(*MetadataListResponse), res.Header.Get("ETag"), nil
	}

	var metadata MetadataListResponse
	if err := res.JSON(&metadata); err != nil {
		return nil, "", err
	}

	cl.cache.SetFor(kind, res, &metadata)

	return &metadata, res.Header.Get("ETag"), nil
}

func (cl *Client) Get(bucket, key string, data interface{}) (string, error) {
	const kind = "get"
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataKeyPath, bucket, key)).
		UseRequest(clients.Cache).Send()
	if err != nil {
		return "", err
	}

	if cached, ok, err := cl.cache.GetFor(kind, res); err != nil {
		return "", err
	} else if ok {
		vt := reflect.ValueOf(data)
		pt := vt.Elem()
		pt.Set(reflect.Indirect(reflect.ValueOf(cached).Convert(vt.Type())))
		return res.Header.Get("ETag"), nil
	}

	if err := res.JSON(data); err != nil {
		return "", err
	}

	cl.cache.SetFor(kind, res, data)
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
