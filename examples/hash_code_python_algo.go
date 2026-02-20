package examples

import (
	"fmt"
	"hash/fnv"
	"sort"
)

const pythonNumBuckets = 10_000

func pythonHash32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func pythonAssign(
	testSeed string,
	id string,
	variantAssignments map[string]int,
	exposureRate float64,
) (string, error) {
	s := testSeed + id
	r := int(pythonHash32(s) % uint32(pythonNumBuckets))
	fmt.Println("id", id, "r", r)
	if float64(r)/float64(pythonNumBuckets) >= exposureRate {
		return "not exposed", nil
	}

	totWeights := 0
	for _, w := range variantAssignments {
		totWeights += w
	}
	if totWeights == 0 {
		return "", fmt.Errorf("total weights must be > 0")
	}

	group := r % totWeights
	return pythonDetermineVariant(variantAssignments, group)
}

func pythonDetermineVariant(variantAssignments map[string]int, group int) (string, error) {
	names := make([]string, 0, len(variantAssignments))
	for k := range variantAssignments {
		names = append(names, k)
	}
	sort.Strings(names)

	accumulated := 0
	for _, name := range names {
		accumulated += variantAssignments[name]
		if group < accumulated {
			return name, nil
		}
	}
	return "", fmt.Errorf("variant not assigned: group=%d, accumulated=%d", group, accumulated)
}
