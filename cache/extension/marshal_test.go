package extension

import (
	"testing"

	mocksCache "github.com/yeqown/gocache/test/mocks/cache"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name  string
	Age   int
	Embed struct {
		Score float64
	}
}

func Test_Marshal_GetAndSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "key"
	value := &testStruct{
		Name: "asdasd",
		Age:  10,
		Embed: struct {
			Score float64
		}{
			Score: 2903.2312,
		},
	}

	marshaler := newMsgpackMarshal()
	data, err := marshaler.Marshal(value)
	assert.Nil(t, err)

	c := mocksCache.NewMockICache(ctrl)
	c.EXPECT().Set(key, data, nil).Return(nil)
	c.EXPECT().Get(key, nil).Return(data, nil)

	marshalCache := WrapWithMarshal(c)
	assert.IsType(t, new(marshalWrapper), marshalCache)

	err = marshalCache.Set(key, value, nil)
	assert.Nil(t, err)

	out, err := marshalCache.Get(key, new(testStruct))
	assert.Nil(t, err)
	assert.Equal(t, value, out)
}
