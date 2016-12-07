package registry

type IdentityListResponseEntry struct {
	Vendor   string `json:"vendor"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Location string `josn:"location"`
}

type IdentityListResponse struct {
	Identities []*IdentityListResponseEntry `json:"identities"`
}

type FileListResponseEntry struct {
	Path     string `json:"file"`
	Hash     string `json:"hash"`
	Location string `json:"location"`
}

type FileListResponse struct {
	Files []*FileListResponseEntry `json:"files"`
}
