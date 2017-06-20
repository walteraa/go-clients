package metadata

import (
	"fmt"

	"strconv"

	"github.com/vtex/go-clients/clients"
	"gopkg.in/h2non/gentleman.v1"
)

type Options struct {
	IncludeValue bool
	Limit        int
	Marker       string
}

type Metadata interface {
	GetBucket(bucket string) (*BucketResponse, string, error)
	SetBucketState(bucket, state string) error
	List(bucket string, options *Options) (*MetadataListResponse, string, error)
	ListAll(bucket string, includeValue bool) (*MetadataListResponse, string, error)
	Get(bucket, key string, data interface{}) (string, error)
	Save(bucket, key string, data interface{}) (string, error)
	SaveAll(bucket string, data map[string]interface{}) (string, error)
	Delete(bucket, key string) (bool, error)
	ListConflicts(bucket string) (MetadataConflictMap, error)
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
	conflictsPath   = "/buckets/%v/conflicts"
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

func (cl *Client) List(bucket string, options *Options) (*MetadataListResponse, string, error) {
	if options.Limit <= 0 {
		options.Limit = 10
	}

	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(metadataPath, bucket)).
		SetQueryParams(map[string]string{
			"value":   strconv.FormatBool(options.IncludeValue),
			"_limit":  strconv.Itoa(options.Limit),
			"_marker": options.Marker,
		}).Send()

	if err != nil {
		return nil, "", err
	}

	var metadata MetadataListResponse
	if err := res.JSON(&metadata); err != nil {
		return nil, "", err
	}

	return &metadata, res.Header.Get(clients.HeaderETag), nil
}

func (cl *Client) ListAll(bucket string, includeValue bool) (*MetadataListResponse, string, error) {
	options := &Options{
		Limit:        100,
		IncludeValue: includeValue,
	}

	list, eTag, err := cl.List(bucket, options)
	if err != nil {
		return nil, "", err
	}

	for {
		if list.NextMarker == "" {
			break
		}
		options.Marker = list.NextMarker

		partialList, newETag, err := cl.List(bucket, options)
		if err != nil {
			return nil, "", err
		}

		list.Data = append(list.Data, partialList.Data...)
		list.NextMarker = partialList.NextMarker
		eTag = newETag
	}
	return list, eTag, nil
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

func (cl *Client) SaveAll(bucket string, data map[string]interface{}) (string, error) {
	res, err := cl.http.Put().
		AddPath(fmt.Sprintf(metadataPath, bucket)).
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

func (cl *Client) ListConflicts(bucket string) (MetadataConflictMap, error) {
	res, err := cl.http.Get().
		AddPath(fmt.Sprintf(conflictsPath, bucket)).
		Send()
	if err != nil {
		return nil, err
	}

	var conflicts map[string]*MetadataConflict
	if err := res.JSON(&conflicts); err != nil {
		return nil, fmt.Errorf("Error unmarshaling metadata conflicts: %v", err)
	}

	return conflicts, nil
}
