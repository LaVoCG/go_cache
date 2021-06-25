package main

import (
	"fmt"
	"log"
	"time"
)

type GenericMemoryCache interface {
	Get(key string) (entry interface{}, found bool)
	Set(key string, data interface{}, ttl time.Duration)
	Delete(key string)
}

type genericMemoryCacheStruct struct {
	data map[string]cacheData
}

type cacheData struct {
	data      interface{}
	keepalive time.Duration
	created   time.Time
}

func (gcache *genericMemoryCacheStruct) Get(key string) (entry interface{}, found bool) {
	res, ok := gcache.data[key]
	entry = res.data
	if !ok {
		return nil, false
	}
	return entry, true
}

func (gcache *genericMemoryCacheStruct) Set(key string, data interface{}, ttl time.Duration) {
	var dataset cacheData = cacheData{
		data:      data,
		keepalive: ttl,
		created:   time.Now(),
	}
	gcache.data[key] = dataset
}

func (gcache *genericMemoryCacheStruct) Delete(key string) {
}

func NewGenericMemoryCache() *genericMemoryCacheStruct {
	return &genericMemoryCacheStruct{
		data: make(map[string]cacheData),
	}
}

func main() {
	var newcache GenericMemoryCache = NewGenericMemoryCache()
	var tmpvalue interface{} = "Just Some Value"
	newcache.Set("stringvalue", tmpvalue, 5*time.Minute)

	// output
	outValue, ok := newcache.Get("stringvalue")
	if !ok {
		log.Print("could not retrieve value: 'stringvalue'")
	}
	fmt.Println("Value returned from cache: ", outValue)
}
