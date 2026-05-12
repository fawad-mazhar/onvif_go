#!/usr/bin/env bash
# capture_fixtures.sh - P0.2: capture C-reference golden fixtures.
#
# POSTs a SOAP body to the running C reference server (see run_c_server.sh)
# for each op in the manifest below and writes paired files to
#   test/fixtures/<service>/<Op>.request.xml
#   test/fixtures/<service>/<Op>.response.xml
#
# The captured response is the parity oracle used by c14n golden-diff
# in CI (P0.5 + 0a).
#
# Skipped intentionally:
#   - SystemReboot (would reboot the container)
#   - Stateful events ops (CreatePullPointSubscription, PullMessages,
#     Subscribe, Renew, Unsubscribe) - captured separately in a stateful
#     second-pass script once the registry semantics are confirmed.
#   - DeleteProfile (Go-only; will be removed for strict parity in Phase 1)
#   - WS-Discovery (UDP multicast, not HTTP/CGI)

set -euo pipefail

HOST="${HOST:-http://localhost:18080}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FIX_ROOT="${REPO_ROOT}/test/fixtures"

NS_S='xmlns:s="http://www.w3.org/2003/05/soap-envelope"'
NS_TDS='xmlns:tds="http://www.onvif.org/ver10/device/wsdl"'
NS_TRT='xmlns:trt="http://www.onvif.org/ver10/media/wsdl"'
NS_TPTZ='xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"'
NS_TEV='xmlns:tev="http://www.onvif.org/ver10/events/wsdl"'
NS_TMD='xmlns:tmd="http://www.onvif.org/ver10/deviceIO/wsdl"'
NS_TT='xmlns:tt="http://www.onvif.org/ver10/schema"'

soap() {
    # $1 = ns decls, $2 = body content
    printf '<?xml version="1.0" encoding="UTF-8"?>\n<s:Envelope %s %s><s:Body>%s</s:Body></s:Envelope>' \
        "$NS_S" "$1" "$2"
}

capture() {
    local service="$1" op="$2" body="$3"
    local out_dir="${FIX_ROOT}/${service}"
    mkdir -p "$out_dir"

    local req="${out_dir}/${op}.request.xml"
    local resp="${out_dir}/${op}.response.xml"

    printf '%s\n' "$body" > "$req"

    local http_code
    http_code=$(curl -s -o "$resp" -w '%{http_code}' \
        -X POST \
        -H 'Content-Type: application/soap+xml; charset=utf-8' \
        --data-binary @"$req" \
        "${HOST}/cgi-bin/onvif/${service}")

    if [[ "$http_code" != "200" ]]; then
        printf '  [!] %-30s HTTP %s\n' "${service}/${op}" "$http_code" >&2
        return 1
    fi

    local bytes
    bytes=$(wc -c < "$resp" | tr -d ' ')
    printf '  [ok] %-30s %s bytes\n' "${service}/${op}" "$bytes"
}

# ----------------------------------------------------------------------
# Device service (9 ops; SystemReboot intentionally skipped)
# ----------------------------------------------------------------------
echo "==> device_service"

capture device_service GetServices \
    "$(soap "$NS_TDS" '<tds:GetServices><tds:IncludeCapability>false</tds:IncludeCapability></tds:GetServices>')"

capture device_service GetDeviceInformation \
    "$(soap "$NS_TDS" '<tds:GetDeviceInformation/>')"

capture device_service GetCapabilities \
    "$(soap "$NS_TDS" '<tds:GetCapabilities><tds:Category>All</tds:Category></tds:GetCapabilities>')"

capture device_service GetScopes \
    "$(soap "$NS_TDS" '<tds:GetScopes/>')"

capture device_service GetSystemDateAndTime \
    "$(soap "$NS_TDS" '<tds:GetSystemDateAndTime/>')"

capture device_service GetUsers \
    "$(soap "$NS_TDS" '<tds:GetUsers/>')"

capture device_service GetWsdlUrl \
    "$(soap "$NS_TDS" '<tds:GetWsdlUrl/>')"

capture device_service GetNetworkInterfaces \
    "$(soap "$NS_TDS" '<tds:GetNetworkInterfaces/>')"

capture device_service GetDiscoveryMode \
    "$(soap "$NS_TDS" '<tds:GetDiscoveryMode/>')"

# ----------------------------------------------------------------------
# Media service (6 ops; DeleteProfile is Go-only, removed in Phase 1)
# ----------------------------------------------------------------------
echo "==> media_service"

capture media_service GetServiceCapabilities \
    "$(soap "$NS_TRT" '<trt:GetServiceCapabilities/>')"

capture media_service GetProfiles \
    "$(soap "$NS_TRT" '<trt:GetProfiles/>')"

capture media_service GetProfile \
    "$(soap "$NS_TRT" '<trt:GetProfile><trt:ProfileToken>Profile_0</trt:ProfileToken></trt:GetProfile>')"

capture media_service GetStreamUri \
    "$(soap "$NS_TRT $NS_TT" '<trt:GetStreamUri><trt:StreamSetup><tt:Stream>RTP-Unicast</tt:Stream><tt:Transport><tt:Protocol>RTSP</tt:Protocol></tt:Transport></trt:StreamSetup><trt:ProfileToken>Profile_0</trt:ProfileToken></trt:GetStreamUri>')"

capture media_service GetSnapshotUri \
    "$(soap "$NS_TRT" '<trt:GetSnapshotUri><trt:ProfileToken>Profile_0</trt:ProfileToken></trt:GetSnapshotUri>')"

capture media_service CreateProfile \
    "$(soap "$NS_TRT" '<trt:CreateProfile><trt:Name>NewProfile</trt:Name></trt:CreateProfile>')" || true

# ----------------------------------------------------------------------
# PTZ service (2 ops)
# ----------------------------------------------------------------------
echo "==> ptz_service"

capture ptz_service GetServiceCapabilities \
    "$(soap "$NS_TPTZ" '<tptz:GetServiceCapabilities/>')"

capture ptz_service GetNodes \
    "$(soap "$NS_TPTZ" '<tptz:GetNodes/>')"

# ----------------------------------------------------------------------
# Events service (2 stateless ops; stateful 6 captured separately)
# ----------------------------------------------------------------------
echo "==> events_service (stateless only)"

capture events_service GetServiceCapabilities \
    "$(soap "$NS_TEV" '<tev:GetServiceCapabilities/>')"

capture events_service GetEventProperties \
    "$(soap "$NS_TEV" '<tev:GetEventProperties/>')"

# ----------------------------------------------------------------------
# DeviceIO service (2 ops)
# ----------------------------------------------------------------------
echo "==> deviceio_service"

capture deviceio_service GetServiceCapabilities \
    "$(soap "$NS_TMD" '<tmd:GetServiceCapabilities/>')"

capture deviceio_service GetRelayOutputs \
    "$(soap "$NS_TMD" '<tmd:GetRelayOutputs/>')"

echo
echo "Captured $(find "$FIX_ROOT" -name '*.response.xml' | wc -l | tr -d ' ') response fixtures under $FIX_ROOT"
