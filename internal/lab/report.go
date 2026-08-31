package lab

import (
	"fmt"
	"path/filepath"
)

func EvaluateMutants(ir IR, contract Contract, inputDir, reportPath string) (Report, error) {
	if err := ValidateIR(ir); err != nil {
		return Report{}, err
	}
	if err := ValidateDeclarations(SourceDecl{Schema: "gooo/semantic-mutation-lab/source/v1", Version: "v1", DenominatorID: ir.DenominatorID, CellCount: ir.CellCount, Authority: ir.Authority, Precedence: ir.Precedence, UnknownFields: ir.UnknownFields, Cells: ir.Cells, Mutations: ir.Mutations, SourceDigest: ir.SourceDigest}, contract); err != nil {
		return Report{}, err
	}
	if err := ValidateDenominator(contract, ir); err != nil {
		return Report{}, err
	}
	var receipt GenerationReceipt
	if err := ReadJSON(filepath.Join(inputDir, "generation-receipt.json"), &receipt); err != nil {
		return Report{}, err
	}
	if receipt.Schema != ReceiptSchema || receipt.Generated != FixedCells || receipt.Attempted != FixedCells || !receipt.CallerOwnedTempOutput || receipt.RepositoryWrites != 0 || receipt.LocalTestExecutions != 0 || receipt.CrossProjectRequiredGates != 0 {
		return Report{}, fmt.Errorf("generation receipt violates the zero-write boundary")
	}
	report := Report{
		Schema:                      ReportSchema,
		Decision:                    "SEMANTIC_MUTATION_CONFORMANCE_REPORTED",
		FixedConformanceDenominator: FixedCells,
		SourceDigest:                ir.SourceDigest,
		ContractDigest:              ir.ContractDigest,
		IRDigest:                    ir.IRDigest,
		Precedence:                  append([]string(nil), ir.Precedence...),
		Authority: Authority{
			RepositoryWrites:          0,
			LocalTestExecutions:       0,
			CrossProjectRequiredGates: 0,
			RootReadmePolicy:          "EXCLUDED_FROM_REPOSITORY_INVENTORY",
		},
		Improvement: Claim{
			State: "UNKNOWN", Stage: "IMPROVEMENT", Step: "compare_before_after",
			Reason: "EXACT_BEFORE_AFTER_IMPROVEMENT_PAIR_NOT_PROVIDED", UnknownClass: "MISSING_EXACT_PAIR",
			NextOperation: "PROVIDE_EXACT_BEFORE_AFTER_PAIR", BlockedBy: []string{"before-after-evidence"},
		},
		Mutants: make([]MutantResult, 0, FixedCells),
	}
	base := BaselineState(ir)
	for _, mutation := range ir.Mutations {
		path := filepath.Join(inputDir, "mutants", mutation.ID+".json")
		var artifact MutantArtifact
		if err := ReadJSON(path, &artifact); err != nil {
			return Report{}, err
		}
		result := classifyMutant(ir, contract, base, mutation, artifact)
		report.Mutants = append(report.Mutants, result)
		report.Summary.Generated++
		report.Summary.Attempted++
		switch result.State {
		case "KILLED":
			report.Summary.Killed++
		case "SURVIVED_UNKNOWN":
			report.Summary.Unknown++
		case "INVALID_REFUTED":
			report.Summary.Refuted++
		}
	}
	if err := WriteJSON(reportPath, report); err != nil {
		return Report{}, err
	}
	if err := WriteText(stringsTrimSuffix(reportPath, ".json")+".md", RenderReport(report)); err != nil {
		return Report{}, err
	}
	return report, nil
}

func classifyMutant(ir IR, contract Contract, base SemanticState, mutation MutationDecl, artifact MutantArtifact) MutantResult {
	result := MutantResult{
		MutantID:            mutation.ID,
		Family:              mutation.Family,
		Cell:                mutation.Cell,
		State:               "INVALID_REFUTED",
		ChangedSemanticEdge: mutation.ChangedSemanticEdge,
		ExpectedDetector:    mutation.ExpectedDetector,
		ObservedDetector:    "REFUTED",
		ArtifactDigest:      artifact.ArtifactDigest,
		Claim: Claim{
			State: "REFUTED", Stage: "ARTIFACT_VALIDATION", Step: "validate_mutant_artifact",
			Reason: "MUTANT_ARTIFACT_CONTRADICTS_SEMANTIC_INVARIANT", UnknownClass: "",
			NextOperation: "REGENERATE_MUTANT_ARTIFACT", BlockedBy: []string{},
		},
	}
	if reason := validateArtifact(ir, contract, base, mutation, artifact); reason != "" {
		result.Claim.Reason = reason
		result.ObservedDetector = reason
		return result
	}
	// Semantic contradiction and artifact refutation have already been checked.
	// Only a valid evidence gap can reach SURVIVED_UNKNOWN, preserving
	// REFUTED > UNKNOWN > CLOSED precedence.
	if mutation.Family == "FAIL_OPEN_UNKNOWN" {
		if artifact.Mutated.Claims[mutation.Cell].HasUnknownTuple() {
			result.State = "SURVIVED_UNKNOWN"
			result.ObservedDetector = "UNKNOWN_EVIDENCE"
			result.Claim = artifact.Mutated.Claims[mutation.Cell]
			return result
		}
		result.State = "KILLED"
		result.ObservedDetector = mutation.ExpectedDetector
		result.Claim = Claim{State: "REFUTED", Stage: "DETECTOR", Step: "reject_fail_open_unknown", Reason: "UNKNOWN_TUPLE_INCOMPLETE", NextOperation: "RESTORE_UNKNOWN_TUPLE", BlockedBy: []string{}}
		return result
	}
	if mutation.Family == "UNKNOWN_FRONTIER_LOSS" {
		result.Claim.Reason = "UNKNOWN_CLAIM_MISSING_BLOCKED_BY_FRONTIER"
		return result
	}
	if mutation.Family == "STALE_EVIDENCE" && artifact.Mutated.Evidence[mutation.Cell].CommandDigest == "" {
		result.State = "SURVIVED_UNKNOWN"
		result.ObservedDetector = "EVIDENCE_FRESHNESS_UNKNOWN"
		result.Claim = Claim{State: "UNKNOWN", Stage: "EVIDENCE", Step: "verify_freshness", Reason: "FRESHNESS_DIGEST_UNAVAILABLE", UnknownClass: "EVIDENCE_MISSING", NextOperation: "RESTORE_CURRENT_COMMAND_DIGEST", BlockedBy: []string{"freshness-digest"}}
		return result
	}
	if observed := observeDetector(base, artifact.Mutated, mutation, contract); observed != mutation.ExpectedDetector {
		result.ObservedDetector = observed
		result.Claim.State = "UNKNOWN"
		result.State = "SURVIVED_UNKNOWN"
		result.Claim.Stage = "DETECTOR"
		result.Claim.Step = "resolve_detector_observation"
		result.Claim.Reason = "EXPECTED_DETECTOR_OBSERVATION_NOT_AVAILABLE"
		result.Claim.UnknownClass = "DETECTOR_DISAGREEMENT"
		result.Claim.NextOperation = "OBTAIN_INDEPENDENT_DETECTOR_EVIDENCE"
		result.Claim.BlockedBy = []string{"detector-observation"}
		return result
	}
	result.State = "KILLED"
	result.ObservedDetector = mutation.ExpectedDetector
	result.Claim = Claim{
		State: "KILLED", Stage: "DETECTOR", Step: "reject_semantic_regression",
		Reason: "EXPECTED_DETECTOR_REJECTED_CONTROLLED_MUTANT", UnknownClass: "",
		NextOperation: "NONE", BlockedBy: []string{},
	}
	return result
}

func observeDetector(base, mutated SemanticState, mutation MutationDecl, contract Contract) string {
	switch mutation.Family {
	case "SELF_ATTESTATION":
		if mutated.Attestation != base.Attestation {
			return mutation.ExpectedDetector
		}
	case "FIXED_POINT_ABUSE":
		if mutated.Decisions[mutation.Cell] == "FIXED_POINT" {
			return mutation.ExpectedDetector
		}
	case "DENOMINATOR_DELETION":
		if len(mutated.DenominatorCellIDs) != len(base.DenominatorCellIDs) {
			return mutation.ExpectedDetector
		}
	case "AUTHORITY_ESCALATION":
		if mutated.Authority.RepositoryWrites != base.Authority.RepositoryWrites {
			return mutation.ExpectedDetector
		}
	case "STALE_EVIDENCE":
		if mutated.Evidence[mutation.Cell].CommandDigest != base.Evidence[mutation.Cell].CommandDigest {
			return mutation.ExpectedDetector
		}
	case "DIGEST_MISMATCH":
		if mutated.Evidence[mutation.Cell].SubjectDigest != base.Evidence[mutation.Cell].SubjectDigest {
			return mutation.ExpectedDetector
		}
	case "DEPENDENCY_EDGE_REMOVAL":
		if _, ok := mutated.Dependencies[mutation.Cell]; !ok {
			return mutation.ExpectedDetector
		}
	case "SOURCE_IR_DRIFT":
		if mutated.SourceIRBindings[mutation.Cell] != base.SourceIRBindings[mutation.Cell] {
			return mutation.ExpectedDetector
		}
	case "ARTIFACT_SUBSTITUTION":
		if mutated.ArtifactBindings[mutation.ID] != base.ArtifactBindings[mutation.ID] {
			return mutation.ExpectedDetector
		}
	case "REFUTATION_PRECEDENCE_INVERSION":
		if len(mutated.Precedence) > 0 && mutated.Precedence[0] != base.Precedence[0] {
			return mutation.ExpectedDetector
		}
	case "FAIL_OPEN_UNKNOWN":
		return "UNKNOWN"
	case "UNKNOWN_FRONTIER_LOSS":
		return "REFUTED"
	}
	if len(contract.Cells) != FixedCells {
		return "REFUTED"
	}
	return "UNKNOWN"
}

func validateArtifact(ir IR, contract Contract, base SemanticState, mutation MutationDecl, artifact MutantArtifact) string {
	if artifact.Schema != ArtifactSchema || artifact.MutantID != mutation.ID || artifact.Family != mutation.Family || artifact.Cell != mutation.Cell || artifact.ChangedSemanticEdge != mutation.ChangedSemanticEdge || artifact.ExpectedDetector != mutation.ExpectedDetector {
		return "MUTANT_ARTIFACT_IDENTITY_CONTRADICTION"
	}
	if artifact.SourceDigest != ir.SourceDigest || artifact.ContractDigest != ir.ContractDigest || artifact.IRDigest != ir.IRDigest {
		return "STALE_OR_MISMATCHED_SOURCE_CONTRACT_IR_DIGEST"
	}
	digest, err := unsignedArtifactDigest(artifact)
	if err != nil || artifact.ArtifactDigest == "" || digest != artifact.ArtifactDigest {
		return "MUTANT_ARTIFACT_DIGEST_MISMATCH"
	}
	baseDigest, err := DigestValue(artifact.Baseline)
	if err != nil {
		return "BASELINE_STATE_DIGEST_UNAVAILABLE"
	}
	expectedBaseDigest, err := DigestValue(base)
	if err != nil || baseDigest != expectedBaseDigest {
		return "BASELINE_STATE_DRIFT"
	}
	mutatedDigest, err := DigestValue(artifact.Mutated)
	if err != nil || mutatedDigest == baseDigest {
		return "MUTANT_DID_NOT_CHANGE_SEMANTIC_STATE"
	}
	if len(artifact.Mutated.DenominatorCellIDs) > FixedCells || len(artifact.Mutated.DenominatorCellIDs) == 0 {
		return "MUTANT_DENOMINATOR_CARDINALITY_INVALID"
	}
	if mutation.Family == "UNKNOWN_FRONTIER_LOSS" {
		claim := artifact.Mutated.Claims[mutation.Cell]
		if claim.State == "UNKNOWN" && claim.Stage != "" && claim.Step != "" && claim.Reason != "" && claim.UnknownClass != "" && claim.NextOperation != "" && claim.BlockedBy == nil {
			return "UNKNOWN_CLAIM_BLOCKED_BY_FRONTIER_IS_MISSING"
		}
	}
	if len(contract.Cells) != FixedCells || len(contract.Families) != FixedCells {
		return "FIXED_DENOMINATOR_CONTRACT_INVALID"
	}
	return ""
}

func stringsTrimSuffix(path, suffix string) string {
	if len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix {
		return path[:len(path)-len(suffix)]
	}
	return path
}

func Run(sourcePath, contractPath, outputDir string) (Report, error) {
	source, err := ParseSource(sourcePath)
	if err != nil {
		return Report{}, err
	}
	contract, err := LoadContract(contractPath)
	if err != nil {
		return Report{}, err
	}
	if err := ValidateDeclarations(source, contract); err != nil {
		return Report{}, err
	}
	ir, err := BuildIR(source, contract)
	if err != nil {
		return Report{}, err
	}
	if err := ValidateDenominator(contract, ir); err != nil {
		return Report{}, err
	}
	if _, err := GenerateMutants(ir, outputDir); err != nil {
		return Report{}, err
	}
	return EvaluateMutants(ir, contract, outputDir, filepath.Join(outputDir, "mutation-report.json"))
}
