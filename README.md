# Gooo Semantic Mutation Lab

Gooo Semantic Mutation Lab is a small, independent conformance laboratory for
semantic evaluators. It answers one bounded question: does an evaluator reject
controlled semantic regressions described by a `.gooo` source declaration?

The lab is deliberately not a text mutator and never edits Go source. The
`.gooo` file declares a fixed 12-cell denominator and exactly 12 semantic IR
mutation families. The CLI performs this chain:

```text
.gooo source → semantic IR → caller-owned mutant artifacts → evaluator → report
```

`gooo-semantic-mutation-lab run` writes `ir.json`, a generation receipt, twelve
JSON mutant artifacts, and `mutation-report.json` only under the caller-provided
output directory. The directory must be outside the repository and empty (or
absent). This makes repository writes observable and keeps generated artifacts
disposable.

## Fixed contract

The v1 contract has exactly twelve mutation families and exactly twelve
conformance cells, in a one-to-one ordinal mapping:

1. `SELF_ATTESTATION`
2. `FIXED_POINT_ABUSE`
3. `FAIL_OPEN_UNKNOWN`
4. `DENOMINATOR_DELETION`
5. `AUTHORITY_ESCALATION`
6. `STALE_EVIDENCE`
7. `DIGEST_MISMATCH`
8. `DEPENDENCY_EDGE_REMOVAL`
9. `SOURCE_IR_DRIFT`
10. `ARTIFACT_SUBSTITUTION`
11. `UNKNOWN_FRONTIER_LOSS`
12. `REFUTATION_PRECEDENCE_INVERSION`

The evaluator records each mutant as `KILLED`, `SURVIVED_UNKNOWN`, or
`INVALID_REFUTED`. A semantic contradiction is resolved before any uncertainty;
therefore `REFUTED` always wins over `UNKNOWN`. An UNKNOWN claim must preserve
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.

The report contains exact `generated`, `attempted`, `killed`, `unknown`, and
`refuted` counts plus per-mutant identity, changed semantic edge, expected and
observed detector, and artifact digest. It intentionally keeps the counts
explicit instead of collapsing them into one aggregate. Exact before/after evidence is
required for any improvement claim; without it, improvement is UNKNOWN.

The authority receipt is fixed at `repository_writes=0`,
`local_test_executions=0`, and `cross_project_required_gates=0`. The lab does
not build, test, patch, or release another project.

## Use

```sh
go run ./cmd/gooo-semantic-mutation-lab run \
  --source examples/semantic-mutation-lab-v1/lab.gooo \
  --contract contracts/denominator-v1.json \
  --out /tmp/gooo-semantic-mutation-lab-run
```

The repository's GitHub Actions workflow is the validation authority. It uses
Go 1.27, runs formatting/build/test/vet/conformance in CI, checks that the
repository remains unchanged, and writes its generated output under the
runner's temporary directory. The root `README.md` is the explicit inventory
exception when CI counts repository files.
