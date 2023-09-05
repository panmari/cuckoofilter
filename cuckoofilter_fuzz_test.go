//go:build go1.18
// +build go1.18

package cuckoo

import (
	"testing"
)

func filledFilter(cf Filter) Filter {
	cf.Insert([]byte{1})
	cf.Insert([]byte{2})
	cf.Insert([]byte{3})
	cf.Insert([]byte{4})
	cf.Insert([]byte{5})
	cf.Insert([]byte{6})
	cf.Insert([]byte{7})
	cf.Insert([]byte{8})
	cf.Insert([]byte{9})
	return cf
}

func FuzzDecode(f *testing.F) {
	f.Add(filledFilter(NewFilter(Config{NumElements: 10})).Encode())
	f.Add(filledFilter(NewFilter(Config{NumElements: 10, Precision: Low})).Encode())
	f.Add(filledFilter(NewFilter(Config{NumElements: 10, Precision: High})).Encode())
	f.Fuzz(func(t *testing.T, encoded []byte) {
		cache, err := Decode(encoded)
		if err != nil {
			// Construction failed, no need to test further.
			return
		}
		cache.Lookup([]byte("hello"))
		insertOk := cache.Insert([]byte("world"))
		if del := cache.Delete([]byte("world")); insertOk && !del {
			t.Errorf("Failed to delete item.")
		}
	})
}
