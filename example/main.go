package main

import (
	"fmt"
	memorycache "gitbucket.home/michal/memory-cache"
	"sync"
	"time"
)

const (
	sampleSize = 100_000_000
)

func concurrency() {
	cache := memorycache.NewCache[int, int](100, 0)

	start := time.Now()

	ch := make(chan int, 10_000)

	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				cache.Set(i, i)
			}
		}()
	}

	for i := 0; i < sampleSize; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()

	fmt.Println(time.Since(start).Seconds(), "s")
}

func single() {
	cache := memorycache.NewCache[int, int](100, 0)

	start := time.Now()
	for i := 0; i < sampleSize; i++ {
		cache.Set(i, i)
	}
	fmt.Println(time.Since(start).Seconds(), "s")
}

func readSingle() {
	cache := memorycache.NewCache[int, int](100, 0)

	start := time.Now()
	for i := 0; i < sampleSize; i++ {
		cache.Set(i, i)
	}
	fmt.Println("init time", time.Since(start).Seconds(), "s")

	start = time.Now()
	for i := 0; i < sampleSize; i++ {
		_, _ = cache.Get(i)
	}
	duration := time.Since(start).Seconds()
	fmt.Println("read time", duration, "s", "(", float64(sampleSize)/duration, ")")
}

func readMulti() {
	cache := memorycache.NewCache[int, int](100, 0)

	start := time.Now()
	for i := 0; i < sampleSize; i++ {
		cache.Set(i, i)
	}
	fmt.Println("init time", time.Since(start).Seconds(), "s")

	start = time.Now()
	ch := make(chan int, sampleSize)
	wg := sync.WaitGroup{}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				cache.Get(i)
			}
		}()
	}

	for i := 0; i < sampleSize; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()

	duration := time.Since(start).Seconds()
	fmt.Println("read time", duration, "s", "(", float64(sampleSize)/duration, ")")
}

func main() {
	readMulti()
}
