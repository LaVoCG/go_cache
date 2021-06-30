package cache

import (
	"sync"
	"time"
)

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

func (gcache *genericMemoryCacheStruct) StartStaleDataCleaner(wg *sync.WaitGroup) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
	inf_loop: //using label for the purpose of breaking out of infinite for loop
		for {
			//using select switch to introduce non-blocking listening on the channels, to have a way to stop goroutine when signal is received
			select {
			case <-gcache.ticker.C:
				gcache.clean()
			case <-gcache.doneCh:
				break inf_loop
			}
		}
		wg.Done()
	}(wg)

}

// constructor function to initialize map data type
func NewGenericMemoryCache(cleaningInterval time.Duration) *genericMemoryCacheStruct {
	cache := &genericMemoryCacheStruct{
		data:   make(map[string]*cacheData),
		ticker: time.NewTicker(cleaningInterval),
		doneCh: make(chan struct{}),
	}
	// start background expired data cleaner
	cache.StartStaleDataCleaner(&wg)
	return cache
}
