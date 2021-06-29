package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type TestDataItem struct {
	inputkey     string
	inputvalue   interface{}
	inputttl     time.Duration
	expectedFail bool
}

//for testing purposes using global interface
var genericCache GenericMemoryCache
var testwg = sync.WaitGroup{}

//helper function to find a record in datastruct without invoking a testable functionality
func searchRecord(keyToFind string, valueToFind interface{}) bool {
	datastruct, ok := genericCache.(*genericMemoryCacheStruct)
	if ok {
		for key, value := range datastruct.data {
			if key == keyToFind && value.data == valueToFind {
				return true
			}
		}
	}
	return false
}

//setup the structs/interface for the test to work with. any other additional setup should be done on each test case
func setupTestCase(t *testing.T, interval time.Duration) func(t *testing.T) {
	t.Log("SetUp Test Case")

	genericCache = &genericMemoryCacheStruct{
		data:   make(map[string]*cacheData),
		ticker: time.NewTicker(interval),
		doneCh: make(chan struct{}),
	}

	//teardown
	return func(t *testing.T) {
		t.Log("Teardown Test Case")
		datastruct, ok := genericCache.(*genericMemoryCacheStruct)
		if ok {
			close(datastruct.doneCh)
			datastruct.ticker.Stop()
			for key := range datastruct.data {
				delete(datastruct.data, key)
			}
			datastruct.data = nil
		}
		genericCache = nil
	}
}
func TestSet(t *testing.T) {
	//setup and teardown test data
	//duration time does not matter in this test case as automatic cleanup is not enabled in setup, but should be added for consistency
	teardownTestCase := setupTestCase(t, time.Duration(1*time.Second))
	defer teardownTestCase(t)
	var found bool
	var testCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(1 * time.Microsecond), false},
		{"integerkey", int(64), time.Duration(1 * time.Microsecond), false},
		{"float32key", float32(64.00), time.Duration(1 * time.Microsecond), false},
		{"float64key", float64(64.00), time.Duration(1 * time.Microsecond), false},
		{"arraykey", [...]int{1, 2, 3, 4}, time.Duration(1 * time.Microsecond), false},
	}

	for _, value := range testCases {
		genericCache.Set(value.inputkey, value.inputvalue, value.inputttl)
		found = searchRecord(value.inputkey, value.inputvalue)
		if !found {
			t.Error(fmt.Sprintf("Value setting failed. Value not found. expected to find key %s and value %s", value.inputkey, value.inputvalue))
		}
	}

	//test for supplying data in random intervals using goroutines to check concurency safety using similar testable, with just changed keys
	var testCasesGoroutines = []TestDataItem{
		{"goteststring", "Just some random test string", time.Duration(1 * time.Microsecond), false},
		{"gointegerkey", int(64), time.Duration(1 * time.Microsecond), false},
		{"gofloat32key", float32(64.00), time.Duration(1 * time.Microsecond), false},
		{"gofloat64key", float64(64.00), time.Duration(1 * time.Microsecond), false},
		{"goarraykey", [...]int{1, 2, 3, 4}, time.Duration(1 * time.Microsecond), false},
	}
	rand.Seed(time.Now().UnixNano())
	for _, value := range testCasesGoroutines {
		testwg.Add(1)
		go func(value TestDataItem) {
			//random sleep time to simulate unknown processing time
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			genericCache.Set(value.inputkey, value.inputvalue, value.inputttl)
			testwg.Done()
		}(value)
	}
	testwg.Wait()

	found = false
	//test if goroutines wrote the data to cache
	for _, value := range testCasesGoroutines {
		found = false
		found = searchRecord(value.inputkey, value.inputvalue)
		if !found {
			t.Error("Value setting in goroutines failed. Value not found. expected to find key %s and value %s", value.inputkey, value.inputvalue)
		}
	}
}

func TestGet(t *testing.T) {
	teardownTestCase := setupTestCase(t, time.Duration(1*time.Second))
	defer teardownTestCase(t)
	// var found bool
	var setupCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(1 * time.Microsecond), false},
		{"integerkey", int(64), time.Duration(1 * time.Microsecond), false},
		{"float32key", float32(64.00), time.Duration(1 * time.Microsecond), false},
		{"float64key", float64(64.00), time.Duration(1 * time.Microsecond), false},
		{"arraykey", [...]int{1, 2, 3, 4}, time.Duration(1 * time.Microsecond), false},
	}

	for _, value := range setupCases {
		genericCache.Set(value.inputkey, value.inputvalue, value.inputttl)
	}

	//start of the test section
	var testCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(1 * time.Microsecond), false},
		{"integerkey", int(64), time.Duration(1 * time.Microsecond), false},
		{"float32key", float32(64.00), time.Duration(1 * time.Microsecond), false},
		{"float64key", float64(64.00), time.Duration(1 * time.Microsecond), false},
		{"arraykey", [...]int{1, 2, 3, 4}, time.Duration(1 * time.Microsecond), false},
		{"nonexisting", 0, time.Duration(1 * time.Microsecond), true},
	}
	for _, value := range testCases {
		entry, ok := genericCache.Get(value.inputkey)
		if ok {
			if entry != value.inputvalue {
				t.Error("Value getting failed. Value is not correct. expected to find value %s, got %s", value.inputvalue, entry)
			}
		} else {
			if !value.expectedFail {
				t.Errorf("Value getting failed. Value not found. expected to find key %s", value.inputkey)
			}
		}
	}

	//test for concurrent reads
	for _, value := range testCases {
		go func(value TestDataItem, cache GenericMemoryCache) {
			entry, ok := genericCache.Get(value.inputkey)
			if ok {
				if entry != value.inputvalue {
					t.Errorf("Value retrieval test FAILED. Value is not correct. expected to find value %s, got %s", value.inputvalue, entry)
				}
			} else {
				t.Errorf("Key retrieval test FAILED. Key not found. expected to find key %s", value.inputkey)
			}
		}(value, genericCache)

	}

}

func TestDelete(t *testing.T) {
	teardownTestCase := setupTestCase(t, time.Duration(1*time.Second))
	defer teardownTestCase(t)
	// var found bool
	var setupCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(1 * time.Microsecond), false},
		{"integerkey", int(64), time.Duration(1 * time.Microsecond), false},
		{"float32key", float32(64.00), time.Duration(1 * time.Microsecond), false},
		{"float64key", float64(64.00), time.Duration(1 * time.Microsecond), false},
		{"arraykey", [...]int{1, 2, 3, 4}, time.Duration(1 * time.Microsecond), false},
	}

	for _, value := range setupCases {
		genericCache.Set(value.inputkey, value.inputvalue, value.inputttl)
	}

	var testCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(1 * time.Microsecond), false},
		{"integerkey", int(64), time.Duration(1 * time.Microsecond), false},
		{"float32key", float32(64.00), time.Duration(1 * time.Microsecond), false},
		{"float64key", float64(64.00), time.Duration(1 * time.Microsecond), false},
		{"arraykey", [...]int{1, 2, 3, 4}, time.Duration(1 * time.Microsecond), false},
		{"nonexisting", 0, time.Duration(1 * time.Microsecond), false}, //this should still pass as there is no need to try and delete anything if it does not exist
	}
	//start testing
	for _, value := range testCases {
		genericCache.Delete(value.inputkey)
		_, ok := genericCache.Get(value.inputkey)
		if ok {
			t.Errorf("Value deletion test FAILED. found key: %s, when it was not supposed to exist", value.inputkey)
		}
	}
}

func TestClean(t *testing.T) {
	teardownTestCase := setupTestCase(t, time.Duration(1*time.Second))
	defer teardownTestCase(t)
	// var found bool
	var setupCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(10 * time.Second), false},
		{"integerkey", int(64), time.Duration(20 * time.Second), false},
	}

	for _, value := range setupCases {
		genericCache.Set(value.inputkey, value.inputvalue, value.inputttl)
	}
	if datastruct, ok := genericCache.(*genericMemoryCacheStruct); ok {
		datastruct.clean() //after this it should still be full dataset
		if len(datastruct.data) != 2 {
			t.Errorf("dataset length test FAILED. expected dataset length %d, got %d", 2, len(datastruct.data))
		}

		time.Sleep(time.Duration(10 * time.Second))
		datastruct.clean()
		if len(datastruct.data) != 1 {
			t.Errorf("dataset length test FAILED. expected dataset length %d, got %d", 1, len(datastruct.data))
		}

		time.Sleep(time.Duration(10 * time.Second))
		datastruct.clean()
		if len(datastruct.data) != 0 {
			t.Errorf("dataset length test FAILED. expected dataset length %d, got %d", 0, len(datastruct.data))
		}
	}
}

func TestStartStaleDataCleaner(t *testing.T) {
	teardownTestCase := setupTestCase(t, time.Duration(1*time.Second))

	datastruct, ok := genericCache.(*genericMemoryCacheStruct)
	if ok {
		datastruct.StartStaleDataCleaner(&testwg)
	}
	defer teardownTestCase(t)
	// var found bool
	var setupCases = []TestDataItem{
		{"teststring", "Just some random test string", time.Duration(1 * time.Second), false},
		{"integerkey", int(64), time.Duration(2 * time.Microsecond), false},
		{"float32key", float32(64.00), time.Duration(5 * time.Microsecond), false},
		{"float64key", float64(64.00), time.Duration(10 * time.Microsecond), false},
		{"arraykey", [...]int{1, 2, 3, 4}, time.Duration(15 * time.Microsecond), false},
	}

	for _, value := range setupCases {
		genericCache.Set(value.inputkey, value.inputvalue, value.inputttl)
	}

}
