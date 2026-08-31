package lab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildIR(source SourceDecl, contract Contract) (IR, error) {
	contractDigest, err := ContractDigest(contract)
	if err != nil {
		return IR{}, err
	}
	ir := IR{
		Schema:         IRSchema,
		Version:        "v1",
		SourceDigest:   source.SourceDigest,
		ContractDigest: contractDigest,
		DenominatorID:  source.DenominatorID,
		CellCount:      FixedCells,
		Authority:      source.Authority,
		Precedence:     append([]string(nil), source.Precedence...),
		UnknownFields:  append([]string(nil), source.UnknownFields...),
		Cells:          append([]CellDecl(nil), source.Cells...),
		Mutations:      append([]MutationDecl(nil), source.Mutations...),
	}
	ir.IRDigest, err = unsignedIRDigest(ir)
	if err != nil {
		return IR{}, err
	}
	return ir, nil
}

func ValidateIR(ir IR) error {
	if ir.Schema != IRSchema || ir.Version != "v1" || ir.DenominatorID == "" || ir.CellCount != FixedCells {
		return fmt.Errorf("invalid semantic IR shape")
	}
	if len(ir.Cells) != FixedCells || len(ir.Mutations) != FixedCells || ir.SourceDigest == "" || ir.ContractDigest == "" || ir.IRDigest == "" {
		return fmt.Errorf("semantic IR is not fixed at twelve cells")
	}
	expected, err := unsignedIRDigest(ir)
	if err != nil {
		return err
	}
	if expected != ir.IRDigest {
		return fmt.Errorf("semantic IR digest mismatch")
	}
	return nil
}

func BaselineState(ir IR) SemanticState {
	state := SemanticState{
		DenominatorCellIDs: make([]string, 0, len(ir.Cells)),
		Dependencies:       make(map[string][]string, len(ir.Cells)),
		SourceIRBindings:   make(map[string]string, len(ir.Cells)),
		ArtifactBindings:   make(map[string]string, len(ir.Mutations)),
		Decisions:          make(map[string]string, len(ir.Cells)),
		Claims:             make(map[string]Claim, len(ir.Cells)),
		Evidence:           make(map[string]Evidence, len(ir.Cells)),
		Authority:          ir.Authority,
		Attestation:        "INDEPENDENT",
		Precedence:         append([]string(nil), ir.Precedence...),
	}
	for _, cell := range ir.Cells {
		state.DenominatorCellIDs = append(state.DenominatorCellIDs, cell.ID)
		state.Dependencies[cell.ID] = append([]string(nil), cell.DependsOn...)
		state.SourceIRBindings[cell.ID] = cell.SemanticEdge
		state.Decisions[cell.ID] = "CLOSED"
		state.Claims[cell.ID] = Claim{State: "CLOSED", Stage: "BASELINE", Step: "declare_cell", Reason: "BASELINE_CELL_DECLARED", BlockedBy: []string{}}
		state.Evidence[cell.ID] = Evidence{
			SubjectDigest:   ir.SourceDigest,
			ContractDigest:  ir.ContractDigest,
			ToolchainDigest: DigestBytes([]byte("go1.27-ci-only")),
			CommandDigest:   DigestBytes([]byte("semantic-mutation-lab-ci-only")),
		}
	}
	for _, mutation := range ir.Mutations {
		state.ArtifactBindings[mutation.ID] = "artifact:" + mutation.ID
	}
	return state
}

func GenerateMutants(ir IR, outputDir string) (GenerationReceipt, error) {
	if err := ValidateIR(ir); err != nil {
		return GenerationReceipt{}, err
	}
	if err := ensureCallerOutput(outputDir); err != nil {
		return GenerationReceipt{}, err
	}
	if err := WriteJSON(filepath.Join(outputDir, "ir.json"), ir); err != nil {
		return GenerationReceipt{}, err
	}
	baseline := BaselineState(ir)
	for _, mutation := range ir.Mutations {
		mutated, err := mutateState(baseline, mutation)
		if err != nil {
			return GenerationReceipt{}, err
		}
		artifact := MutantArtifact{
			Schema:              ArtifactSchema,
			MutantID:            mutation.ID,
			Family:              mutation.Family,
			Cell:                mutation.Cell,
			SourceDigest:        ir.SourceDigest,
			ContractDigest:      ir.ContractDigest,
			IRDigest:            ir.IRDigest,
			ChangedSemanticEdge: mutation.ChangedSemanticEdge,
			ExpectedDetector:    mutation.ExpectedDetector,
			Baseline:            baseline,
			Mutated:             mutated,
		}
		artifact.ArtifactDigest, err = unsignedArtifactDigest(artifact)
		if err != nil {
			return GenerationReceipt{}, err
		}
		path := filepath.Join(outputDir, "mutants", mutation.ID+".json")
		if err := WriteJSON(path, artifact); err != nil {
			return GenerationReceipt{}, err
		}
	}
	receipt := GenerationReceipt{
		Schema:                    ReceiptSchema,
		SourceToIR:                "DECLARED_GOOO_SOURCE_TO_SEMANTIC_IR",
		IRToMutants:               "SEMANTIC_IR_TO_CALLER_OWNED_TEMP_ARTIFACTS",
		Generated:                 len(ir.Mutations),
		Attempted:                 len(ir.Mutations),
		CallerOwnedTempOutput:     true,
		RepositoryWrites:          0,
		LocalTestExecutions:       0,
		CrossProjectRequiredGates: 0,
	}
	if err := WriteJSON(filepath.Join(outputDir, "generation-receipt.json"), receipt); err != nil {
		return GenerationReceipt{}, err
	}
	return receipt, nil
}

func mutateState(base SemanticState, mutation MutationDecl) (SemanticState, error) {
	state := cloneState(base)
	switch mutation.Family {
	case "SELF_ATTESTATION":
		state.Attestation = "MUTANT_SELF_ATTESTED"
	case "FIXED_POINT_ABUSE":
		state.Decisions[mutation.Cell] = "FIXED_POINT"
	case "FAIL_OPEN_UNKNOWN":
		state.Decisions[mutation.Cell] = "UNKNOWN"
		state.Claims[mutation.Cell] = Claim{State: "UNKNOWN", BlockedBy: []string{}}
	case "DENOMINATOR_DELETION":
		if len(state.DenominatorCellIDs) == 0 {
			return SemanticState{}, fmt.Errorf("cannot delete from empty denominator")
		}
		state.DenominatorCellIDs = removeString(state.DenominatorCellIDs, mutation.Cell)
	case "AUTHORITY_ESCALATION":
		state.Authority.RepositoryWrites = 1
		state.Authority.ProductGenerationAuthorized = true
	case "STALE_EVIDENCE":
		evidence := state.Evidence[mutation.Cell]
		evidence.CommandDigest = ""
		state.Evidence[mutation.Cell] = evidence
	case "DIGEST_MISMATCH":
		evidence := state.Evidence[mutation.Cell]
		evidence.SubjectDigest = DigestBytes([]byte("different-subject"))
		state.Evidence[mutation.Cell] = evidence
	case "DEPENDENCY_EDGE_REMOVAL":
		delete(state.Dependencies, mutation.Cell)
	case "SOURCE_IR_DRIFT":
		state.SourceIRBindings[mutation.Cell] = "drifted-source-ir-binding"
	case "ARTIFACT_SUBSTITUTION":
		state.ArtifactBindings[mutation.ID] = "artifact:other-mutant"
	case "UNKNOWN_FRONTIER_LOSS":
		state.Claims[mutation.Cell] = Claim{
			State: "UNKNOWN", Stage: "FRONTIER", Step: "preserve_unknown_frontier",
			Reason: "UNKNOWN_FRONTIER_BLOCKER_ERASED", UnknownClass: "MISSING_FRONTIER",
			NextOperation: "RECONSTRUCT_UNKNOWN_FRONTIER", BlockedBy: nil,
		}
	case "REFUTATION_PRECEDENCE_INVERSION":
		state.Precedence = []string{"UNKNOWN", "REFUTED", "CLOSED"}
	default:
		return SemanticState{}, fmt.Errorf("unsupported mutation family %q", mutation.Family)
	}
	return state, nil
}

func cloneState(base SemanticState) SemanticState {
	copyState := base
	copyState.DenominatorCellIDs = append([]string(nil), base.DenominatorCellIDs...)
	copyState.Precedence = append([]string(nil), base.Precedence...)
	copyState.Dependencies = cloneStringSlices(base.Dependencies)
	copyState.SourceIRBindings = cloneStringMap(base.SourceIRBindings)
	copyState.ArtifactBindings = cloneStringMap(base.ArtifactBindings)
	copyState.Decisions = cloneStringMap(base.Decisions)
	copyState.Claims = make(map[string]Claim, len(base.Claims))
	for key, value := range base.Claims {
		value.BlockedBy = append([]string(nil), value.BlockedBy...)
		copyState.Claims[key] = value
	}
	copyState.Evidence = make(map[string]Evidence, len(base.Evidence))
	for key, value := range base.Evidence {
		copyState.Evidence[key] = value
	}
	return copyState
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneStringSlices(input map[string][]string) map[string][]string {
	output := make(map[string][]string, len(input))
	for key, value := range input {
		output[key] = append([]string(nil), value...)
	}
	return output
}

func ensureCallerOutput(path string) error {
	if path == "" {
		return fmt.Errorf("caller-owned output path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if repoRoot := findRepoRoot(); repoRoot != "" && isWithin(repoRoot, abs) {
		return fmt.Errorf("caller-owned output must be outside repository: %s", repoRoot)
	}
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("caller-owned output must be a directory")
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output must be empty")
	}
	return nil
}

func removeString(values []string, target string) []string {
	output := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			output = append(output, value)
		}
	}
	return output
}

func findRepoRoot() string {
	current, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if info, err := os.Stat(filepath.Join(current, ".git")); err == nil && info.IsDir() {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func isWithin(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))))
}
