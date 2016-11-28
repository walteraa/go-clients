package workspaces

// BucketResponse is the description of a bucket's state
type BucketResponse struct {
	Hash  string `json:"hash"`
	State string `json:"state"`
}

// FileListEntryResponse is the description of an entry in a FileListResponse
type FileListEntryResponse struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// FileListResponse is the description of file list
type FileListResponse struct {
	Files      []*FileListEntryResponse `json:"files"`
	NextMarker string                   `json:"nextMarker"`
}
