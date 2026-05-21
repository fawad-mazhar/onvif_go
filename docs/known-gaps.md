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

### G-002 — Plain-text password accepted

**Summary**: Go's `ValidateUsernameToken` accepts a plain-text password
with no Nonce/Created fields (`@/Users/fawadmazhar/github/codes/oma/public/onvif_go/internal/auth/auth.go:42-48`).
C rejects this path (`auth_error = 3 or 4`).

**Masking**: Go is more permissive; fixtures captured with fresh
digest tokens don't exercise this path.

**Fix window**: Phase 3 (auth hardening) — delete the 6-line plain-text branch.
Phase 0e window lapsed without action; deferred here to avoid scope-creep
during media/PTZ parity work.

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

**Fix window**: Phase 2/3 — plumb `*http.Request` (or at minimum the
`Host:` header) down into `sendSOAPError`. Phase 1 left the in-code
comment at `onvif_server.go:320-321` stale; updated separately.

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

## Opened 2026-05-13 (Phase 0d follow-up)

### G-010 — `internal/exec`: deferred shell-exec security scope

**Summary**: Phase 0d landed the no-shell guarantee (`exec.Command`, no
`sh -c`), absolute-path check, `DefaultTimeout` (5 s), and `RunFmt` for
multi-arg templates. Three plan-specified items were explicitly deferred:

1. **Command allow-list not implemented.** `tokenize` only checks that
   `argv[0]` starts with `/`. The plan called for restricting executions to
   a configured binary-prefix allow-list (e.g. only binaries under
   `/usr/local/bin/`). Currently any absolute path is accepted.
   *Masking*: PTZ command strings originate from the config file only — not
   from SOAP request bodies. An attacker who can write the config file has
   already won; the allow-list guards against misconfiguration, not SOAP
   injection. Low blast radius at current scope.

2. **No per-command output size cap in `Output()`.** A misbehaving script
   that writes megabytes to stdout fills `bytes.Buffer` unbounded. The C
   reference caps reads via `fgets(out, MAX_LEN, fp)`. Add a
   `io.LimitReader(cmd.Stdout, 64*1024)` or document the assumption.

3. **Injection corpus test is owed at the SOAP-handler layer.** The corpus
   test in `exec_test.go` confirms that shell metacharacters (`;`, `$(...)`,
   `` `...` ``, `&&`, `|`, `>>`) become inert argv tokens after tokenization.
   However, SOAP-derived string args (`PresetName`, etc.) can carry
   `--flag`-style tokens that the wrapped script may honour. Validation of
   SOAP-derived command arguments belongs in the PTZ handler layer, not here.

**Locations**:
- `@/Users/fawadmazhar/github/codes/oma/public/onvif_go/internal/exec/exec.go`

**Masking**: No SOAP-derived args reach `exec` in Phase 0d — the PTZ
handlers that call `Run*`/`Output` don't exist yet.

**Fix window**: Phase 4 (PTZ handlers).

**Block-merge condition for Phase 4**: SOAP-arg validation must exist at
the handler layer and reject the character set: `;`, `$(...)`, `` `...` ``,
`&&`, `|`, `>>`, and `--`-prefixed tokens used as non-positional flags.

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

## Procedure for adding a new gap

1. Assign next `G-nnn` id.
2. Document: summary, locations (absolute paths), masking, fix window.
3. Link from the relevant phase in `~/.windsurf/plans/onvif-go-c-parity-v3-46c14e.md` if the gap blocks or shapes that phase.
