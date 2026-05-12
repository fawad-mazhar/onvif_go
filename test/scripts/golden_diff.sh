#!/usr/bin/env bash
# golden_diff.sh - parity golden-diff CI gate (Phase 0a).
#
# Runs the in-process Go server against the captured C-reference fixtures,
# scrubs volatile fields, canonicalizes both sides via c14n, and enforces
# the regression baseline in test/fixtures/.baseline.
#
# Exit code:
#   0 if every op in .baseline still passes (gain allowed, loss forbidden)
#   non-zero if any previously-passing op regressed, or the harness errored.
#
# To refresh the baseline after a legitimate improvement:
#   GOLDEN_UPDATE_BASELINE=1 test/scripts/golden_diff.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

if [[ ! -d test/fixtures ]]; then
    echo "ERROR: test/fixtures/ missing (run test/scripts/capture_fixtures.sh)" >&2
    exit 2
fi

# c14n spike is a useful canary for the canonicalizer itself.
echo "--- c14n spike ---"
go run ./test/scripts/c14n_spike

echo
echo "--- golden diff ---"
go test ./test/ -run GoldenDiff -count=1 -v
