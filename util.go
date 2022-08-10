package cuckoo

import (
	"encoding/binary"

	metro "github.com/dgryski/go-metro"
)

func getAltIndex[T fingerprintsize](fp T, i uint, bucketIndexMask uint, fingerprintSizeBits int) uint {
	b := make([]byte, fingerprintSizeBits/8)
	binary.LittleEndian.PutUint16(b, uint16(fp))
	hash := uint(metro.Hash64(b, 1337))
	return (i ^ hash) & bucketIndexMask
}

func getFingerprint[T fingerprintsize](hash, maxFingerprint uint64, fingerprintSizeBits int) T {
	// maxFingerprint := uint64((1 << fingerprintSizeBits) - 1)
	// Use most significant bits for fingerprint.
	shifted := hash >> (64 - fingerprintSizeBits)
	// Valid fingerprints are in range [1, maxFingerprint], leaving 0 as the special empty state.
	fp := shifted%(maxFingerprint-1) + 1
	return T(fp)
}

// getIndexAndFingerprint returns the primary bucket index and fingerprint to be used
func getIndexAndFingerprint[T fingerprintsize](data []byte, bucketIndexMask uint, maxFingerprint uint64, fingerprintSize int) (uint, T) {
	hash := metro.Hash64(data, 1337)
	f := getFingerprint[T](hash, maxFingerprint, fingerprintSize)
	// Use least significant bits for deriving index.
	i1 := uint(hash) & bucketIndexMask
	return i1, f
}

func getNextPow2(n uint64) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return uint(n)
}
