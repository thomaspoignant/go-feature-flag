package examples

import (
	"fmt"
	"testing"
)

func TestPython_ShrinkingVariantSendsAllUsersToOneAdjacentVariant(t *testing.T) {
	seed := "test-flag"
	numUsers := 10_000
	exposureRate := 1.0

	before := map[string]int{"varA": 33, "varB": 34, "varC": 33}
	after := map[string]int{"varA": 45, "varB": 45, "varC": 10}

	type transition struct {
		from string
		to   string
	}
	transitions := make(map[transition]int)

	for i := 0; i < numUsers; i++ {
		id := fmt.Sprintf("user-%d", i)

		varBefore, err := pythonAssign(seed, id, before, exposureRate)
		if err != nil {
			t.Fatalf("unexpected error for before: %v", err)
		}

		varAfter, err := pythonAssign(seed, id, after, exposureRate)
		if err != nil {
			t.Fatalf("unexpected error for after: %v", err)
		}

		if varBefore != varAfter {
			transitions[transition{from: varBefore, to: varAfter}]++
		}
	}

	t.Logf("Transitions when shrinking varC from 33%% to 10%%, increasing varA and varB equally:")
	for tr, count := range transitions {
		t.Logf("  %s -> %s: %d users", tr.from, tr.to, count)
	}

	// With contiguous buckets, lost varC users should only go to the adjacent
	// bucket (varB in alphabetical order). Check if that's the case, or if
	// some also end up in varA.
	t.Logf("varC -> varA: %d users", transitions[transition{from: "varC", to: "varA"}])
	t.Logf("varC -> varB: %d users", transitions[transition{from: "varC", to: "varB"}])
}

func TestPython_TwoVariantsIncreaseSendsNoOneAway(t *testing.T) {
	seed := "test-flag-2var"
	numUsers := 10_000
	exposureRate := 1.0

	before := map[string]int{"varA": 50, "varB": 50}
	after := map[string]int{"varA": 70, "varB": 30}

	type transition struct {
		from string
		to   string
	}
	transitions := make(map[transition]int)

	for i := 0; i < numUsers; i++ {
		id := fmt.Sprintf("user-%d", i)

		varBefore, err := pythonAssign(seed, id, before, exposureRate)
		if err != nil {
			t.Fatalf("unexpected error for before: %v", err)
		}

		varAfter, err := pythonAssign(seed, id, after, exposureRate)
		if err != nil {
			t.Fatalf("unexpected error for after: %v", err)
		}

		if varBefore != varAfter {
			transitions[transition{from: varBefore, to: varAfter}]++
		}
	}

	t.Logf("Transitions when changing from 50/50 to 70/30:")
	for tr, count := range transitions {
		t.Logf("  %s -> %s: %d users", tr.from, tr.to, count)
	}

	if transitions[transition{from: "varB", to: "varA"}] == 0 {
		t.Error("expected some varB users to move to varA, but none did")
	}
	if n := transitions[transition{from: "varA", to: "varB"}]; n != 0 {
		t.Errorf("expected no varA users to move to varB, but %d did", n)
	}
}

func TestPython_ChangingSumCausesWidespreadReshuffle(t *testing.T) {
	seed := "test-flag-sum-change"
	numUsers := 10_000
	exposureRate := 1.0

	before := map[string]int{"varA": 50, "varB": 50}
	after := map[string]int{"varA": 60, "varB": 50}

	changedCount := 0
	for i := 0; i < numUsers; i++ {
		id := fmt.Sprintf("user-%d", i)

		varBefore, err := pythonAssign(seed, id, before, exposureRate)
		if err != nil {
			t.Fatalf("unexpected error for before: %v", err)
		}

		varAfter, err := pythonAssign(seed, id, after, exposureRate)
		if err != nil {
			t.Fatalf("unexpected error for after: %v", err)
		}

		if varBefore != varAfter {
			changedCount++
		}
	}

	changeRate := float64(changedCount) / float64(numUsers)
	t.Logf("When sum changes (100 -> 110): %.1f%% of users changed variant", changeRate*100)

	if changeRate <= 0.05 {
		t.Errorf("expected widespread reshuffling (>5%%) when sum changes, but only %.1f%% changed", changeRate*100)
	}
}

func TestPython_ExposureRateChangeDoesNotAffectVariantAssignment(t *testing.T) {
	seed := "test-flag-exposure"
	numUsers := 10_000

	variants := map[string]int{"varA": 50, "varB": 50}
	lowExposure := 0.1
	highExposure := 0.3

	changedVariant := 0
	bothExposed := 0

	for i := 0; i < numUsers; i++ {
		id := fmt.Sprintf("user-%d", i)

		varLow, err := pythonAssign(seed, id, variants, lowExposure)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		varHigh, err := pythonAssign(seed, id, variants, highExposure)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if varLow != "not exposed" && varHigh != "not exposed" {
			bothExposed++
			if varLow != varHigh {
				changedVariant++
			}
		}
	}

	t.Logf("Users exposed in both configs: %d", bothExposed)
	t.Logf("Of those, users whose variant changed: %d", changedVariant)

	if changedVariant != 0 {
		t.Errorf("expected zero variant changes among users exposed in both configs, but %d changed", changedVariant)
	}
	if bothExposed == 0 {
		t.Error("expected some users to be exposed in both configs")
	}
}
