# Known Parity Gaps (tracked)

Running list of Go-vs-C divergences that are **deferred but known**.
Each item names the fix window (phase) and the masking assumption that
makes the deferral safe.

## Opened 2026-05-12 (Phase 0a)

### ~~G-001 — Device-service auth exemption~~ (FIXED)

**Resolved 2026-05-20** (Phase 1). `createONVIFHandler` in
`internal/server/onvif_server.go` now has an `authExempt` guard that
skips WS-UsernameToken validation for the five C-exempt ops:
`GetSystemDateAndTime`, `GetUsers`, `GetCapabilities`, `GetServices`,
`GetServiceCapabilities`. Matches `onvif_simple_server.c:442-448`.

---

### ~~G-002 — Plain-text password accepted~~ (FIXED)

**Resolved 2026-05-21** (Phase 3).

Removed the 6-line plain-text branch from `ValidateUsernameToken`
(`internal/auth/auth.go`). Only digest auth (Nonce + Created + SHA1)
is now accepted, matching C's `auth_error = 3/4` rejection.

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

### ~~G-005 — Fault rendering doesn't know device/service address~~ (FIXED)

**Resolved 2026-05-21** (Phase 3 + Phase 3.1).

All fault-emitting paths now derive `DeviceAddress` and `ServiceAddress`
from `r.Host` + `r.URL.Path`, filling `Fault.xml`'s `%ADDRESS%` /
`%SERVICE%` placeholders correctly:

- Phase 3: `sendSOAPError(w, r, msg)` — `*http.Request` threaded through
  `handleServiceError`, `sendUnsupportedResponse`, `handleUnsupportedService`.
- Phase 3.1: three handler-level `WriteFault` sites that bypassed the
  central path:
  - `device.GetCapabilitiesHTTP(w, r, soapRequest)` — unknown-category
    and PTZ-disabled faults.
  - `media.sendMediaFault(w, r, ...)` — profile-not-found and URI faults
    (`GetProfileHTTP`, `getMediaUriHTTP`, `CreateProfileHTTP`,
    `DeleteProfileHTTP`).

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

### G-008 — Go handler templates diverge from C reference (umbrella)

**Summary**: After G-003/G-004 fixes, every captured request now routes
through the appropriate Go handler. None of the 21 captured fixtures
matched the C reference byte-for-byte yet — Go's templates predate the
parity effort.

**Phase 1 progress (2026-05-20)**:
- All 9 `device_service` ops now pass golden diff (baseline 10/21).
- Root changes: C verbatim templates in `service_files/device/`,
  `ProcessTemplate` now replicates C `cat()` line-trim+concat semantics,
  `ServiceContext` carries `Interface`/`EventsEnable`/audio/relay fields,
  `GetServiceCapabilities` route added, `DeleteProfile` removed.
- `GetCapabilities` and `GetServices` now select template variant based
  on PTZ / Media2 / IncludeCapability flags (matching C logic).
- `GetUsers` returns `send_empty_response` output via `WriteEmpty`.
- `GetSystemDateAndTime` uses real UTC clock with no zero-padding
  (`%d` semantics matching C `sprintf`).
- `getInterfaceIP` falls back to `127.0.0.1` for unknown interface
  names (handles `lo` on macOS where the name is `lo0`).

**Fix window**: Continuing across phases 2-5 for remaining services.
Each phase brings one or two services to byte-identical parity.

**Regression-gate**: `test/fixtures/.baseline` is now 10/21.
Any commit that drops a passing op fails CI.

---

### ~~G-009 — Latent string-index parser in `utils.ExtractSOAPElement`~~ (FIXED)

**Resolved 2026-05-21** (Phase 3).

`utils.ExtractSOAPElement` deleted from `internal/utils/utils.go`.
Both callers in `pkg/services/media/service.go` (`GetProfileHTTP`,
`getMediaUriHTTP`) now call `xml.ExtractElement` (namespace-agnostic,
`encoding/xml` backed). `sendMediaFault` also migrated from an inline
`fmt.Sprintf` fault body to `xmlfault.WriteFault` for consistency.

---

## Opened 2026-05-13 (Phase 0d follow-up)

### ~~G-010 — `internal/exec`: deferred shell-exec security scope~~ (FIXED)

**Resolved Phase 4.**

**Block-merge condition met**: `validateSOAPArg` in `pkg/services/ptz/service.go`
rejects `;`, `$(`, `` ` ``, `&&`, `||`, `>>`, `|`, and `--`-prefixed tokens
at the PTZ handler layer before any string reaches `exec.RunFmt`. Applied to
`SetPresetHTTP` (the sole path where a SOAP-derived string — `PresetName` —
is passed as a `%s` argument to a subprocess). All integer/float SOAP args
(preset numbers, coordinates) are parsed to Go numeric types before use,
eliminating string injection risk on those paths.

Unit tests in `pkg/services/ptz/service_test.go::TestValidateSOAPArg` cover
all 8 banned sequences plus the `--prefix` rule (10 sub-tests, all green).

Remaining deferred items (allow-list, output size cap) are tracked separately
as low-priority hardening items outside Phase 4 scope.

---

### ~~G-011 — WSD `parseSOAPAction` / `parseMessageID` fragility and offset bug~~ (FIXED)

**Resolved 2026-05-21** (Phase 2).

1. **`parseSOAPAction` / `parseMessageID`**: replaced string-index parsers
   with one-liner delegates to `xml.ExtractElement` (backed by
   `encoding/xml`). Accepts any namespace prefix; the off-by-N bug (advancing
   by `len("<wsa:Action>")` after matching `<a:Action>`) is gone.

2. **Duplicate signal handler**: `WSDServer.setupSignalHandler` removed.
   `StartWSDServer` now accepts a `context.Context` owned by the caller.
   `main()` creates a single handler via `signal.NotifyContext` and passes
   the resulting ctx; cancellation propagates to `listenForMessages` which
   sends Bye and returns.

3. **Bye-on-shutdown**: Bye message is now sent in the `ctx.Done()` branch
   of `listenForMessages`, not in the deleted signal goroutine.

**Tests added** (`internal/server/wsd_server_test.go`):
- `TestParseSOAPAction_WSAPrefix` / `_APrefix` / `_Empty`
- `TestParseMessageID_WSAPrefix` / `_APrefix` / `_Empty`
  The `a:` prefix tests document the exact symptom the old parser produced
  (`"tp://..."` instead of `"http://..."`).

---

## Opened 2026-05-20 (Phase 1 review)

### ~~G-012 — Missing `GetServices_with_capabilities_*.xml` templates~~ (FIXED)

**Summary**: `servicesTemplate(ptz, m2, true)` selected one of four
`GetServices_with_capabilities_*` file names that were never shipped.
A client sending `IncludeCapability=true` would silently hit the
`Empty.xml` fallback and receive malformed XML with a literal
`<%METHOD% />` body.

**Fixed in**: Phase 1 review — copied all four C reference templates
(`GetServices_with_capabilities_{ptz,no_ptz}_{media2,no_media2}.xml`)
to `service_files/device/`. `ProcessServiceTemplate` now also logs a
warning and injects `%METHOD%` on any future fallback so the defect
is visible without a fixture.

---

### ~~G-013 — `GetCapabilitiesHTTP` ignored `<Category>` element~~ (FIXED)

**Summary**: The Phase 1 implementation always returned the full
all-categories response regardless of what the client requested in
`<Category>`. A client requesting `Category=Device`, `Media`, `PTZ`,
or `Events` received the wrong response shape with a 200 OK. An
unknown category value should produce a SOAP fault; it instead
returned the full response too. The fixture sends `Category=All`
so the parity test passed by coincidence.

**C reference**: `device_service.c:444-596` — five `icategory` dispatch
paths, each with its own template file and different placeholder set.

**Fixed in**: Phase 1 review —
- Extracted `categoryCode(string) int` helper (testable, matches C's
  `strcasecmp` dispatch via `strings.ToLower`).
- `GetCapabilitiesHTTP` now dispatches on `icategory`:
  1 → `GetDeviceCapabilities.xml`, 2 → `GetMediaCapabilities.xml`,
  4 → `GetPTZCapabilities.xml` (or fault if `!PTZEnable`),
  8 → `GetEventsCapabilities.xml`, 15 → full `GetCapabilities_{ptz,no_ptz}.xml`.
- Unknown category → `SOAP-ENV:Receiver / ter:ActionNotSupported / ter:NoSuchService`.
- Copied all four Category-specific C reference templates to
  `service_files/device/`.

---

### G-014 — No graceful shutdown for HTTP and Notification servers

**Summary**: After the G-011 signal-handler refactor, `StartWSDServer` accepts
a `context.Context` and sends WS-Discovery Bye before exiting. The HTTP and
Notification servers (`StartHTTPServer`, `StartNotificationServer`) do not
receive a context and are killed mid-flight when `main()` exits after
`<-ctx.Done()`. In-flight SOAP requests can be truncated and the Notification
server connection state is abandoned.

**Locations**:
- `cmd/onvif-server/main.go:41-54` (HTTP and Notification goroutines)
- `internal/server/onvif_server.go` (`StartHTTPServer`)
- `internal/server/notification_server.go` (`StartNotificationServer`)

**Masking**: No integration tests exercise the shutdown path. The 1-second
`time.Sleep` that existed in the original main.go was the only protection;
it was removed in Phase 2.

**Fix window**: Phase 3 (HTTP server parity) or Phase 5 (Events).
- Replace `net/http.ListenAndServe` with `http.Server.Shutdown(ctx)` pattern
  in `StartHTTPServer`.
- Propagate the same ctx to `StartNotificationServer`.
- `main.go` waits for all goroutines via a `sync.WaitGroup` before exiting.

---

### G-015 — Events stateful-ops parity placeholders

**Summary**: Four hardcoded values in the Events service will diverge from
the C reference once fixtures are captured for Subscribe, Renew, and
CreatePullPointSubscription:

1. **`CreatePullPointSubscriptionHTTP`** — address uses `"http://localhost:%d/..."`.
   Should derive host from `r.Host` (same G-005 pattern applied to PTZ/device).
   `pkg/services/events/service.go:307`.

2. **`SubscribeHTTP`** — same hardcoded `localhost` address for `%REFERENCE%`
   and consumer reference.
   `pkg/services/events/service.go:~370`.

3. **`Renew` + `Subscribe` UUIDs** — `%MSG_UUID%` uses `fmt.Sprintf("uuid-%d",
   time.Now().UnixNano())` (not RFC 4122); `%REL_TO_UUID%` is literal
   `"uuid-relates-to"` instead of being parsed from the request's
   `wsa:MessageID` header.
   `pkg/services/events/service.go:419-420,~380`.

4. **`GetServiceCapabilitiesHTTP`** — `%EVENTS_BASESUBSCRIPTION%` and
   `%EVENTS_PULLPOINT%` hardcoded to `"true"` regardless of `cfg.EventsEnable`
   bitmask (`1` = base, `2` = pull-point, `3` = both).
   `pkg/services/events/service.go:269-272`.

**Masking**: No golden-diff fixtures exist for Subscribe, Renew, or
CreatePullPointSubscription yet (those ops require stateful request sequences).
`GetServiceCapabilities` fixture happens to match because test config has
`events_enable=3`.

**Fix window**: Phase 6 (before fixture capture for events stateful ops).
- Items 1 & 2: extract `scheme + r.Host` and pass via `ServiceContext.BaseURL`
  (same approach as PTZ/device).
- Item 3: use a proper UUID v4 generator (e.g. `crypto/rand`-based) for
  `%MSG_UUID%`; parse `wsa:MessageID` from the incoming SOAP header for
  `%REL_TO_UUID%`.
- Item 4: map `cfg.EventsEnable & 1` → `%EVENTS_BASESUBSCRIPTION%`,
  `cfg.EventsEnable & 2` → `%EVENTS_PULLPOINT%`.

---

## Procedure for adding a new gap

1. Assign next `G-nnn` id.
2. Document: summary, locations (absolute paths), masking, fix window.
3. Link from the relevant phase in `~/.windsurf/plans/onvif-go-c-parity-v3-46c14e.md` if the gap blocks or shapes that phase.
