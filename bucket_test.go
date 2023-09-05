package cuckoo

import (
	"reflect"
	"testing"
)

func TestBucket_Reset(t *testing.T) {
	var bkt bucket[uint16]
	for i := uint16(0); i < bucketSize; i++ {
		bkt[i] = i
	}
	bkt.reset()

	var want bucket[uint16]
	if !reflect.DeepEqual(bkt, want) {
		t.Errorf("bucket.reset() got %v, want %v", bkt, want)
	}
}
