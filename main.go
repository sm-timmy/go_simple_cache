package main

import (
	context "context"
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	items     map[string]Item
	cleanTime int
	sync.RWMutex
}

func (c *Cache) Get(key string) interface{} {
	if c.Contains(key) {
		return c.items[key].val
	}
	return nil
}

func (c *Cache) Load(key string, val interface{}, ttl int) {
	if !c.Contains(key) {
		c.Lock()
		it := Item{
			val:     val,
			ttl:     ttl,
			created: time.Now(),
		}
		c.items[key] = it
		c.Unlock()
	}
}

func (c *Cache) Contains(key string) bool {
	_, ok := c.items[key]
	if ok {
		return true
	}
	return false
}

func (c *Cache) Cleanup(ctx context.Context) {
	//wait time + cleanTime
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Tick(time.Duration(c.cleanTime))
			now := time.Now()
			for key, it := range c.items {
				if now.Second() > (it.created.Second() + it.ttl) {
					c.Lock()
					delete(c.items, key)
					c.Unlock()
				}
			}
		}
	}
}

type Item struct {
	val     interface{}
	created time.Time
	ttl     int
}

func New(cleanTime int, ctx context.Context) *Cache {
	c := &Cache{
		items:     map[string]Item{},
		cleanTime: cleanTime,
	}
	go c.Cleanup(ctx)
	return c
}

func main() {
	ctx := context.Background()
	c := New(15, ctx)
	c.Load("key", "password1", 10)
	time.Sleep(20 * time.Second)
	fmt.Println(c.Get("key"))
	ctx.Done()
}
