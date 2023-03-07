package openai

import (
	"crypto/sha1"
	"fmt"
	"sync"
)

type SafeMap map[string]*SafeMapShard

type SafeMapShard struct {
	items map[string]interface{}
	mu    *sync.RWMutex
}

func NewSafeMap() SafeMap {
	c := make(SafeMap, 256)
	for i := 0; i < 256; i++ {
		c[fmt.Sprintf("%02x", i)] = &SafeMapShard{
			items: make(map[string]interface{}),
			mu:    new(sync.RWMutex),
		}
	}
	return c
}

func (c SafeMap) get(key string) (item interface{}, exists bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	item, exists = shard.items[key]
	return item, exists
}

func (c SafeMap) set(key string, item interface{}) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.items[key] = item
}

func (c SafeMap) merge(key string, newItemGenerator func(string) interface{}) (interface{}, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	if item, exists := shard.items[key]; exists {
		shard.mu.RUnlock()
		return item, true
	}
	shard.mu.RUnlock()
	shard.mu.Lock()
	defer shard.mu.Unlock()
	if item, exists := shard.items[key]; exists {
		return item, true
	} else {
		//TODO: Check if function
		if newItemGenerator == nil {
			return nil, false
		}
		newItem := newItemGenerator(key)
		shard.items[key] = newItem
		return newItem, false
	}
}

func (c SafeMap) del(key string) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.items, key)
}

func (c SafeMap) getShard(key string) (shard *SafeMapShard) {
	hasher := sha1.New()
	hasher.Write([]byte(key))
	shardKey := fmt.Sprintf("%x", hasher.Sum(nil))[0:2]
	return c[shardKey]
}
