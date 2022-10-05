package cuckoo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
)

// maxCuckooKickouts is the maximum number of times reinsert
// is attempted.
const maxCuckooKickouts = 500

// Filter is a probabilistic counter.
type Filter interface {
	// Lookup returns true if data is in the filter.
	Lookup(data []byte) bool
	// Insert data into the filter. Returns false if insertion failed. In the resulting state, the filter
	// * Might return false negatives
	// * Deletes are not guaranteed to work
	// To increase success rate of inserts, create a larger filter.
	Insert(data []byte) bool
	// Delete data from the filter. Returns true if the data was found and deleted.
	Delete(data []byte) bool
	// Count returns the number of items in the filter.
	Count() uint

	// LoadFactor returns the fraction slots that are occupied.
	LoadFactor() float64
	// Reset removes all items from the filter, setting count to 0.
	Reset()
	// Encode returns a byte slice representing a Cuckoofilter.
	Encode() []byte
}

type filter[T fingerprintsize] struct {
	buckets             []bucket[T]
	fingerprintSizeBits int
	count               uint
	// Bit mask set to len(buckets) - 1. As len(buckets) is always a power of 2,
	// applying this mask mimics the operation x % len(buckets).
	bucketIndexMask        uint
	maxFingerprintMinusOne uint64
}

func numBuckets(numElements uint) uint {
	numBuckets := getNextPow2(uint64(numElements / bucketSize))
	if float64(numElements)/float64(numBuckets*bucketSize) > 0.96 {
		numBuckets <<= 1
	}
	if numBuckets == 0 {
		numBuckets = 1
	}
	return numBuckets
}

func maxFingerprintMinusOne(fingerprintSizeBits int) uint64 {
	return uint64((1 << fingerprintSizeBits) - 2)
}

// NewFilter returns a new cuckoofilter suitable for the given number of elements.
// When inserting more elements, insertion speed will drop significantly and insertions might fail altogether.
// A capacity of 1000000 is a normal default, which allocates
// about ~2MB on 64-bit machines.
func NewFilter(cfg Config) Filter {
	numBuckets := numBuckets(cfg.NumElements)
	switch cfg.Precision {
	case Low:
		return &filter[uint8]{
			buckets:                make([]bucket[uint8], numBuckets),
			count:                  0,
			bucketIndexMask:        uint(numBuckets - 1),
			fingerprintSizeBits:    8,
			maxFingerprintMinusOne: maxFingerprintMinusOne(8),
		}
	case High:
		return &filter[uint32]{
			buckets:                make([]bucket[uint32], numBuckets),
			count:                  0,
			bucketIndexMask:        uint(numBuckets - 1),
			fingerprintSizeBits:    32,
			maxFingerprintMinusOne: maxFingerprintMinusOne(32),
		}
	default:
		return &filter[uint16]{
			buckets:                make([]bucket[uint16], numBuckets),
			count:                  0,
			bucketIndexMask:        uint(numBuckets - 1),
			fingerprintSizeBits:    16,
			maxFingerprintMinusOne: maxFingerprintMinusOne(16),
		}
	}
}

func (cf *filter[T]) Lookup(data []byte) bool {
	i1, fp := getIndexAndFingerprint[T](data, cf.bucketIndexMask, cf.maxFingerprintMinusOne, cf.fingerprintSizeBits)
	if b := cf.buckets[i1]; b.contains(fp) {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.bucketIndexMask)
	b := cf.buckets[i2]
	return b.contains(fp)
}

func (cf *filter[T]) Reset() {
	for i := range cf.buckets {
		cf.buckets[i].reset()
	}
	cf.count = 0
}

func (cf *filter[T]) Insert(data []byte) bool {
	i1, fp := getIndexAndFingerprint[T](data, cf.bucketIndexMask, cf.maxFingerprintMinusOne, cf.fingerprintSizeBits)
	if cf.insertIntoBucket(fp, i1) {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.bucketIndexMask)
	if cf.insertIntoBucket(fp, i2) {
		return true
	}
	return cf.cuckooInsert(fp, i1)
}

func (cf *filter[T]) insertIntoBucket(fp T, i uint) bool {
	if cf.buckets[i].insert(fp) {
		cf.count++
		return true
	}
	return false
}

func (cf *filter[T]) cuckooInsert(fp T, i uint) bool {
	// Apply cuckoo kickouts until a free space is found.
	for k := 0; k < maxCuckooKickouts; k++ {
		j := rand.Intn(bucketSize)
		// Swap fingerprint with bucket entry.
		cf.buckets[i][j], fp = fp, cf.buckets[i][j]

		// Move kicked out fingerprint to alternate location.
		i = getAltIndex(fp, i, cf.bucketIndexMask)
		if cf.insertIntoBucket(fp, i) {
			return true
		}
	}
	return false
}

func (cf *filter[T]) Delete(data []byte) bool {
	i1, fp := getIndexAndFingerprint[T](data, cf.bucketIndexMask, cf.maxFingerprintMinusOne, cf.fingerprintSizeBits)
	i2 := getAltIndex(fp, i1, cf.bucketIndexMask)
	return cf.delete(fp, i1) || cf.delete(fp, i2)
}

func (cf *filter[T]) delete(fp T, i uint) bool {
	if cf.buckets[i].delete(fp) {
		cf.count--
		return true
	}
	return false
}

func (cf *filter[T]) Count() uint {
	return cf.count
}

func (cf *filter[T]) LoadFactor() float64 {
	return float64(cf.count) / float64(len(cf.buckets)*bucketSize)
}

func (cf *filter[T]) Encode() []byte {
	res := new(bytes.Buffer)
	bytesPerBucket := bucketSize * cf.fingerprintSizeBits / 8
	res.Grow(len(cf.buckets)*bytesPerBucket + 1)
	binary.Write(res, binary.LittleEndian, uint8(cf.fingerprintSizeBits))
	for _, b := range cf.buckets {
		for _, fp := range b {
			binary.Write(res, binary.LittleEndian, fp)
		}
	}
	return res.Bytes()
}

// Decode returns a Cuckoofilter from a byte slice created using Encode.
func Decode(data []byte) (Filter, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data can not be empty")
	}
	fingerprintSizeBits, data := int(data[0]), data[1:]
	if len(data)%bucketSize != 0 {
		return nil, fmt.Errorf("data must to be multiple of %d, got %d", bucketSize, len(data))
	}
	bytesPerBucket := bucketSize * fingerprintSizeBits / 8
	if bytesPerBucket == 0 {
		return nil, fmt.Errorf("bytesPerBucket can not be zero")
	}
	numBuckets := len(data) / bytesPerBucket
	if numBuckets < 1 {
		return nil, fmt.Errorf("data can not be smaller than %d, size in bytes is %d", bytesPerBucket, len(data))
	}
	if getNextPow2(uint64(numBuckets)) != uint(numBuckets) {
		return nil, fmt.Errorf("numBuckets must to be a power of 2, got %d", numBuckets)
	}
	switch fingerprintSizeBits {
	case 8:
		return decode[uint8](fingerprintSizeBits, numBuckets, data), nil
	case 16:
		return decode[uint16](fingerprintSizeBits, numBuckets, data), nil
	case 32:
		return decode[uint32](fingerprintSizeBits, numBuckets, data), nil
	default:
		return nil, fmt.Errorf("fingerprint size bits must be 8, 16 or 32, got %d", fingerprintSizeBits)
	}
}

func decode[T fingerprintsize](fingerprintSizeBits, numBuckets int, data []byte) *filter[T] {
	var count uint
	buckets := make([]bucket[T], numBuckets)
	reader := bytes.NewReader(data)
	for i, b := range buckets {
		for j := range b {
			binary.Read(reader, binary.LittleEndian, &buckets[i][j])
			if buckets[i][j] != 0 {
				count++
			}
		}
	}
	return &filter[T]{
		buckets:                buckets,
		count:                  count,
		bucketIndexMask:        uint(len(buckets) - 1),
		fingerprintSizeBits:    fingerprintSizeBits,
		maxFingerprintMinusOne: maxFingerprintMinusOne(fingerprintSizeBits),
	}
}
