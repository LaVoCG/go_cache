# go_cache
 This is just a simple project to with main intent to learn golang

 code to get started:

 ```
    package main

import (
	"time"
	cache "github.com/LaVoCG/go_cache"
)

func main() {
	cachestruct, closure := cache.NewGenericMemoryCache(time.Duration(1 * time.Second))
	defer closure(cachestruct)
	var genericCache cache.GenericMemoryCache = cachestruct
}
 ```

 where afterwards you can use genericCache variable (or whatever name you choose)