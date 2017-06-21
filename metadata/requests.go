package metadata

type MetadataPatchRequest []*PatchOperation

type OperationType int

const (
	_ OperationType = iota
	OperationTypeSave
	OperationTypeDelete
)

type PatchOperation struct {
	Type  OperationType
	Key   string
	Value interface{}
}
