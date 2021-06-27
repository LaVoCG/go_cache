package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

var lock = sync.RWMutex{}
var wg = sync.WaitGroup{}

type GenericMemoryCache interface {
	Get(key string) (entry interface{}, found bool)
	Set(key string, data interface{}, ttl time.Duration)
	Delete(key string)
}

type genericMemoryCacheStruct struct {
	sync.RWMutex
	data   map[string]*cacheData
	ticker *time.Ticker
	doneCh chan struct{} // channel with empty struct, cause that does not require memory initialization. but channel notifies that a message was sent
}

type cacheData struct {
	data      interface{}
	expiresAt int64 // time will be saved as UNIX
}

func (gcache *genericMemoryCacheStruct) Get(key string) (entry interface{}, found bool) {
	// lock the mutex to make sure only one goroutine can access the data at a given time
	gcache.RLock()
	defer gcache.RUnlock()
	res, ok := gcache.data[key]
	if !ok {
		return nil, false
	}
	return res.data, true
}

func (gcache *genericMemoryCacheStruct) Set(key string, data interface{}, ttl time.Duration) {
	// lock the mutex to make sure only one goroutine can access the data at a given time
	gcache.Lock()
	defer gcache.Unlock() // defer the mutex unlocking, as a syntactic sugar, to avoid the chance of forgetting to add Unlock function

	gcache.data[key] = &cacheData{
		data:      data,
		expiresAt: time.Now().Add(ttl).Unix(),
	}

}

func (gcache *genericMemoryCacheStruct) Delete(key string) {
	// lock the mutex to make sure only one goroutine can access the data at a given time
	gcache.Lock()
	defer gcache.Unlock()
	_, ok := gcache.data[key] // check if the record actually exists, before attempting to delete it
	if ok {
		delete(gcache.data, key)
	}
}

func (gcache *genericMemoryCacheStruct) clean() {
	now := time.Now().Unix()

	gcache.RLock()
	defer gcache.RUnlock()

	for key, el := range gcache.data {
		if now >= el.expiresAt {
			delete(gcache.data, key)
		}
	}
}

// constructor function to initialize map data type
func NewGenericMemoryCache(cleaningInterval time.Duration) *genericMemoryCacheStruct {
	cache := &genericMemoryCacheStruct{
		data:   make(map[string]*cacheData),
		ticker: time.NewTicker(cleaningInterval),
		doneCh: make(chan struct{}),
	}
	// start background expired data cleaner and add it to waitgroup
	wg.Add(1)
	cache.StartStaleDataCleaner()
	return cache
}

func supplyData(cache GenericMemoryCache) {

	rand.Seed(time.Now().UnixNano())
	forrange := rand.Intn(50) + 10
	for i := 0; i < forrange; i++ {
		//random sleep time to simulate unknown processing time
		go func() {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		}()
	}
}

func (gcache *genericMemoryCacheStruct) StartStaleDataCleaner() {
	wg.Add(1)
	go func() {
		for {
			//using select switch to introduce non-blocking listening on the channels, to have a way to stop goroutine when
			select {
			case <-gcache.ticker.C:
				gcache.clean()
			case <-gcache.doneCh:
				wg.Done()
				break
			}
		}
	}()

}

func cleanup(cache GenericMemoryCache) {
	datastruct, ok := cache.(*genericMemoryCacheStruct)
	if ok {
		close(datastruct.doneCh)
	}
}

func main() {
	var newcache GenericMemoryCache = NewGenericMemoryCache(time.Duration(2 * time.Second))

	defer cleanup(newcache)
	var tmpvalue interface{} = "Just Some Random Value"
	newcache.Set("stringvalue", tmpvalue, 5*time.Minute)

	outValue, ok := newcache.Get("stringvalue")
	if !ok {
		log.Print("could not retrieve value: 'stringvalue'")
	}
	fmt.Println("Value returned from cache: ", outValue)

	//adding a simple goroutine, to gracefully shutdown after certain amount of time.
	wg.Add(1)
	go func() {
		time.Sleep(time.Duration(20 * time.Second))
		cleanup(newcache)
	}()

	wg.Wait()
}
