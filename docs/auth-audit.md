# P0.4 — WS-UsernameToken Auth Audit (Go vs C)

Read-only audit comparing `internal/auth/auth.go` against the C reference's
WS-Security UsernameToken handling in `onvif_simple_server.c:367-455` and
`utils.c` (hashSHA1, b64_encode).

Audited 2026-05-12.

## Algorithmic Parity

Digest formula is identical:
`Base64( SHA1( Base64Decode(Nonce) + Created + Password ) )`

- C: `onvif_simple_server.c:412-418` (memcpy + hashSHA1 + b64_encode)
- Go: `internal/auth/auth.go:61-65` (`sha1.New().Write(nonce, Created, Password)` + base64)

**Verdict**: ✅ parity. No fix needed.

## Functional Divergences

| # | Behaviour | C reference | Go current | Severity | Impact on fixtures |
|---|---|---|---|---|---|
| 1 | Plain-text password (no Nonce/Created) | **Rejected** (`auth_error=3 or 4`) | **Accepted** (`auth.go:42-48`) | Medium | None — Go is more permissive, C-captured fixtures unaffected because they will always include Nonce/Created |
| 2 | Auth-exempt device ops (`GetSystemDateAndTime`, `GetUsers`, `GetCapabilities`, `GetServices`, `GetServiceCapabilities`) | **Exempt** (`onvif_simple_server.c:442-448`: `auth_error=0` reset) | **Always required** (`onvif_server.go:368-385` checks every request when cfg.User set) | **High** | **Fixtures will diverge**: Go returns 401 for these ops, C returns the data |
| 3 | Timestamp skew tolerance | **None** (`Created` parsed but never compared to current time) | 300s max age (`NonceMaxAgeSeconds`, `auth.go:142`) | Low | None — fixture capture uses fresh tokens |
| 4 | Nonce cache (replay protection) | **None** (no dedup) | **None** (matches C) | — | None |
| 5 | SOAP header XML parsing | `ezxml` library (proper XML parser) | `strings.Index("<Username>", ...)` (`auth.go:79-122`) | Medium | Potential — Go parser rejects namespaced tags `<wsse:Username>` whereas C parser handles them |
| 6 | Per-failure error codes | 6 distinct `auth_error` values (1, 2, 3, 4, 10, 11) | Single boolean | Low | None |
| 7 | Empty config (`user=` unset) | Auth disabled (`security.enable=0`) | Auth disabled (`auth.go:31-33`) | — | None — both match |

## Gaps That Block Phase 0-prereq

### Blocker #2 — Device-service auth exemption

The C server unconditionally accepts unauthenticated requests for five ops
on `device_service`. These are the first five ops most ONVIF clients call
during discovery, and they appear in our 33 currently-wired Go ops list.

If the canonical fixture config (`test/fixtures/config/server.conf`) is
captured with `user=` **commented out** (current default), this divergence
is masked — both servers skip auth entirely. P0.2 fixture capture will run
in this mode, so blocker #2 is **deferred** until a later phase enables
authenticated fixtures.

**Recommendation**: capture P0.2 fixtures with the current unauthenticated
config; defer the device-exemption fix to a small follow-up PR before any
phase that flips `user=` on. File issue with reference to this audit.

### Non-blocker #5 — XML parsing fragility

`strings.Index("<Username>", ...)` will miss `<wsse:Username>` (the
canonical ONVIF form). This is latent but doesn't affect golden-diff
capture because the C server is producing **responses**, not consuming
authenticated requests, during P0.2. Fix in Phase 0a when the SOAP fault
library lands and we switch to a real XML parser pass.

## Fixes Deferred (not done in P0.4)

- Plain-text password path removal (divergence #1): trivial 6-line delete;
  defer until 0e doc-baseline PR.
- Device-exemption table (divergence #2): add to 0a fault-library PR.
- XML parser swap (divergence #5): part of 0a.
- Distinct error codes (divergence #6): low value, defer to 0a fault
  library (each error code maps to a different `ter:*` fault subcode).

## P0.2 Capture Posture

Audit confirms the canonical fixture config (`user=` commented out) is
the **correct** capture mode for P0.2: it neutralizes divergences #1, #2,
#3 and produces fixtures that exercise the SOAP body machinery without
auth complications. Authenticated-mode fixtures will be captured in a
follow-up phase after the 0a fault library lands.

**Conclusion**: P0.4 surfaces gaps but no merge-blocker for P0.2.
Phase 0-prereq sizing stays at **M**, not L.
