#!/usr/bin/env bash
# golden_diff.sh - P0.5 stub for the parity golden-diff CI gate.
#
# Full implementation lands in Phase 0a, which adds:
#   - the volatile-field scrubber (see docs/testing-strategy.md)
#   - the Go-server harness that runs against test/fixtures/config/server.conf
#   - per-fixture pass/fail reporting
#
# Until then this stub only:
#   - confirms the fixture tree is present
#   - runs the c14n spike as a minimal canaries

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

if [[ ! -d test/fixtures ]]; then
    echo "ERROR: test/fixtures/ missing (run test/scripts/capture_fixtures.sh)" >&2
    exit 2
fi

fixture_count=$(find test/fixtures -name '*.response.xml' | wc -l | tr -d ' ')
echo "Fixtures present: $fixture_count"

if [[ "$fixture_count" -lt 21 ]]; then
    echo "ERROR: expected >= 21 fixtures, found $fixture_count" >&2
    exit 1
fi

echo "Running c14n spike..."
go run ./test/scripts/c14n_spike

echo
echo "STUB: full Go-vs-C golden diff will run here after 0a."
echo "PASS (stub)."
