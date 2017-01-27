package registry

type Metadata struct {
	Vendor           string            `json:"vendor"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Categories       []string          `json:"categories"`
	Dependencies     map[string]string `json:"dependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	SettingsSchema   interface{}       `json:"settingsSchema"`
	ID               string            `json:"_id"`
	Publisher        string            `json:"_publisher"`
	PublicationDate  string            `json:"_publicationDate"`
}
