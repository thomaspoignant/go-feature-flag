// examples/hash_code.go
package examples

import (
	"fmt"
	"hash/fnv"
	"sort"
)

const percentageMultiplier = float64(1000)

type percentageBucket struct {
	start float64
	end   float64
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func buildHash(flagName, bucketingKey string, maxPercentage uint32) uint32 {
	if maxPercentage == 0 {
		return 0
	}
	return hash(flagName+bucketingKey) % maxPercentage
}

func buildBuckets(percentages map[string]float64) map[string]percentageBucket {
	buckets := make(map[string]percentageBucket, len(percentages))

	names := make([]string, 0, len(percentages))
	for k := range percentages {
		names = append(names, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	for i, name := range names {
		start := float64(0)
		if i != 0 {
			start = buckets[names[i-1]].end
		}
		buckets[name] = percentageBucket{
			start: start,
			end:   start + percentages[name]*percentageMultiplier,
		}
	}
	return buckets
}

func Assign(flagName, bucketingKey string, percentages map[string]float64) (string, error) {
	total := 0.0
	for _, p := range percentages {
		total += p
	}
	maxPercentage := uint32(total * percentageMultiplier)
	h := buildHash(flagName, bucketingKey, maxPercentage)

	for name, bucket := range buildBuckets(percentages) {
		if uint32(bucket.start) <= h && uint32(bucket.end) > h {
			return name, nil
		}
	}
	return "", fmt.Errorf("impossible to find variation")
}