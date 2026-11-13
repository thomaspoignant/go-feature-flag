package utils

import "hash/fnv"

// Hash is taking a string and convert.
func Hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// BuildHash is building the hash based on the different properties of the evaluation.
func BuildHash(flagName string, bucketingKey string, maxPercentage uint32) uint32 {
	// this is not supposed to happen, but to avoid a crash if maxPercentage is 0 we are returning 0
	if maxPercentage == uint32(0) {
		return uint32(0)
	}
	// FNV-1a over flagName then bucketingKey, same result as Hash(flagName+bucketingKey)
	// without allocating the concatenated string.
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	h := uint32(offset32)
	for i := 0; i < len(flagName); i++ {
		h ^= uint32(flagName[i])
		h *= prime32
	}
	for i := 0; i < len(bucketingKey); i++ {
		h ^= uint32(bucketingKey[i])
		h *= prime32
	}
	return h % maxPercentage
}
