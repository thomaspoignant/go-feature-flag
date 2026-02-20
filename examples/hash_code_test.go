package examples

import (
	"fmt"
	"testing"
)

func TestShrinkingVariantSendsAllUsersToOneAdjacentVariant(t *testing.T) {
	flagName := "test-flag"
	numUsers := 10_000

	before := map[string]float64{"varC": 33, "varB": 34, "varA": 33}
	after := map[string]float64{"varC": 10, "varB": 57, "varA": 33}

	type transition struct {
		from string
		to   string
	}
	transitions := make(map[transition]int)

	for i := 0; i < numUsers; i++ {
		key := fmt.Sprintf("user-%d", i)

		varBefore, err := Assign(flagName, key, before)
		if err != nil {
			t.Fatalf("unexpected error for before: %v", err)
		}

		varAfter, err := Assign(flagName, key, after)
		if err != nil {
			t.Fatalf("unexpected error for after: %v", err)
		}

		if varBefore != varAfter {
			transitions[transition{from: varBefore, to: varAfter}]++
		}
	}

	t.Logf("Transitions when shrinking varC from 33%% to 10%%:")
	for tr, count := range transitions {
		t.Logf("  %s -> %s: %d users", tr.from, tr.to, count)
	}

	if transitions[transition{from: "varC", to: "varB"}] == 0 {
		t.Error("expected some varC users to move to varB, but none did")
	}
	if n := transitions[transition{from: "varC", to: "varA"}]; n != 0 {
		t.Errorf("expected no varC users to move to varA, but %d did (all should go to varB)", n)
	}
	if n := transitions[transition{from: "varA", to: "varB"}]; n != 0 {
		t.Errorf("expected varA users to stay, but %d moved to varB", n)
	}
	if n := transitions[transition{from: "varA", to: "varC"}]; n != 0 {
		t.Errorf("expected varA users to stay, but %d moved to varC", n)
	}
	if n := transitions[transition{from: "varB", to: "varA"}]; n != 0 {
		t.Errorf("expected varB users to stay, but %d moved to varA", n)
	}
	if n := transitions[transition{from: "varB", to: "varC"}]; n != 0 {
		t.Errorf("expected varB users to stay, but %d moved to varC", n)
	}
}

func TestTwoVariantsIncreaseSendsNoOneAway(t *testing.T) {
	flagName := "test-flag-2var"
	numUsers := 10_000

	before := map[string]float64{"varA": 50, "varB": 50}
	after := map[string]float64{"varA": 70, "varB": 30}

	type transition struct {
		from string
		to   string
	}
	transitions := make(map[transition]int)

	for i := 0; i < numUsers; i++ {
		key := fmt.Sprintf("user-%d", i)

		varBefore, err := Assign(flagName, key, before)
		if err != nil {
			t.Fatalf("unexpected error for before: %v", err)
		}

		varAfter, err := Assign(flagName, key, after)
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

func TestChangingSumCausesWidespreadReshuffle(t *testing.T) {
	flagName := "test-flag-sum-change"
	numUsers := 10_000

	before := map[string]float64{"varA": 50, "varB": 50}
	after := map[string]float64{"varA": 60, "varB": 50}

	changedCount := 0
	for i := 0; i < numUsers; i++ {
		key := fmt.Sprintf("user-%d", i)

		varBefore, err := Assign(flagName, key, before)
		if err != nil {
			t.Fatalf("unexpected error for before: %v", err)
		}

		varAfter, err := Assign(flagName, key, after)
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
