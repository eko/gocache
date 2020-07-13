package main

import (
	"fmt"
	"time"

	"github.com/yeqown/gocache/cache"
	"github.com/yeqown/gocache/cache/extension"
	"github.com/yeqown/gocache/store"

	"github.com/allegro/bigcache"
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

	withChain()

	withMetrics()
}

func withMarshal() {
	client, err := bigcache.NewBigCache(
		bigcache.DefaultConfig(5 * time.Minute))
	if err != nil {
		panic(err)
	}
	s := store.NewBigcache(client, nil)
	c := extension.WrapWithMarshal(cache.New(s))

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

	recv := new(user)
	_, err = c.Get("key", recv)
	fmt.Printf("withMarshal output c.Get() val=%+v, err=%v", recv, err)
}

// TODO:
func withChain() {

}

// TODO:
func withMetrics() {

}
