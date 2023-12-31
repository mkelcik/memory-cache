package main

import (
	"fmt"
	memorycache "github.com/mkelcik/memory-cache"
	"sync"
	"time"
)

const (
	sampleSize = 10_000_000
)

func concurrency() {
	cache := memorycache.NewCache[int, int](100, false, 0, 0)

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
	cache := memorycache.NewCache[int, int](100, false, 0, 0)

	start := time.Now()
	for i := 0; i < sampleSize; i++ {
		cache.Set(i, i)
	}
	fmt.Println(time.Since(start).Seconds(), "s")
}

func ttlTest() {
	cache := memorycache.NewCache[int, int](100, false, 10*time.Second, 0)

	start := time.Now()
	for i := 0; i < 10; i++ {
		cache.Set(i, i)
		time.Sleep(2 * time.Second)
	}
	fmt.Println("init time", time.Since(start).Seconds(), "s")

	start = time.Now()
	for i := 0; i < 10; i++ {
		v, _ := cache.Get(i)
		fmt.Println(v)
	}

	time.Sleep(5 * time.Second)
	cache.GCRun()

	for i := 0; i < 10; i++ {
		if v, ok := cache.Get(i); ok {
			fmt.Println(v)
		} else {
			fmt.Println("no value")
		}
	}
	duration := time.Since(start).Seconds()
	fmt.Println("read time", duration, "s", "(", float64(sampleSize)/duration, ")")
}

func gcSchedulerTest() {
	cache := memorycache.NewCache[int, int](100, false, 60*time.Second, 10*time.Second)
	for i := 0; i < 10; i++ {
		cache.Set(i, i)
		fmt.Println("key", i, "added")
		time.Sleep(5 * time.Second)
	}

	for {
		if cache.Len() == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func readSingle() {
	cache := memorycache.NewCache[int, int](100, false, 0, 0)

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
	cache := memorycache.NewCache[int, int](100, false, 0, 0)

	toCache := make(map[int]int, sampleSize)
	for i := 0; i < sampleSize; i++ {
		toCache[i] = i
	}
	start := time.Now()
	cache.SetMany(toCache)
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
	gcSchedulerTest()
}
