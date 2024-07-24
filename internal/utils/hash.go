package utils

import "hash/fnv"

// Hash is taking a string and convert.
func Hash(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	// if we have a problem to get the hash we return 0
	if err != nil {
		return 0
	}
	return h.Sum32()
}

// BuildHash is building the hash based on the different properties of the evaluation.
func BuildHash(flagName string, bucketingKey string, max uint32) uint32 {
	return Hash(flagName+bucketingKey) % max
}
