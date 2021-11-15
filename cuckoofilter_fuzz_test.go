//go:build go1.18
// +build go1.18

package cuckoo

import (
	"testing"
)

func FuzzDecode(f *testing.F) {
	cf := NewFilter(10)
	cf.Insert([]byte{1})
	cf.Insert([]byte{2})
	cf.Insert([]byte{3})
	cf.Insert([]byte{4})
	cf.Insert([]byte{5})
	cf.Insert([]byte{6})
	cf.Insert([]byte{7})
	cf.Insert([]byte{8})
	cf.Insert([]byte{9})
	f.Add(cf.Encode())
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
