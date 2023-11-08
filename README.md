## Simple thread save memory cache

Features:
* generic
* thread save
* TTL
* optional fixed size (after defined capacity is reached, oldest element is distracted)
* async GC of expired keys

### How to use

```go
package main

import (
	"fmt"
	cache "github.com/mkelcik/memory-cache"
)

type CacheValue struct {
	Value int
}

func main() {
	// define type of key and value
	// first parameter is pre allocation size
	// if second parameter is true, cache is limited to this size, if false the cache can grow beyond this capacity
	// third parameter is ttl for cache records in seconds, if is set to 0 the cache items never expire
	// last parameter is GC interval in seconds, if set to 0, GC will never start automatically
	cache := cache.NewCache[int, CacheValue](100, false, 60, 120)

	// set data to cache
	for i := 1; i <= 100; i++ {
		cache.Set(i, CacheValue{Value: 1})
	}
	
	// read from cache 
	itemFromCache, ok := cache.Get(1)
	if ok {
		fmt.Println("value is:", itemFromCache.Value)
    }
}
```

#### Execute GC manually

If you need more control over destroying expired key, you can run GC manually when you see fit 
```go
cache.GCRun()
```

