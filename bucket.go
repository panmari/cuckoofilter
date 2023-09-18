package cuckoo

import (
	"bytes"
	"fmt"
	"math/bits"
)

// fingerprint represents a single entry in a bucket.
type fingerprint uint16

// bucket keeps track of fingerprints hashing to the same index.
type bucket uint64

const (
	bucketSize          = 4
	fingerprintSizeBits = 16
	maxFingerprint      = (1 << fingerprintSizeBits) - 1
)

// insert a fingerprint into a bucket. Returns true if there was enough space and insertion succeeded.
// Note it allows inserting the same fingerprint multiple times.
func (b *bucket) insert(fp fingerprint) bool {
	if i := findZeros(uint64(*b)); i != 0 {
		*b |= bucket(fp) << (bits.Len64(i) - fingerprintSizeBits)
		return true
	}
	return false
}

// delete a fingerprint from a bucket.
// Returns true if the fingerprint was present and successfully removed.
func (b *bucket) delete(fp fingerprint) bool {
	if i := findValue(uint64(*b), uint16(fp)); i != 0 {
		*b &= ^(maxFingerprint << (bits.Len64(i) - fingerprintSizeBits))
		return true
	}
	return false
}

func (b *bucket) swap(i uint64, fp fingerprint) fingerprint {
	p := (*b) >> (i * fingerprintSizeBits) & maxFingerprint
	*b = (*b) & ^(maxFingerprint<<(i*fingerprintSizeBits)) | (bucket(fp) << (i * fingerprintSizeBits))
	return fingerprint(p)
}

func (b *bucket) contains(needle fingerprint) bool {
	return findValue(uint64(*b), uint16(needle)) != 0
}

func (b *bucket) nullsCount() uint {
	return uint(bits.OnesCount64(findZeros(uint64(*b))))
}

// reset deletes all fingerprints in the bucket.
func (b *bucket) reset() {
	*b = 0
}

func (b *bucket) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i := 3; i >= 0; i-- {
		buf.WriteString(fmt.Sprintf("%5d ", ((*b)>>(i*fingerprintSizeBits))&maxFingerprint))
	}
	buf.WriteString("]")
	return buf.String()
}
