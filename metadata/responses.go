package metadata

import "encoding/json"

type BucketResponse struct {
	Hash string `json:"hash"`
}

type MetadataListResponse struct {
	Data []*MetadataResponse `json:"data"`
}

type MetadataResponse struct {
	Key   string          `json:"key"`
	Hash  string          `json:"hash"`
	Value json.RawMessage `json:"value"`
}
