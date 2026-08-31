# RFC: Semantic Mutation Lab v1

Status: experimental, independent, and fail-closed.

## Scope

The lab is an evaluator conformance harness. It does not mutate source text,
rewrite Go syntax, run a product build, or modify a product repository. A
`.gooo` source declaration is compiled into a semantic IR. Each IR operator
then creates one caller-owned JSON artifact containing a baseline state and a
mutated state. The evaluator reads the artifact and emits a per-mutant state.

The only generated files are written beneath an output directory supplied by
the caller. The CLI refuses a repository descendant and refuses a non-empty
output directory, before creating anything.

## Fixed denominator

Version v1 contains exactly twelve cells and twelve mutation families. The
contract is fixed by ordinal, ID, changed edge, operator, and expected
detector. A source declaration that adds, removes, reorders, or rebinds a cell
or family is rejected before generation. This prevents denominator deletion
from looking like progress.

## Resolution

The evaluator applies `REFUTED > UNKNOWN > CLOSED`. It first collects semantic
contradictions, then considers evidence insufficiency. This means an artifact
with a digest mismatch remains `INVALID_REFUTED` even if another field would
otherwise be UNKNOWN. A valid mutation that cannot be decided because an exact
freshness digest is missing is `SURVIVED_UNKNOWN` and carries the complete
UNKNOWN tuple: `stage`, `step`, `reason`, `unknown_class`, `next_operation`,
and `blocked_by`.

`KILLED` means the evaluator detected and rejected the semantic regression.
`SURVIVED_UNKNOWN` means the mutant was attempted but the evaluator could not
decide it from available evidence. `INVALID_REFUTED` means the artifact itself
failed trust or shape validation.

## Reporting

The report has exact integer counts for generated, attempted, killed, unknown,
and refuted mutants. Each row carries the mutant ID, changed semantic edge,
expected detector, observed detector, artifact digest, state, and claim. No
aggregate mutation percentage or score is part of the contract. Improvement is
UNKNOWN unless an exact before/after pair has matching scenario, input,
contract, and toolchain digests.

The authority boundary is explicit and non-escalating: repository writes,
local test executions, and cross-project required gates are all zero.
