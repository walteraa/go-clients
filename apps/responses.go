package apps

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
type Manifest struct {
	Registry       string            `json:"registry"`
	ID             string            `json:"id"`
	Vendor         string            `json:"vendor"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Dependencies   map[string]string `json:"dependencies"`
	DependencyTree DependencyTree    `json:"dependencyTree"`
	Linked         bool              `json:"link"`
}

type FileList struct {
	Files []*File `json:"data"`
}

type File struct {
	Path string `json:"file"`
	Hash string `json:"hash"`
}
