package serializer

type Serializer interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}
