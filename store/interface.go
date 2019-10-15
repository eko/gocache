package store

// StoreInterface is the interface for all available stores
type StoreInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, options *Options) error
	GetType() string
}
