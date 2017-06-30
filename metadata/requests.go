package metadata

type MetadataPatchRequest []*PatchOperation

type PatchOperation struct {
	Type  OperationType `json:"op"`
	Key   string        `json:"path"`
	Value interface{}   `json:"value,omitempty"`
}

type OperationType string

const (
	OperationTypeAdd     = OperationType("add")
	OperationTypeReplace = OperationType("replace")
	OperationTypeRemove  = OperationType("remove")
)
