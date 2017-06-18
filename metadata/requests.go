package metadata

type MetadataPatchRequest []*PatchOperation

type OperationType int

const (
	_ OperationType = iota
	SaveOperation
	DeleteOperation
)

type PatchOperation struct {
	Operation OperationType
	Key       string
	Value     interface{}
}
