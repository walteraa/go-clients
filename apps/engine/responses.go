package engine

// DependencyTree is the recursive representation of dependencies
//      {
//          "foo.bar@1.2.3": {
//              "boo.zaz@2.0.1-beta": {},
//              "boo.yay@1.0.0": {}
//          },
//          "lala.ha@0.1.0": {}
//      }
type DependencyTree map[string]DependencyTree

// Manifest is an installed app's manifest
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
	DependencyTree   DependencyTree    `json:"_dependencyTree"`
	DependencySet    []string          `json:"_dependencySet"`
	ActivationDate   string            `json:"_activationDate"`
	Link             string            `json:"_link,omitempty"`
}
