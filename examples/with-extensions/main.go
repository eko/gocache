package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/yeqown/gocache"
	"github.com/yeqown/gocache/store"
	"github.com/yeqown/gocache/types"
	"github.com/yeqown/gocache/wrapper"

	"github.com/go-redis/redis/v7"
)

type user struct {
	Name  string
	Age   int
	Embed struct {
		Addr string
	}
}

func main() {
	withMarshal()

	// withChain()
}

func withMarshal() {
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

	// wrap Cache with marshal
	c = wrapper.WrapWithMarshal(c)
	u := user{
		Name: "test",
		Age:  10,
		Embed: struct {
			Addr string
		}{
			"New York. United States",
		},
	}

	if err := c.Set("key", u, nil); err != nil {
		fmt.Printf("set failed, err=%v\n", err)
		return
	}

	data, err := c.Get("key")
	errCheck(err)

	if _, ok := c.(wrapper.IMarshal); !ok {
		errCheck(errors.New("not implement"))
	}

	gotUser := new(user)
	err = c.(wrapper.IMarshal).Unmarshal(data, gotUser)
	errCheck(err)

	fmt.Printf("withMarshal output c.Get() val=%+v, err=%v\n", gotUser, err)
}

//
//func withChain() {
//	cache1, err := gocache.New(gocache.WithBigCacheConfig(bigcache.DefaultConfig(5 * time.Minute)))
//	errCheck(err)
//	cache2, err := gocache.New(gocache.WithRedisOptions(&redis.Options{Addr: "127.0.0.1:6379"}))
//	errCheck(err)
//
//	// Initialize chained cache
//	chain := wrapper.WrapAsChain(cache1, cache2)
//	err = chain.Set("key", []byte("chain value"), nil)
//	errCheck(err)
//
//	out, err := chain.Get("key")
//	errCheck(err)
//	fmt.Printf("withChain from chain out=%s\n", out)
//
//	// check from redis
//	val2, err := cache2.Get("key")
//	errCheck(err)
//	fmt.Printf("withChain from redis out=%s\n", val2)
//
//	// check from big-cache
//	val1, err := cache1.Get("key")
//	errCheck(err)
//	fmt.Printf("withChain from bigcache out=%s\n", val1)
//}

func errCheck(err error) {
	if err == nil {
		return
	}

	fmt.Printf("[Error] errCheck: err=%v\n", err)
	panic(err)
}
