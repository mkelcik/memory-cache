package main

import (
	"fmt"
	memorycache "gitbucket.home/michal/memory-cache"
	"sync"
)

func main() {
	cache := memorycache.NewCache[int, int](100, 0)
	cache.Set(0, 1)
	cache.Set(1, 2)

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			fmt.Println(cache.Get(key))
			cache.Set(key, 5)
		}(i % 2)
	}

	wg.Wait()
}
