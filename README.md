# go_cache
 This is just a simple project to with main intent to learn golang

 code to get started:

 ```
    cachestruct, closure := NewGenericMemoryCache(time.Duration(1 * time.Second))
	defer closure(cachestruct)
	var genericCache GenericMemoryCache = cachestruct
 ```

 where afterwards you can use genericCache variable (or whatever name you choose)