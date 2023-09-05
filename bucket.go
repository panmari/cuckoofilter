package cuckoo

import (
	"bytes"
	"fmt"
)

type fingerprintsize interface {
	uint8 | uint16 | uint32
}

// bucket keeps track of fingerprints hashing to the same index.
type bucket[T fingerprintsize] [bucketSize]T

const (
	nullFp     = 0
	bucketSize = 4
)

// insert a fingerprint into a bucket. Returns true if there was enough space and insertion succeeded.
// Note it allows inserting the same fingerprint multiple times.
func (b *bucket[T]) insert(fp T) bool {
	for i, tfp := range b {
		if tfp == nullFp {
			b[i] = fp
			return true
		}
	}
	return false
}

// delete a fingerprint from a bucket.
// Returns true if the fingerprint was present and successfully removed.
func (b *bucket[T]) delete(fp T) bool {
	for i, tfp := range b {
		if tfp == fp {
			b[i] = nullFp
			return true
		}
	}
	return false
}

func (b *bucket[T]) contains(needle T) bool {
	for _, fp := range b {
		if fp == needle {
			return true
		}
	}
	return false
}

// reset deletes all fingerprints in the bucket.
func (b *bucket[T]) reset() {
	for i := range b {
		b[i] = nullFp
	}
}

func (b *bucket[T]) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for _, by := range b {
		buf.WriteString(fmt.Sprintf("%5d ", by))
	}
	buf.WriteString("]")
	return buf.String()
}
