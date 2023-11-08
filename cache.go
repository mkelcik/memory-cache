package memory_cache

import (
	"fmt"
	"sync"
	"sync/atomic"
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
	length  atomic.Uint64

	size       uint
	strictSize bool

	ttl time.Duration

	lock       sync.RWMutex
	gcInterval time.Duration
	gcDone     chan bool

	first *Item[K, T]
	last  *Item[K, T]
}

func (c *Cache[K, T]) GCSchedulerStart() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.gcInterval == 0 {
		return
	}
	c.gcDone = make(chan bool)

	go func(done <-chan bool) {
		ticker := time.NewTicker(c.gcInterval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				fmt.Println("gc stop")
				return
			case <-ticker.C:
				fmt.Println("gc tick")
				c.GCRun()
			}
		}
	}(c.gcDone)
}

func (c *Cache[K, T]) GCSchedulerStop() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.gcDone == nil {
		return
	}

	select {
	case c.gcDone <- true:
	default:
	}
	close(c.gcDone)
}

func (c *Cache[K, T]) Close() error {
	c.GCSchedulerStop()

	c.lock.Lock()
	defer c.lock.Unlock()

	clear(c.storage)
	c.first = nil
	c.last = nil

	close(c.gcDone)
	return nil
}

func NewCache[K comparable, T any](size uint, strictSize bool, ttl time.Duration, gcInterval time.Duration) *Cache[K, T] {
	s := map[K]*Item[K, T]{}
	// preallocate map
	if size > 0 {
		s = make(map[K]*Item[K, T], size)
	}

	c := &Cache[K, T]{
		storage:    s,
		size:       size,
		strictSize: strictSize,
		ttl:        ttl,
		lock:       sync.RWMutex{},
		gcInterval: gcInterval,
	}
	c.GCSchedulerStart()
	return c
}

func (c *Cache[K, T]) checkAndCleanFirst(now time.Time) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.first == nil {
		return false
	}

	if c.first.created.Add(c.ttl).Before(now) {
		fmt.Println(c.first.key, "deleted")
		c.deleteFromMap(c.first.key)
		return true
	}
	return false
}

func (c *Cache[K, T]) GCRun() {
	now := time.Now()

	for {
		// clear until last element has valid TTL then stop
		if !c.checkAndCleanFirst(now) {
			return
		}
	}
}

func (c *Cache[K, T]) Flush() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.first = nil
	c.last = nil
	c.storage = make(map[K]*Item[K, T], c.size)
	c.length.Store(0)
}

func (c *Cache[K, T]) deleteFromMap(key K) {
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

		c.length.Add(-1)
	}
	delete(c.storage, key)
}

func (c *Cache[K, T]) Len() int64 {
	return int64(c.length.Load())
}

func (c *Cache[K, T]) Del(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.deleteFromMap(key)
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
	c.length.Add(1)

	if c.last != nil {
		c.last.next = item
	}
	c.last = item

	if c.first == nil {
		c.first = item
	}
}

func (c *Cache[K, T]) getFromMap(key K) (T, bool) {
	if i, ok := c.storage[key]; ok && i.created.Add(c.ttl).After(time.Now()) {
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
