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

type MetadataConflictListResponse struct {
	Data []*MetadataConflict `json:"data"`
}

type MetadataConflict struct {
	Key    string                  `json:"key"`
	Values *MetadataConflictValues `json:"values"`
}

type MetadataConflictValues struct {
	Master json.RawMessage `json:"master"`
	Base   json.RawMessage `json:"base"`
	Mine   json.RawMessage `json:"mine"`
}
