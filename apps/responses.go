package apps

type File struct {
	Path     string `json:"path"`
	Hash     string `json:"hash,omitempty"`
	Location string `json:"location,omitempty"`
}

type FileList struct {
	Files []*File `json:"data"`
}
