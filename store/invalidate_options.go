package store

// InvalidateOptions represents the cache invalidation available options
type InvalidateOptions struct {
	// Tags allows to specify associated tags to the current value
	Tags []string
}

// TagsValue returns the tags option value
func (o InvalidateOptions) TagsValue() []string {
	return o.Tags
}
