package cuckoo

import (
	metro "github.com/dgryski/go-metro"
)

func getAltIndex[T fingerprintsize](fp T, i uint, bucketIndexMask uint) uint {
	// NOTE(panmari): hash was originally computed as uint(metro.Hash64(fp, 1337)).
	// Multiplying with a constant has a similar effect and is cheaper.
	// 0x5bd1e995 is the hash constant from MurmurHash2
	const murmurConstant = 0x5bd1e995
	hash := uint(fp) * murmurConstant
	return (i ^ hash) & bucketIndexMask
}

func getFingerprint[T fingerprintsize](hash, maxFingerprintMinusOne uint64, fingerprintSizeBits int) T {
	// Use most significant bits for fingerprint.
	shifted := hash >> (64 - fingerprintSizeBits)
	// Valid fingerprints are in range [1, maxFingerprint], leaving 0 as the special empty state.
	fp := shifted%(maxFingerprintMinusOne) + 1
	return T(fp)
}

// getIndexAndFingerprint returns the primary bucket index and fingerprint to be used
func getIndexAndFingerprint[T fingerprintsize](data []byte, bucketIndexMask uint, maxFingerprintMinusOne uint64, fingerprintSize int) (uint, T) {
	hash := metro.Hash64(data, 1337)
	f := getFingerprint[T](hash, maxFingerprintMinusOne, fingerprintSize)
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
