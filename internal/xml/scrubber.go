// Package xml: volatile-field scrubber for parity golden-diffs.
//
// The C reference and Go server will never produce byte-identical output
// for fields whose value is derived from the runtime environment (clock,
// host:port, MAC address, generated UUIDs). The scrubber normalizes
// those fields to placeholder tokens on BOTH sides of the diff so the
// remaining bytes can be compared directly.
//
// All passes are idempotent: Scrub(Scrub(x)) == Scrub(x).
//
// Covered volatile fields (see docs/testing-strategy.md for the
// exhaustive list):
//
//   - http(s)://HOST:PORT/... -> http://<HOST>/...
//   - rtsp://HOST:PORT/...     -> rtsp://<HOST>/...
//   - <tt:UTCDateTime>...</tt:UTCDateTime>        -> normalized placeholder
//   - <tt:LocalDateTime>...</tt:LocalDateTime>    -> normalized placeholder
//   - <tt:DaylightSavings>...</tt:DaylightSavings> -> <DST/>
//   - <wsa5:MessageID>..."urn:uuid:..."</wsa5:MessageID> -> <UUID/>
//   - <wsa5:MessageID>urn:uuid:...</wsa5:MessageID>      -> <UUID/>
//   - <tt:HwAddress>...</tt:HwAddress>            -> <MAC/>
//   - Leading <?xml ...?> prolog stripped         (C faults omit it)

package xml

import (
	"fmt"
	"regexp"
)

// Go's regexp (RE2) does not support backreferences, so we enumerate the
// namespace prefixes that actually appear in the C-reference and Go
// outputs. If a new prefix shows up in a future fixture, add it here.
var knownPrefixes = []string{"", "tt:", "wsa5:", "wsa:", "wsnt:", "env:"}

// stripElementRE builds a regex that matches <prefix-NAME>...</prefix-NAME>
// for each known prefix, replacing the inner content with the given marker.
// Returns one compiled regex per prefix (concatenated alternation is
// avoided so $1 replacement semantics stay simple).
func stripElementRE(name string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(knownPrefixes))
	for _, p := range knownPrefixes {
		// (?s) => dot matches newline, lazy between open/close.
		pat := fmt.Sprintf(`(?s)<%s%s\b[^>]*>.*?</%s%s>`, p, name, p, name)
		out = append(out, regexp.MustCompile(pat))
	}
	return out
}

var (
	// Only scrub URLs that appear as element content (between > and <).
	// Skips URLs in attribute values (xmlns namespace declarations,
	// wsa:Action URIs) which are stable and must not be rewritten.
	//
	// The host:port portion is replaced with HOSTSCRUBBED but the
	// path is preserved, so a future bug that emits the wrong service
	// path (e.g. /onvif/device_service vs /onvif/media_service in a
	// fault's wsa5:To) still shows up in the diff. Group $1 captures
	// the path including the leading slash, or "" if absent.
	reHTTPInText = regexp.MustCompile(`>https?://[^/<"\s]+(/[^<"\s]*)?`)
	reRTSPInText = regexp.MustCompile(`>rtsp://[^/<"\s]+(/[^<"\s]*)?`)

	reXMLProlog = regexp.MustCompile(`^\s*<\?xml[^?]*\?>\s*`)

	reUTCDateTime = stripElementRE("UTCDateTime")
	reLocalDT     = stripElementRE("LocalDateTime")
	reDST         = stripElementRE("DaylightSavings")
	reMessageID   = stripElementRE("MessageID")
	reHwAddress   = stripElementRE("HwAddress")
)

// Scrub applies every scrub pass to data and returns the normalized form.
// All passes are idempotent.
func Scrub(data []byte) []byte {
	data = reXMLProlog.ReplaceAll(data, nil)

	// Preserve the leading '>' so the XML structure stays valid. Use
	// plain text tokens (no angle brackets) so the token cannot itself
	// become invalid XML in any future placement.
	data = reHTTPInText.ReplaceAll(data, []byte(">http://HOSTSCRUBBED$1"))
	data = reRTSPInText.ReplaceAll(data, []byte(">rtsp://HOSTSCRUBBED$1"))

	data = replaceElement(data, reUTCDateTime, "UTCDateTime", "UTCSCRUBBED")
	data = replaceElement(data, reLocalDT, "LocalDateTime", "LOCALSCRUBBED")
	data = replaceElement(data, reDST, "DaylightSavings", "DSTSCRUBBED")
	data = replaceElement(data, reMessageID, "MessageID", "UUIDSCRUBBED")
	data = replaceElement(data, reHwAddress, "HwAddress", "MACSCRUBBED")
	return data
}

// replaceElement replaces <PREFIX-name>...</PREFIX-name> with
// <PREFIX-name>MARKER</PREFIX-name> for every known prefix.
func replaceElement(data []byte, res []*regexp.Regexp, name, marker string) []byte {
	for i, re := range res {
		p := knownPrefixes[i]
		// Replacement uses the literal prefix + name so no backrefs needed.
		repl := []byte(fmt.Sprintf(`<%s%s>%s</%s%s>`, p, name, marker, p, name))
		data = re.ReplaceAll(data, repl)
	}
	return data
}
