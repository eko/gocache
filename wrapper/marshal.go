package wrapper

import (
	"github.com/yeqown/gocache"
	"github.com/yeqown/gocache/types"

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
	_ gocache.ICache = marshalWrapper{}
)

// 序列化装饰器
type marshalWrapper struct {
	IMarshal

	c gocache.ICache
}

// WrapWithMarshal .
func WrapWithMarshal(c gocache.ICache) gocache.ICache {
	return &marshalWrapper{
		IMarshal: newMsgpackMarshal(),
		c:        c,
	}
}

func (m marshalWrapper) Get(key string) ([]byte, error) {
	result, err := m.c.Get(key)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m marshalWrapper) Set(key string, object interface{}, options *types.StoreOptions) error {
	bytes, err := m.Marshal(object)
	if err != nil {
		return err
	}

	return m.c.Set(key, bytes, options)
}

func (m marshalWrapper) Invalidate(opt types.InvalidateOptions) error { return m.c.Invalidate(opt) }
func (m marshalWrapper) Delete(key string) error                      { return m.c.Delete(key) }
func (m marshalWrapper) GetType() string                              { return m.c.GetType() + ".marshal" }

// func (m marshalWrapper) Clear() error                                 { return m.c.Clear() }
