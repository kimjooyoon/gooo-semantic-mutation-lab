package lab

import (
	"path/filepath"
	"testing"
)

func TestRunProducesFixedCountsAndExplicitUnknown(t *testing.T) {
	root := filepath.Join("..", "..")
	output := t.TempDir()
	report, err := Run(
		filepath.Join(root, "examples", "semantic-mutation-lab-v1", "lab.gooo"),
		filepath.Join(root, "contracts", "denominator-v1.json"),
		output,
	)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary != (Summary{Generated: 12, Attempted: 12, Killed: 10, Unknown: 1, Refuted: 1}) {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	for _, mutant := range report.Mutants {
		if mutant.State == "SURVIVED_UNKNOWN" && !mutant.Claim.HasUnknownTuple() {
			t.Fatalf("unknown claim is incomplete: %+v", mutant.Claim)
		}
	}
}
