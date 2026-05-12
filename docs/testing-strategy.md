# Testing Strategy — Parity Golden-Diff

Captures the decisions and workflow for validating `onvif_go` against
the C reference server (`onvif_simple_server`).

## Parity Oracle

- **C reference server** built from
  `@/Users/fawadmazhar/github/codes/oma/playground/onvif_simple_server`
  (image `onvif-c-reference:latest` via `test/fixtures/c-reference/Dockerfile`).
- **Canonical fixture config** at `test/fixtures/config/server.conf` is
  loaded by both the C server (P0.1) and the Go server under CI. Never
  depend on upstream `onvif_simple_server.conf.example` directly.
- **Captured fixtures** live at
  `test/fixtures/<service>/<Op>.{request,response}.xml`.

## P0.3 — c14n Library Decision

**Choice**: `github.com/ucarion/c14n` v0.1.0 (Exclusive XML Canonicalization).

**Spike**: `test/scripts/c14n_spike` (runnable via `go run ./test/scripts/c14n_spike`).

Spike results against the 21 fixtures captured in P0.2:

| Check | Result |
|---|---|
| Parses every fixture without error | 21/21 ✅ |
| Canonical form is idempotent (canon(canon(x)) == canon(x)) | 21/21 ✅ |
| Canonical form is prefix-agnostic | 0/21 ❌ (by design) |

**Interpretation**: Exclusive c14n preserves namespace prefixes
literally — this is XML-Signature's required behaviour. For our
strict-parity diffs this is **not** a problem because the Go
templates will be copied verbatim from the C reference, so prefixes
will already match. Exclusive c14n still normalizes the differences
that *do* trip naive diffs:

- whitespace and line endings
- self-closing vs open/close tag pairs
- attribute ordering
- `xmlns` declaration positioning

**Fallback (unused)**: `github.com/beevik/etree` + custom
canonicalizer, only if Exclusive c14n is found inadequate during
Phase 0a handler work.

## Diff Workflow

1. Capture C-reference response: `test/scripts/capture_fixtures.sh`
   writes `test/fixtures/<svc>/<Op>.response.xml`.
2. Run Go server against the same fixture config and same SOAP
   request, capture its response.
3. Canonicalize both via `ucarion/c14n`.
4. Apply a volatility-scrubbing pass (see below) to both canonical
   forms.
5. Byte-compare.

### Volatile Fields to Scrub Before Diff

Known volatile fields in captured C responses (these must be
normalized or blanked before byte-comparison):

- `<tt:UTCDateTime>` / `<tt:DaylightSavings>` in `GetSystemDateAndTime`
- `wsa5:MessageID` and `wsa5:Address` in SOAP fault headers (echo of
  `Host:` and generated UUIDs)
- `<tt:HwAddress>` in `GetNetworkInterfaces` (depends on container
  netif)

The 0a PR introduces the scrubber. Until then the fixtures are
committed as-is for reference only.

## CI Gate (wired in P0.5)

`test/scripts/golden_diff.sh` (P0.5 stub) runs the full diff in CI
and fails the build on any regression. Gate activates with 0a.
