package store

// StoreInterface is the interface for all available stores
type StoreInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, options *Options) error
	Delete(key interface{}) error
	Invalidate(options InvalidateOptions) error
	Clear() error
	GetType() string
}
