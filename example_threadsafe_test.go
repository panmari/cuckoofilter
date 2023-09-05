package cuckoo_test

import (
	"fmt"
	"sync"

	cuckoo "github.com/panmari/cuckoofilter"
)

// Small wrapper around cuckoo filter making it thread safe.
type threadSafeFilter struct {
	cf cuckoo.Filter
	mu sync.RWMutex
}

func (f *threadSafeFilter) insert(item []byte) {
	// Concurrent inserts need a Write lock.
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cf.Insert(item)
}

func (f *threadSafeFilter) lookup(item []byte) bool {
	// Concurrent lookups need a read lock.
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cf.Lookup(item)
}

func Example_threadSafe() {
	cf := &threadSafeFilter{
		cf: cuckoo.NewFilter(cuckoo.Config{NumElements: 1000}),
	}

	var wg sync.WaitGroup
	// Insert items concurrently...
	for i := byte(0); i < 50; i++ {
		wg.Add(1)
		go func(item byte) {
			defer wg.Done()
			cf.insert([]byte{item})
		}(i)
	}

	// ...while also doing lookups concurrently.
	for i := byte(0); i < 100; i++ {
		wg.Add(1)
		go func(item byte) {
			defer wg.Done()
			// State is not well-defined here, so we can't define expectations.
			cf.lookup([]byte{item})
		}(i)
	}
	wg.Wait()

	// Simple lookups to verify initialization.
	fmt.Println(cf.lookup([]byte{1}))
	fmt.Println(cf.lookup([]byte{99}))

	// Output:
	// true
	// false
}
