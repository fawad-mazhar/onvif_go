// Package xml: structured SOAP fault rendering, byte-compatible with the
// C reference server (onvif_simple_server/fault.c).
//
// The C reference emits four distinct fault/error response shapes:
//
//   1. send_fault()            -> generic_files/Fault.xml
//   2. send_empty_response()   -> generic_files/Empty.xml
//   3. send_pull_messages_fault() -> generic_files/PullMessagesFaultResponse.xml
//   4. send_authentication_error() -> generic_files/AuthenticationError.xml
//
// Go quirks we replicate verbatim for byte-identical output:
//   - MessageID is wrapped in LITERAL quotes: `"urn:uuid:..."`
//   - Fault.xml template does NOT contain an <?xml ...?> prolog
//   - AuthenticationError.xml DOES contain a prolog (with lowercase "utf-8")
//   - HTTP status: 500 for Fault, 400 for AuthenticationError

package xml

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"path/filepath"
)

// GenericTemplateDir is the default directory containing the four generic
// fault/response templates. Tests can override it.
var GenericTemplateDir = "service_files/generic"

// Fault describes a SOAP fault per ONVIF conventions.
type Fault struct {
	// Service is the lowercase service name (e.g. "device_service"),
	// used to build the service_address value in the fault header.
	Service string

	// DeviceAddress (http://host:port/onvif) and ServiceAddress
	// (http://host:port/onvif/<service>) echo into wsa5:ReplyTo/To.
	// If empty, callers should fill from their config/request context.
	DeviceAddress  string
	ServiceAddress string

	// RecSend is "Receiver" or "Sender" per SOAP 1.2 Code.Value.
	RecSend string

	// Subcode and SubcodeEx are ONVIF error codes, e.g. "ter:Action"
	// and "ter:ActionNotSupported".
	Subcode   string
	SubcodeEx string

	// Reason is a short human message; Detail is a longer one.
	Reason string
	Detail string
}

// FaultAddrs derives the device and service address strings from r,
// matching the C reference send_fault() convention:
//
//	DeviceAddress  = "http://" + r.Host + "/onvif"
//	ServiceAddress = "http://" + r.Host + r.URL.Path
//
// TODO scheme: hardcoded http:// matches C reference; revisit if HTTPS-fronted.
func FaultAddrs(r *http.Request) (device, service string) {
	base := "http://" + r.Host
	return base + "/onvif", base + r.URL.Path
}

// RenderFault renders f against service_files/generic/Fault.xml and returns
// the XML body plus the HTTP status to emit (500). It is byte-compatible
// with the C reference's send_fault() output.
func RenderFault(f Fault) ([]byte, int, error) {
	uuid, err := genUUID()
	if err != nil {
		return nil, 0, err
	}

	path := filepath.Join(GenericTemplateDir, "Fault.xml")
	body, err := ProcessTemplate(path, map[string]string{
		"%UUID%":       uuid,
		"%ADDRESS%":    f.DeviceAddress,
		"%SERVICE%":    f.ServiceAddress,
		"%REC_SEND%":   f.RecSend,
		"%SUBCODE%":    f.Subcode,
		"%SUBCODE_EX%": f.SubcodeEx,
		"%REASON%":     f.Reason,
		"%DETAIL%":     f.Detail,
	})
	if err != nil {
		return nil, 0, err
	}
	return []byte(body), http.StatusInternalServerError, nil
}

// RenderAuthenticationError renders the AuthenticationError.xml template.
// Byte-compatible with the C reference's send_authentication_error().
func RenderAuthenticationError() ([]byte, int, error) {
	path := filepath.Join(GenericTemplateDir, "AuthenticationError.xml")
	body, err := ProcessTemplate(path, nil)
	if err != nil {
		return nil, 0, err
	}
	return []byte(body), http.StatusBadRequest, nil
}

// RenderEmpty renders an empty <%NS%:%METHOD%Response/> body using
// generic_files/Empty.xml. ns is the namespace prefix (e.g. "tds"),
// method is the op name (e.g. "SystemReboot").
func RenderEmpty(ns, method string) ([]byte, int, error) {
	path := filepath.Join(GenericTemplateDir, "Empty.xml")
	body, err := ProcessTemplate(path, map[string]string{
		"%METHOD%": fmt.Sprintf("%s:%sResponse", ns, method),
	})
	if err != nil {
		return nil, 0, err
	}
	return []byte(body), http.StatusOK, nil
}

// RenderPullMessagesFault renders the PullMessages-specific fault
// response. Byte-compatible with the C reference's
// send_pull_messages_fault(). maxTimeout is ISO-8601 duration,
// maxMessageLimit is integer-as-string.
func RenderPullMessagesFault(maxTimeout, maxMessageLimit string) ([]byte, int, error) {
	path := filepath.Join(GenericTemplateDir, "PullMessagesFaultResponse.xml")
	body, err := ProcessTemplate(path, map[string]string{
		"%MAX_TIMEOUT%":       maxTimeout,
		"%MAX_MESSAGE_LIMIT%": maxMessageLimit,
	})
	if err != nil {
		return nil, 0, err
	}
	return []byte(body), http.StatusInternalServerError, nil
}

// WriteFault is a convenience that renders and writes in one shot.
func WriteFault(w http.ResponseWriter, f Fault) error {
	body, status, err := RenderFault(f)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}

// WriteEmpty writes an empty <ns:methodResponse/> body.
func WriteEmpty(w http.ResponseWriter, ns, method string) error {
	body, status, err := RenderEmpty(ns, method)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}

// WriteAuthenticationError writes a standard auth-failed fault.
func WriteAuthenticationError(w http.ResponseWriter) error {
	body, status, err := RenderAuthenticationError()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}

// genUUID returns an RFC-4122 v4 UUID string like C's gen_uuid(). The C
// server wraps this value in literal quotes inside the Fault.xml template
// (see %UUID% in Fault.xml line 7), so the string returned here is
// plain (no surrounding quotes).
func genUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Version 4
	b[6] = (b[6] & 0x0f) | 0x40
	// Variant 10
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
