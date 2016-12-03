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
