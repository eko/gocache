package extension

import (
	"github.com/yeqown/gocache/cache"
	"github.com/yeqown/gocache/store"

	"github.com/vmihailenco/msgpack"
)

type IMarshal interface {
	Marshal(in interface{}) ([]byte, error)
	Unmarshal(data []byte, out interface{}) error
}

type msgpackMarshal struct{}

func newMsgpackMarshal() *msgpackMarshal {
	return &msgpackMarshal{}
}

func (m msgpackMarshal) Marshal(in interface{}) ([]byte, error) {
	return msgpack.Marshal(in)
}

func (m msgpackMarshal) Unmarshal(data []byte, out interface{}) error {
	return msgpack.Unmarshal(data, out)
}

var (
	_ cache.ICache = marshalWrapper{}
)

// 序列化装饰器
type marshalWrapper struct {
	IMarshal

	c cache.ICache
}

// WrapWithMarshal .
func WrapWithMarshal(c cache.ICache) cache.ICache {
	return &marshalWrapper{
		IMarshal: newMsgpackMarshal(),
		c:        c,
	}
}

func (m marshalWrapper) Get(key interface{}, returnObj interface{}) (interface{}, error) {
	result, err := m.c.Get(key, nil)
	if err != nil {
		return nil, err
	}

	switch result.(type) {
	case []byte:
		err = m.Unmarshal(result.([]byte), returnObj)
	case string:
		err = m.Unmarshal([]byte(result.(string)), returnObj)
	}

	if err != nil {
		return nil, err
	}

	return returnObj, nil
}

func (m marshalWrapper) Set(key, object interface{}, options *store.Options) error {
	bytes, err := m.Marshal(object)
	if err != nil {
		return err
	}

	return m.c.Set(key, bytes, options)
}

func (m marshalWrapper) Delete(key interface{}) error {
	return m.c.Delete(key)
}

func (m marshalWrapper) Invalidate(options store.InvalidateOptions) error {
	return m.c.Invalidate(options)
}

func (m marshalWrapper) Clear() error {
	return m.c.Clear()
}

func (m marshalWrapper) GetType() string {
	return m.c.GetType() + ".marshal"
}

func (m marshalWrapper) GetStats() *cache.Stats {
	return m.c.GetStats()
}
