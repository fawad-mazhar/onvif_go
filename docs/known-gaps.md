# Known Parity Gaps (tracked)

Running list of Go-vs-C divergences that are **deferred but known**.
Each item names the fix window (phase) and the masking assumption that
makes the deferral safe.

## Opened 2026-05-12 (Phase 0a)

### G-001 — Device-service auth exemption

**Summary**: C exempts five `device_service` ops from WS-UsernameToken
validation (`GetSystemDateAndTime`, `GetUsers`, `GetCapabilities`,
`GetServices`, `GetServiceCapabilities`). Go always requires auth when
`user=` is configured.

**Locations**:
- C: `@/Users/fawadmazhar/github/codes/oma/playground/onvif_simple_server/onvif_simple_server.c:442-448`
- Go: `@/Users/fawadmazhar/github/codes/oma/public/onvif_go/internal/server/onvif_server.go:368-385`

**Masking**: `test/fixtures/config/server.conf` has `user=` commented out,
so both servers skip auth entirely during fixture capture. Fault-mode
fixtures would expose this; we don't currently capture those.

**Fix window**: Phase 1 (device-service expansion) — add an exemption
allow-list around the auth middleware. Needs a paired fixture-capture
pass in authenticated mode.

**Reference**: `@/Users/fawadmazhar/github/codes/oma/public/onvif_go/docs/auth-audit.md` §Blocker-#2.

---

### G-002 — Plain-text password accepted

**Summary**: Go's `ValidateUsernameToken` accepts a plain-text password
with no Nonce/Created fields (`@/Users/fawadmazhar/github/codes/oma/public/onvif_go/internal/auth/auth.go:42-48`).
C rejects this path (`auth_error = 3 or 4`).

**Masking**: Go is more permissive; fixtures captured with fresh
digest tokens don't exercise this path.

**Fix window**: Phase 0e (doc baseline PR) — delete the 6-line branch.

---

### ~~G-003 — Brittle SOAP header parsing~~ (FIXED)

**Resolved 2026-05-12** in commit following Phase 0a. `auth.ParseSOAPHeader`
now delegates to `internal/xml.ExtractUsernameToken`, which uses
`encoding/xml` namespace-aware decoding. Tolerates any prefix
(`wsse:`, `u:`, etc.).

---

### ~~G-004 — `parseSOAPAction` doesn't match `<s:Body>`~~ (FIXED)

**Resolved 2026-05-12** in commit following Phase 0a. `parseSOAPAction`
now delegates to `internal/xml.ExtractBodyAction`, which uses
`encoding/xml` namespace-aware decoding. Accepts every SOAP 1.2
envelope prefix (`s:`, `soap:`, `SOAP-ENV:`, `env:`).

**Aftermath**: handler routing now works for every captured fixture,
but Go handler outputs still differ from C templates byte-for-byte.
The gap shifted from parsing to handler output; tracked as G-008.

---

### G-005 — Fault rendering doesn't know device/service address

**Summary**: `sendSOAPError` renders `Fault.xml` with empty
`%ADDRESS%` / `%SERVICE%` placeholders because the handler layer
doesn't thread the request's `Host:` header through. The scrubber
normalizes host URLs on both sides so diffs still work, but the
Go fault body is technically empty-stringed where C fills in
`http://host:port/onvif/<svc>`.

**Fix window**: Phase 1 — plumb `*http.Request` down to `sendSOAPError`.

---

### G-006 — `adv_fault_if_unknown` / `adv_fault_if_set` not implemented ✅ fixed

**Summary**: C supports `adv_fault_if_unknown` (switches unsupported-op
handling between "empty 200" and "fault 500") and `adv_fault_if_set`
(makes five media `Set*` ops fault unconditionally). Go always returned
a fault regardless of config.

**Fixed in**: Phase 0b (`97cd2ec`) — `sendUnsupportedResponse` in
`internal/server/onvif_server.go` now branches on `cfg.AdvFaultIfUnknown`
and passes the canonical ONVIF ns prefix (`tds`/`trt`/`tptz`/`tev`/`tmd`)
to `xmlfault.RenderEmpty`, matching C's `send_empty_response(ns, method)`.

---

### G-007 — C datetime drift in recaptured fixtures

**Summary**: `GetSystemDateAndTime.response.xml` bakes in the
capture-time clock values. Recapturing in later phases will produce
a noisy diff in repo history. Scrubber masks it for diff, but the
file keeps changing.

**Fix window**: Phase 0a follow-up — either add `libfaketime` to the
C Dockerfile and pin the clock at `2026-01-01T00:00:00Z`, or
post-process captured fixtures to blank out the volatile fields.

---

### G-009 — Latent string-index parser in `utils.ExtractSOAPElement`

**Summary**: `internal/utils/utils.go:ExtractSOAPElement` uses the same
prefix-enumeration pattern as the now-fixed G-003/G-004:

```go
for _, ns := range []string{"trt:", "tt:", "ter:", "tns1:"} {
    openTag := "<" + ns + elementName + ">"
```

Two callers in `pkg/services/media/service.go` extract `<ProfileToken>`
from `GetStreamUri` and `GetSnapshotUri` requests. Works for the
captured fixtures (which use `<trt:ProfileToken>`), breaks silently for
any other prefix.

**Masking**: Current test fixtures use the `trt:` prefix throughout;
the harness doesn't test alternate prefixes for media ops.

**Fix window**: Phase 2 (Media service) — the media handlers will be
substantially rewritten in that phase. Replace with a call to
`internal/xml.ExtractElementValue` (backed by `encoding/xml`, same
pattern as `ExtractBodyAction`). Don't gate Phase 0b on this.

---

### G-008 — Go handler templates diverge from C reference (umbrella)

**Summary**: After G-003/G-004 fixes, every captured request now routes
through the appropriate Go handler. None of the 21 captured fixtures
match the C reference byte-for-byte yet — Go's templates under
`service_files/<svc>/` predate the parity effort and were authored
independently of the C source's `<svc>_service_files/`.

**Root cause — corrected diagnosis (Phase 0b follow-up)**: The initial
diagnosis attributed the `GetProfiles` 329 vs 8304 byte gap entirely to
template divergence. That was partially wrong. The Go config parser only
understood dotted Go-format keys (`profile.0.name=`, `scope.0=`, etc.)
and silently dropped every flat C-format key in the canonical fixture
(`name=`, `scope=`, `ptz=`, `idle_state=`, `topic=`, …). Calling
`LoadConfig("test/fixtures/config/server.conf")` returned `cfg.Profiles`
of length 0. Fixed in the Phase 0b follow-up commit by adding
`parseFlatConfig` to `internal/config/config.go` and the canary test
`TestLoadConfig_CanonicalFixture`. After that fix profiles are loaded,
so the byte gap on `GetProfiles` will narrow substantially when templates
are aligned.

**Examples** (from `TestGoldenDiff` summary, before config fix):
- `device_service/GetUsers`        Go 252 vs C 223 bytes canonical
- `device_service/GetCapabilities` Go 3081 vs C 2523 bytes
- `media_service/GetProfiles`      Go 329 vs C 8304 bytes (zero profiles
  loaded; both config parser gap AND template divergence)

**Fix window**: Distributed across phases 1-5 of the parity plan.
Each phase brings one or two services to byte-identical parity by
either (a) replacing `service_files/<svc>/<op>.xml` with the C
verbatim template, or (b) introducing the C-style header/middle/footer
composer for list ops.

**Regression-gate behaviour**: `test/fixtures/.baseline` records the
set of currently-passing ops. As each phase lands, the baseline grows;
any commit that drops an op from passing fails CI. The baseline is
empty today (0/21).

---

## Procedure for adding a new gap

1. Assign next `G-nnn` id.
2. Document: summary, locations (absolute paths), masking, fix window.
3. Link from the relevant phase in `~/.windsurf/plans/onvif-go-c-parity-v3-46c14e.md` if the gap blocks or shapes that phase.
