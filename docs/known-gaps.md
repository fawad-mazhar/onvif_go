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

### G-003 — Brittle SOAP header parsing

**Summary**: `auth.ParseSOAPHeader` uses `strings.Index("<Username>", ...)`
which doesn't match the canonical `<wsse:Username>` form that real
ONVIF clients send. Latent bug; doesn't affect response capture.

**Fix window**: Phase 0a (this PR's follow-up) or Phase 1 — switch
to a real XML parser once the fault library is in use.

---

### G-004 — `parseSOAPAction` doesn't match `<s:Body>`

**Summary**: `parseSOAPAction` in `onvif_server.go:403-405` only
recognizes `<soap:Body>` and `<SOAP-ENV:Body>`. The fixture SOAP
requests use `<s:Body>` (a legitimate alias), so every fixture POST
currently returns a fault. Visible in the `TestGoldenDiff` log.

**Impact**: 0/21 golden diffs currently pass. After fixing this gap,
the baseline should be regenerable with more passes.

**Fix window**: Phase 0a follow-up or Phase 1 — rewrite with a real
XML namespace-aware parser or a broader prefix alternation.

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

### G-006 — `adv_fault_if_unknown` not implemented

**Summary**: C supports an `adv_fault_if_unknown` config flag that
switches unsupported-op handling between "empty 200" and "fault 500".
Go always returns a fault. See plan §0b.

**Fix window**: Phase 0b.

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

## Procedure for adding a new gap

1. Assign next `G-nnn` id.
2. Document: summary, locations (absolute paths), masking, fix window.
3. Link from the relevant phase in `~/.windsurf/plans/onvif-go-c-parity-v3-46c14e.md` if the gap blocks or shapes that phase.
