package main

import (
	"fmt"
	"time"

	"github.com/allegro/bigcache"
	"github.com/yeqown/gocache/cache"
	"github.com/yeqown/gocache/store"
)

func main() {
	client, err := bigcache.NewBigCache(
		bigcache.DefaultConfig(5 * time.Minute))
	if err != nil {
		panic(err)
	}
	s := store.NewBigcache(client, nil)
	c := cache.New(s)

	if err := c.Set("key", []byte("thisismyvalue"), nil); err != nil {
		fmt.Printf("set failed, err=%v\n", err)
		return
	}

	val, err := c.Get("key", nil)
	fmt.Printf("c.Get() val=%s, err=%v", val, err)
}
