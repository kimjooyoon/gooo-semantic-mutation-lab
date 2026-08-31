# Semantic Mutation Lab v1

## Purpose

The lab answers one narrow question: does an evaluator detect semantic
regressions when a mutation is applied to semantic IR rather than to source
A `.gooo` declaration is parsed into an IR containing the declared
semantic edges and exactly twelve mutation operators. Generation receives only
that IR.

## Artifact boundary

Generation writes `ir.json`, one `mutants/<mutant-id>.json` per fixed cell, and
`generation-receipt.json` beneath a caller-owned output path. It never
writes the repository, source tree, or a Go file. Each artifact contains the
base source digest, base IR digest, exact changed semantic edge, before/after
values, mutated IR, and a digest over the artifact contents.

## Fixed denominator

The denominator is the 12-cell set in
`contracts/denominator-v1.json`. It is not inferred from
how many artifacts happen to exist. Validation rejects a missing, duplicate, or
unbound cell. The twelve families are `SELF_ATTESTATION`, `FIXED_POINT_ABUSE`,
`FAIL_OPEN_UNKNOWN`, `DENOMINATOR_DELETION`, `AUTHORITY_ESCALATION`,
`STALE_EVIDENCE`, `DIGEST_MISMATCH`, `DEPENDENCY_EDGE_REMOVAL`,
`SOURCE_IR_DRIFT`, `ARTIFACT_SUBSTITUTION`, `UNKNOWN_FRONTIER_LOSS`, and
`REFUTATION_PRECEDENCE_INVERSION`.

## Classification

For each cell, the evaluator first validates the artifact digest, source/IR
binding, exact-one-edge mutation, and semantic IR invariants. A contradiction is
`INVALID_REFUTED`. Only an artifact that passes invariant checks can become
`KILLED` or `SURVIVED_UNKNOWN`. The latter is valid only when all six fields
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`
are present and non-empty. This gives refutation strict precedence over
uncertainty.

## Reporting

`report.json` and `summary.md` contain exact generated, attempted, killed,
unknown, and refuted counts and the per-cell evidence listed above. The
protocol keeps these counts explicit instead of collapsing them into one
aggregate. Improvement is
`UNKNOWN` unless an exact before/after improvement pair is supplied; a mutant's
own before/after values do not close that separate claim.

## Authority and verification

The normal receipt records zero repository writes, zero local test executions,
and zero cross-project required gates. The root README inventory is explicitly
excluded. CI uses Go 1.27, executes format/vet/test/build and conformance only
on the runner, runs the source-to-report path twice, and compares the resulting
JSON reports for deterministic replay.
