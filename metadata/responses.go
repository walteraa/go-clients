package metadata

import "encoding/json"

type BucketResponse struct {
	Hash string `json:"hash"`
}

type MetadataListResponse struct {
	Data       []*MetadataResponseEntry `json:"data"`
	NextMarker string                   `json:"next"`
}

type MetadataResponseEntry struct {
	Key   string          `json:"key"`
	Hash  string          `json:"hash"`
	Value json.RawMessage `json:"value"`
}
