package lab

import (
	"fmt"
	"os"
	"strings"
)

func WriteText(path, value string) error {
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func RenderReport(report Report) string {
	var b strings.Builder
	b.WriteString("# Semantic mutation conformance\n\n")
	b.WriteString("Fixed denominator cells: 12\n\n")
	b.WriteString("- generated: ")
	b.WriteString(fmt.Sprint(report.Summary.Generated))
	b.WriteString("\n- attempted: ")
	b.WriteString(fmt.Sprint(report.Summary.Attempted))
	b.WriteString("\n- killed: ")
	b.WriteString(fmt.Sprint(report.Summary.Killed))
	b.WriteString("\n- unknown: ")
	b.WriteString(fmt.Sprint(report.Summary.Unknown))
	b.WriteString("\n- refuted: ")
	b.WriteString(fmt.Sprint(report.Summary.Refuted))
	b.WriteString("\n- repository_writes: ")
	b.WriteString(fmt.Sprint(report.Authority.RepositoryWrites))
	b.WriteString("\n- local_test_executions: ")
	b.WriteString(fmt.Sprint(report.Authority.LocalTestExecutions))
	b.WriteString("\n- cross_project_required_gates: ")
	b.WriteString(fmt.Sprint(report.Authority.CrossProjectRequiredGates))
	b.WriteString("\n- improvement: UNKNOWN (exact before/after pair not provided)\n\n")
	b.WriteString("| mutant id | family | changed semantic edge | expected detector | observed detector | status | artifact digest |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, mutant := range report.Mutants {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n", mutant.MutantID, mutant.Family, mutant.ChangedSemanticEdge, mutant.ExpectedDetector, mutant.ObservedDetector, mutant.State, mutant.ArtifactDigest)
	}
	b.WriteString("\nResolution precedence: REFUTED > UNKNOWN > CLOSED\n")
	return b.String()
}
