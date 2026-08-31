#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: conformance.sh PATH_TO_MUTATIONLAB" >&2
  exit 64
fi

bin=$(realpath "$1")
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tmp=$(mktemp -d "${RUNNER_TEMP:-/tmp}/gooo-semantic-mutation-lab.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
source="$root/examples/semantic-mutation-lab-v1/lab.gooo"
contract="$root/contracts/denominator-v1.json"
run="$tmp/run"

before=$(git -C "$root" status --porcelain=v1)
"$bin" run --source "$source" --contract "$contract" --out "$run" >/dev/null
report="$run/mutation-report.json"
jq -e '
  .schema == "gooo/semantic-mutation-lab/report/v1" and
  .fixed_conformance_denominator == 12 and
  .precedence == ["REFUTED", "UNKNOWN", "CLOSED"] and
  .summary == {generated:12,attempted:12,killed:10,unknown:1,refuted:1} and
  ([.mutants[] | select(.state == "KILLED")]|length) == 10 and
  ([.mutants[] | select(.state == "SURVIVED_UNKNOWN")]|length) == 1 and
  ([.mutants[] | select(.state == "INVALID_REFUTED")]|length) == 1 and
  ([.mutants[] | select((.mutant_id|length)>0 and (.changed_semantic_edge|length)>0 and (.expected_detector|length)>0 and (.observed_detector|length)>0 and (.artifact_digest|test("^sha256:"))) ]|length) == 12 and
  ([.mutants[] | select(.state == "SURVIVED_UNKNOWN" and .claim.state == "UNKNOWN" and (.claim.stage|length)>0 and (.claim.step|length)>0 and (.claim.reason|length)>0 and (.claim.unknown_class|length)>0 and (.claim.next_operation|length)>0 and (.claim.blocked_by|type)=="array")]|length) == 1 and
  (has("percentage")|not) and (has("score")|not)
' "$report" >/dev/null

test "$(find "$run/mutants" -maxdepth 1 -type f -name '*.json' | wc -l | tr -d ' ')" -eq 12
test "$(find "$run" -type f | wc -l | tr -d ' ')" -eq 15
forbidden="$root/.gooo-semantic-mutation-lab-forbidden-output"
set +e
"$bin" run --source "$source" --contract "$contract" --out "$forbidden" >/dev/null 2>&1
status=$?
set -e
test "$status" -ne 0
test ! -e "$forbidden"
after=$(git -C "$root" status --porcelain=v1)
test "$before" = "$after"
echo "semantic mutation conformance: PASS (generated=12 attempted=12 killed=10 unknown=1 refuted=1)"
