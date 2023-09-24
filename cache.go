package memory_cache

import (
	"sync"
	"time"
)

type Item[T any] struct {
	value   T
	created time.Time
}

type Cache[K comparable, T any] struct {
	storage map[K]Item[T]
	maxSize uint
	lock    sync.RWMutex
	ttl     time.Duration
}

func NewCache[K comparable, T any](maxSize uint, ttl time.Duration) *Cache[K, T] {
	s := map[K]Item[T]{}
	// preallocate map
	if maxSize > 0 {
		s = make(map[K]Item[T], maxSize)
	}

	return &Cache[K, T]{
		storage: s,
		maxSize: maxSize,
		ttl:     ttl,
		lock:    sync.RWMutex{},
	}
}

func (c *Cache[K, T]) Flush() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.storage = make(map[K]Item[T], c.maxSize)
}

func (c *Cache[K, T]) Del(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.storage, key)
}

func (c *Cache[K, T]) GetOrSet(key K, callback func() (T, bool)) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if v, ok := c.getFromMap(key); ok {
		return v, true
	}

	var out T
	var ok bool
	if callback == nil {
		return out, ok
	}

	out, ok = callback()

	c.setToMap(key, out)
	return out, ok
}

func (c *Cache[K, T]) setToMap(key K, value T) {
	c.storage[key] = Item[T]{
		value:   value,
		created: time.Now(),
	}
}

func (c *Cache[K, T]) getFromMap(key K) (T, bool) {
	if i, ok := c.storage[key]; ok && i.created.Add(c.ttl).Before(time.Now()) {
		return i.value, true
	}

	var out T
	return out, false
}

func (c *Cache[K, T]) Set(key K, value T) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.setToMap(key, value)
}

func (c *Cache[K, T]) Get(key K) (T, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.getFromMap(key)
}
