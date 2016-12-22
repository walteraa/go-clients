package registry

type Manifest struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Categories   []string          `json:"categories"`
	Dependencies map[string]string `json:"dependencies"`
	Services     []string          `json:"_services"`
}

type IdentityListResponseEntry struct {
	Vendor   string `json:"vendor"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Location string `josn:"location"`
}

type IdentityListResponse struct {
	Identities []*IdentityListResponseEntry `json:"identities"`
}

type File struct {
	Path     string `json:"file"`
	Hash     string `json:"hash"`
	Location string `json:"location"`
}

type FileList struct {
	Files []*File `json:"data"`
}
