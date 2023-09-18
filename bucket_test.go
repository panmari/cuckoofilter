package cuckoo

import (
	"reflect"
	"testing"
)

func TestBucket_Reset(t *testing.T) {
	var bkt bucket
	for i := fingerprint(0); i < bucketSize; i++ {
		bkt.insert(i + 1)
	}

	bkt.reset()

	var want bucket
	if !reflect.DeepEqual(bkt, want) {
		t.Errorf("bucket.reset() got %v, want %v", bkt, want)
	}
}

func TestBucket_Insert(t *testing.T) {
	var bkt bucket
	for i := fingerprint(0); i < bucketSize; i++ {
		if !bkt.insert(i + 1) {
			t.Error("bucket insert failed")
		}
	}
	if bkt.insert(5) {
		t.Error("expected bucket insert to fail after overflow")
	}
}

func TestBucket_Delete(t *testing.T) {
	var bkt bucket
	for i := fingerprint(0); i < bucketSize; i++ {
		bkt.insert(i + 1)
	}

	for i := fingerprint(0); i < bucketSize; i++ {
		if !bkt.delete(i + 1) {
			t.Error("bucket delete failed")
		}
		if !bkt.insert(i + 1) {
			t.Error("bucket insert after delete failed")
		}
	}
}

func TestBucket_Swap(t *testing.T) {
	var bkt bucket
	bkt.insert(123)
	if prev := bkt.swap(3, 321); prev != 123 {
		t.Errorf("swap returned unexpected value %d", prev)
	}
	if !bkt.contains(321) {
		t.Errorf("contains after swap failed")
	}
}

func TestBucket_Contains(t *testing.T) {
	var bkt bucket
	for i := fingerprint(0); i < bucketSize; i++ {
		bkt.insert(i + 1)
	}

	for i := fingerprint(0); i < bucketSize; i++ {
		if !bkt.contains(i + 1) {
			t.Error("bucket contains failed")
		}
	}
}
