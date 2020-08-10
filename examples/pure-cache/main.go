package main

import (
	"fmt"
	"time"

	"github.com/yeqown/gocache"
	"github.com/yeqown/gocache/store"
	"github.com/yeqown/gocache/types"
	"github.com/yeqown/gocache/wrapper"

	"github.com/go-redis/redis/v7"
)

func main() {
	redisDemo()
}

func redisDemo() {
	cli := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		Username:     "",
		Password:     "",
		DB:           1,
		MaxRetries:   3,
		PoolSize:     50,
		MinIdleConns: 10,
	})
	errCheck(cli.Ping().Err())

	c, err := gocache.New(
		store.NewRedis(cli, nil),
		gocache.WithStoreOption(&types.StoreOptions{
			Expiration: 20 * time.Second,
			Tags:       []string{"local"},
		}),
	)
	errCheck(err)

	c = wrapper.WrapWithMarshal(c)
	type med struct {
		Title string
		Desc  string
		ID    uint32
		Score float64
	}
	value := med{
		Title: "t",
		Desc:  "d",
		ID:    1,
		Score: 1231.12312,
	}

	// Set
	err = c.Set("key", value, &types.StoreOptions{Expiration: time.Hour, Tags: []string{"tag-a", "tag-b"}})
	errCheck(err)

	// Get
	data, err := c.Get("key")
	errCheck(err)

	// 反序列化
	got := new(med)
	err = c.(wrapper.IMarshal).Unmarshal(data, got)
	errCheck(err)
	fmt.Printf("got=%+v, want=%+v", got, value)

	// 过期
	err = c.Invalidate(types.InvalidateOptions{Tags: []string{"tag-a"}})
	errCheck(err)
	// 过期
	err = c.Invalidate(types.InvalidateOptions{Tags: []string{"tag-a"}})
	errCheck(err)
	// 过期
	err = c.Invalidate(types.InvalidateOptions{Tags: []string{"tag-b"}})
	errCheck(err)
	// 过期
	err = c.Invalidate(types.InvalidateOptions{Tags: []string{"tag-c"}})
	errCheck(err)

	// 设置
	err = c.Set("key", value, &types.StoreOptions{Expiration: time.Hour, Tags: []string{"tag-a", "tag-b"}})
	errCheck(err)

	// delete
	err = c.Delete("key")
	errCheck(err)

	_ = c.Set("key1", "jaja", nil)
	_ = c.Set("key2", "jaja", nil)
	_ = c.Set("key3", "jaja", nil)
	_ = c.Set("key4", "jaja", nil)
	_ = c.Set("key5", "jaja", nil)
}

func errCheck(err error) {
	if err == nil {
		return
	}

	panic(err)
}
