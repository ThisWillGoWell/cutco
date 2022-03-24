package storage

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sync"
)

type lockedMap struct {
	m    map[string]*Cache
	lock *sync.Mutex
}

var caches *lockedMap

func init() {
	caches = &lockedMap{
		m:    make(map[string]*Cache),
		lock: &sync.Mutex{},
	}
}

type Cache struct {
	queries map[string]*Items
	PkSk    map[string][]string
}

func NewCache() *Cache {

	c := Cache{
		queries: make(map[string]*Items),
		PkSk:    make(map[string][]string),
	}

	return &c
}

func hashInput(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return string(h.Sum(nil))
}

func extractCache(ctx context.Context) *Cache {
	rID := ctx.Value("request-id")
	var requestID string
	if rID != nil {
		requestID = rID.(string)
	}

	cache, ok := caches.m[requestID]
	if !ok {
		cache = NewCache()
		caches.m[requestID] = cache
	}

	return cache
}

func Add(ctx context.Context, inputHash string, items Items) {
	caches.lock.Lock()
	defer caches.lock.Unlock()
	cache := extractCache(ctx)

	cache.queries[inputHash] = &items
	var key string
	for _, item := range items {
		key = Item(item).CacheKey()
		cache.PkSk[key] = append(cache.PkSk[key], inputHash)
	}
}

func AddSingle(ctx context.Context, inputHash string, item Item) {
	// why is this so wrong?
	var items Items
	items = append(items, item)
	Add(ctx, inputHash, items)
}

func Get(ctx context.Context, input string) (string, Items) {
	caches.lock.Lock()
	defer caches.lock.Unlock()
	hash := hashInput(input)
	cache := extractCache(ctx)
	if vals, ok := cache.queries[hash]; ok {
		return hash, *vals
	}
	return hash, nil
}

func Bust(ctx context.Context, entry DdbEntry) {
	caches.lock.Lock()
	defer caches.lock.Unlock()
	cache := extractCache(ctx)
	cacheKey := fmt.Sprintf("%s-%s", *entry.PK(), *entry.SK())
	hashes, ok := cache.PkSk[cacheKey]
	if !ok {
		return
	}
	delete(cache.PkSk, cacheKey)
	for _, h := range hashes {
		delete(cache.queries, h)
	}
}
