# Development Status

*Note: Status reflects parity against the C reference server
(`onvif_simple_server`), which is the primary goal of this project. ✅ means
the Go response is byte-identical to the C reference after c14n + scrubbing.*

## Phase 0 Foundation (complete as of 2026-05-13)

| Phase | Deliverable | Commit |
|-------|-------------|--------|
| 0a | Fault library, XML composer, scrubber, c14n harness | — |
| 0b | `adv_fault_if_unknown` / `adv_fault_if_set` + C-format config parser | 97cd2ec |
| 0c | JSON loader + extension dispatch + sparse-defaults parity | 34b0de9 |
| 0d | Shell-exec helper + timeout + gzip template loader | 5a30100 |
| 0d+ | `RunFmt`, injection corpus gate, `gzipRC.Close` fix | 5a30100+ |
| 0e | Logger cleanup, CLI flags (`-log-level`, `-template-dir`), G-011 filed | — |

---

## 1. Device Service (`/onvif/device_service`)

| Operation | Routed | Parity | Notes |
|-----------|--------|--------|-------|
| GetCapabilities | ✅ | ⚠️ | Template diverges from C reference (G-008) |
| GetDeviceInformation | ✅ | ⚠️ | Template diverges |
| GetServices | ✅ | ⚠️ | Template diverges |
| GetSystemDateAndTime | ✅ | ⚠️ | Volatile datetime fields; scrubber masks |
| SystemReboot | ✅ | ⚠️ | Template diverges |
| GetScopes | ✅ | ⚠️ | Template diverges |
| GetUsers | ✅ | ⚠️ | Template diverges |
| GetWsdlUrl | ✅ | ⚠️ | Template diverges |
| GetNetworkInterfaces | ✅ | ⚠️ | Template diverges |
| GetDiscoveryMode | ✅ | ⚠️ | Template diverges |
| GetServiceCapabilities | ❌ | ❌ | Not routed; falls to unsupported handler |

**Fix window**: Phase 1 — replace all device service templates with C verbatim
copies; add `GetServiceCapabilities` route; fix auth exemption (G-001).

---

## 2. Media Service (`/onvif/media_service`)

| Operation | Routed | Parity | Notes |
|-----------|--------|--------|-------|
| GetServiceCapabilities | ✅ | ⚠️ | Template diverges |
| GetProfiles | ✅ | ⚠️ | Template diverges; profiles now loaded correctly (Phase 0b) |
| GetProfile | ✅ | ⚠️ | Template diverges |
| GetStreamUri | ✅ | ⚠️ | Template diverges; G-009 (prefix fragility) |
| GetSnapshotUri | ✅ | ⚠️ | Template diverges; G-009 |
| CreateProfile | ✅ | ⚠️ | Template diverges |
| DeleteProfile | ✅ | ⚠️ | Template diverges |
| SetVideo/Audio*Configuration | ✅ | ✅ | `adv_fault_if_set` fault matches C (Phase 0b) |
| GetVideoSources | ❌ | ❌ | Not routed |
| GetVideoSourceConfiguration | ❌ | ❌ | Not routed |
| GetVideoEncoderConfiguration | ❌ | ❌ | Not routed |
| GetAudioSources | ❌ | ❌ | Not routed |
| GetAudioSourceConfigurations | ❌ | ❌ | Not routed |
| GetAudioEncoderConfigurations | ❌ | ❌ | Not routed |

**Fix window**: Phase 3 — Media service template parity + missing routes.

---

## 3. Events Service (`/onvif/events_service`)

| Operation | Routed | Parity | Notes |
|-----------|--------|--------|-------|
| GetServiceCapabilities | ✅ | ⚠️ | Template diverges |
| GetEventProperties | ✅ | ⚠️ | Template diverges |
| CreatePullPointSubscription | ❌ | ❌ | Not routed |
| PullMessages | ❌ | ❌ | Not routed |
| Renew | ❌ | ❌ | Not routed |
| Unsubscribe | ❌ | ❌ | Not routed |
| SetSynchronizationPoint | ❌ | ❌ | Not routed |
| Subscribe | ❌ | ❌ | Not routed |

**Fix window**: Phase 5 — Events service.

---

## 4. WS-Discovery (UDP multicast 239.255.255.250:3702)

| Operation | Implemented | Parity | Notes |
|-----------|-------------|--------|-------|
| Hello (on startup) | ✅ | ⚠️ | Sent; template may diverge |
| Bye (on shutdown) | ✅ | ⚠️ | Sent via signal handler (G-011: race with main) |
| Probe / ProbeMatch | ✅ | ⚠️ | `parseSOAPAction` fragility (G-011) |
| Resolve / ResolveMatch | ✅ | ⚠️ | Go handles Resolve; C reference does not |

**Known gaps**: G-011 — `parseSOAPAction`/`parseMessageID` prefix fragility +
off-by-N offset bug + duplicate signal handler race.

**Fix window**: Phase 2 — replace string-index parsers with `encoding/xml`.

---

## 5. PTZ Service (`/onvif/ptz_service`)

| Operation | Routed | Notes |
|-----------|--------|-------|
| GetServiceCapabilities | ✅ | Template diverges |
| GetNodes | ✅ | Template diverges |
| GetNode | ❌ | Not routed |
| ContinuousMove | ❌ | Not routed |
| AbsoluteMove | ❌ | Not routed |
| RelativeMove | ❌ | Not routed |
| Stop | ❌ | Not routed |
| GetPresets | ❌ | Not routed |
| SetPreset | ❌ | Not routed |
| RemovePreset | ❌ | Not routed |
| GotoPreset | ❌ | Not routed |
| GotoHomePosition | ❌ | Not routed |
| SetHomePosition | ❌ | Not routed |

**Fix window**: Phase 4 — PTZ handler wiring via `internal/exec`.

---

## 6. DeviceIO Service (`/onvif/deviceio_service`)

| Operation | Routed | Parity | Notes |
|-----------|--------|--------|-------|
| GetServiceCapabilities | ✅ | ⚠️ | Template diverges |
| GetRelayOutputs | ✅ | ⚠️ | Template diverges |
| SetRelayOutputSettings | ❌ | ❌ | Not routed |
| SetRelayOutputState | ❌ | ❌ | Not routed |

**Fix window**: Phase 5 — DeviceIO service.

---

## 7. Core Infrastructure

| Component | Status | Location |
|-----------|--------|----------|
| HTTP server + router | ✅ | `internal/server/onvif_server.go` |
| SOAP action routing | ✅ | `internal/server/onvif_server.go` |
| `adv_fault_if_unknown` / `adv_fault_if_set` | ✅ | `internal/server/onvif_server.go` |
| Config loader (.conf + .json + extension dispatch) | ✅ | `internal/config/` |
| XML template engine (plain + gzip + .gz fallback) | ✅ | `internal/xml/processor.go` |
| XML composer (header/items/footer pattern) | ✅ | `internal/xml/composer.go` |
| SOAP fault rendering | ✅ | `internal/xml/fault.go` |
| WS-UsernameToken auth | ✅ | `internal/auth/auth.go` |
| Shell-exec helper (Run / RunFmt / Output + timeout) | ✅ | `internal/exec/exec.go` |
| WS-Discovery (UDP multicast) | ✅ | `internal/server/wsd_server.go` |
| Notification server | ✅ | `internal/server/notification_server.go` |
| Structured logging (logrus) | ✅ | `internal/logger/logger.go` |
| Golden-diff test harness | ✅ | `test/golden_diff_test.go` |
| c14n + scrubber | ✅ | `internal/xml/c14n.go`, `internal/xml/scrubber.go` |

---

## Legend

- ✅ **Implemented and correct** — matches C reference or is Go-only infrastructure
- ⚠️ **Routed but template diverges** — handler fires, output differs from C reference
- ❌ **Not implemented** — route not wired or feature absent

---

## Quick Reference

| File | Purpose |
|------|---------|
| `cmd/onvif-server/main.go` | Entry point; flags: `-config`, `-log-level`, `-template-dir` |
| `internal/server/onvif_server.go` | HTTP router + SOAP action dispatch |
| `internal/server/wsd_server.go` | WS-Discovery UDP multicast |
| `internal/config/config.go` | `.conf` parser + `Load()` dispatcher |
| `internal/config/config_json.go` | JSON config loader |
| `internal/xml/processor.go` | Template engine (plain + gzip) |
| `internal/xml/composer.go` | Header/items/footer XML composer |
| `internal/exec/exec.go` | Secure PTZ command runner |
| `internal/auth/auth.go` | WS-UsernameToken validation |
| `test/golden_diff_test.go` | Fixture-based parity regression tests |