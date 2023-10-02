package memory_cache

import (
	"sync"
	"time"
)

type Item[K comparable, T any] struct {
	value   T
	key     K
	created time.Time

	prev *Item[K, T]
	next *Item[K, T]
}

type Cache[K comparable, T any] struct {
	storage map[K]*Item[K, T]
	maxSize uint
	lock    sync.RWMutex
	ttl     time.Duration

	first *Item[K, T]
	last  *Item[K, T]
}

func NewCache[K comparable, T any](maxSize uint, ttl time.Duration) *Cache[K, T] {
	s := map[K]*Item[K, T]{}
	// preallocate map
	if maxSize > 0 {
		s = make(map[K]*Item[K, T], maxSize)
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

	c.first = nil
	c.last = nil
	c.storage = make(map[K]*Item[K, T], c.maxSize)
}

func (c *Cache[K, T]) Del(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// rebuild my list
	item, ok := c.storage[key]
	if ok {
		if item.prev != nil {
			item.prev.next = nil
			if item.next != nil {
				item.prev.next = item.next
			}
		}

		if item.next != nil {
			item.next.prev = item.prev
		}

		// handle first and last pointers
		if c.first != nil && c.first.key == key {
			if c.first.next != nil {
				c.first = c.first.next
			} else {
				c.first = nil
			}
		}

		if c.last != nil && c.last.key == key {
			if c.last.prev != nil {
				c.last = c.last.prev
			} else {
				c.last = nil
			}
		}
	}
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
	item := &Item[K, T]{
		value:   value,
		key:     key,
		created: time.Now(),
		prev:    c.last,
	}

	c.storage[key] = item

	if c.last != nil {
		c.last.next = item
	}
	c.last = item

	if c.first == nil {
		c.first = item
	}
}

func (c *Cache[K, T]) getFromMap(key K) (T, bool) {
	if i, ok := c.storage[key]; ok && i.created.Add(c.ttl).Before(time.Now()) {
		return i.value, true
	}

	var out T
	return out, false
}

func (c *Cache[K, T]) SetMany(items map[K]T) {
	if len(items) == 0 {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	for k, v := range items {
		c.setToMap(k, v)
	}
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
