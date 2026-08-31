package lab

const (
	IRSchema       = "gooo/semantic-mutation-lab/ir/v1"
	ArtifactSchema = "gooo/semantic-mutation-lab/mutant-artifact/v1"
	ReportSchema   = "gooo/semantic-mutation-lab/report/v1"
	ReceiptSchema  = "gooo/semantic-mutation-lab/generation-receipt/v1"
	States         = 3
	FixedCells     = 12
)

type Authority struct {
	RepositoryWrites            int    `json:"repository_writes"`
	LocalTestExecutions         int    `json:"local_test_executions"`
	CrossProjectRequiredGates   int    `json:"cross_project_required_gates"`
	ProductGenerationAuthorized bool   `json:"product_generation_authorized"`
	RootReadmePolicy            string `json:"root_readme_policy"`
}

type Claim struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	BeforeDigest  string   `json:"before_digest"`
	AfterDigest   string   `json:"after_digest"`
}

func (c Claim) HasUnknownTuple() bool {
	return c.State == "UNKNOWN" && c.Stage != "" && c.Step != "" && c.Reason != "" &&
		c.UnknownClass != "" && c.NextOperation != "" && c.BlockedBy != nil
}

type Evidence struct {
	SubjectDigest   string `json:"subject_digest"`
	ContractDigest  string `json:"contract_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	CommandDigest   string `json:"command_digest"`
}

type CellDecl struct {
	Ordinal      int      `json:"ordinal"`
	ID           string   `json:"id"`
	SemanticEdge string   `json:"semantic_edge"`
	DependsOn    []string `json:"depends_on"`
}

type MutationDecl struct {
	Ordinal             int    `json:"ordinal"`
	ID                  string `json:"id"`
	Family              string `json:"family"`
	Cell                string `json:"cell"`
	ChangedSemanticEdge string `json:"changed_semantic_edge"`
	Operator            string `json:"operator"`
	ExpectedDetector    string `json:"expected_detector"`
}

type SourceDecl struct {
	Schema        string         `json:"schema"`
	Version       string         `json:"version"`
	DenominatorID string         `json:"denominator_id"`
	CellCount     int            `json:"cell_count"`
	Authority     Authority      `json:"authority"`
	Precedence    []string       `json:"precedence"`
	UnknownFields []string       `json:"unknown_fields"`
	Cells         []CellDecl     `json:"cells"`
	Mutations     []MutationDecl `json:"mutations"`
	SourceDigest  string         `json:"source_digest"`
}

type Contract struct {
	Schema    string         `json:"schema"`
	ID        string         `json:"id"`
	Version   string         `json:"version"`
	CellCount int            `json:"cell_count"`
	Fixed     bool           `json:"fixed"`
	Cells     []CellDecl     `json:"cells"`
	Families  []MutationDecl `json:"families"`
}

type SemanticState struct {
	DenominatorCellIDs []string            `json:"denominator_cell_ids"`
	Dependencies       map[string][]string `json:"dependencies"`
	SourceIRBindings   map[string]string   `json:"source_ir_bindings"`
	ArtifactBindings   map[string]string   `json:"artifact_bindings"`
	Decisions          map[string]string   `json:"decisions"`
	Claims             map[string]Claim    `json:"claims"`
	Evidence           map[string]Evidence `json:"evidence"`
	Authority          Authority           `json:"authority"`
	Attestation        string              `json:"attestation"`
	Precedence         []string            `json:"precedence"`
}

type IR struct {
	Schema         string         `json:"schema"`
	Version        string         `json:"version"`
	SourceDigest   string         `json:"source_digest"`
	ContractDigest string         `json:"contract_digest"`
	DenominatorID  string         `json:"denominator_id"`
	CellCount      int            `json:"cell_count"`
	Authority      Authority      `json:"authority"`
	Precedence     []string       `json:"precedence"`
	UnknownFields  []string       `json:"unknown_fields"`
	Cells          []CellDecl     `json:"cells"`
	Mutations      []MutationDecl `json:"mutations"`
	IRDigest       string         `json:"ir_digest,omitempty"`
}

type MutantArtifact struct {
	Schema              string        `json:"schema"`
	MutantID            string        `json:"mutant_id"`
	Family              string        `json:"family"`
	Cell                string        `json:"cell"`
	SourceDigest        string        `json:"source_digest"`
	ContractDigest      string        `json:"contract_digest"`
	IRDigest            string        `json:"ir_digest"`
	ChangedSemanticEdge string        `json:"changed_semantic_edge"`
	ExpectedDetector    string        `json:"expected_detector"`
	Baseline            SemanticState `json:"baseline"`
	Mutated             SemanticState `json:"mutated"`
	ArtifactDigest      string        `json:"artifact_digest,omitempty"`
}

type MutantResult struct {
	MutantID            string `json:"mutant_id"`
	Family              string `json:"family"`
	Cell                string `json:"cell"`
	State               string `json:"state"`
	ChangedSemanticEdge string `json:"changed_semantic_edge"`
	ExpectedDetector    string `json:"expected_detector"`
	ObservedDetector    string `json:"observed_detector"`
	ArtifactDigest      string `json:"artifact_digest"`
	Claim               Claim  `json:"claim"`
}

type Report struct {
	Schema                      string         `json:"schema"`
	Decision                    string         `json:"decision"`
	FixedConformanceDenominator int            `json:"fixed_conformance_denominator"`
	SourceDigest                string         `json:"source_digest"`
	ContractDigest              string         `json:"contract_digest"`
	IRDigest                    string         `json:"ir_digest"`
	Precedence                  []string       `json:"precedence"`
	Authority                   Authority      `json:"authority"`
	Summary                     Summary        `json:"summary"`
	Mutants                     []MutantResult `json:"mutants"`
	Improvement                 Claim          `json:"improvement"`
}

type Summary struct {
	Generated int `json:"generated"`
	Attempted int `json:"attempted"`
	Killed    int `json:"killed"`
	Unknown   int `json:"unknown"`
	Refuted   int `json:"refuted"`
}

type GenerationReceipt struct {
	Schema                    string `json:"schema"`
	SourceToIR                string `json:"source_to_ir"`
	IRToMutants               string `json:"ir_to_mutants"`
	Generated                 int    `json:"generated"`
	Attempted                 int    `json:"attempted"`
	CallerOwnedTempOutput     bool   `json:"caller_owned_temp_output"`
	RepositoryWrites          int    `json:"repository_writes"`
	LocalTestExecutions       int    `json:"local_test_executions"`
	CrossProjectRequiredGates int    `json:"cross_project_required_gates"`
}
