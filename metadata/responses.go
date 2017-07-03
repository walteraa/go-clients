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
	Key    string                 `json:"key"`
	Master *MetadataConflictEntry `json:"master"`
	Base   *MetadataConflictEntry `json:"base"`
	Mine   *MetadataConflictEntry `json:"mine"`
}

type MetadataConflictEntry struct {
	Value   json.RawMessage `json:"value"`
	Deleted bool            `json:"deleted"`
}
