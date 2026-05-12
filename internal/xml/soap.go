// Package xml: namespace-aware extractors for SOAP request payloads.
//
// Replaces the brittle string-index parsers that previously lived in
// internal/server (parseSOAPAction) and internal/auth (ParseSOAPHeader).
// Both shared the same root cause: they hard-coded a single prefix
// (<soap:Body>, <Username>) and silently dropped any payload that used
// a different one (<s:Body>, <wsse:Username>).
//
// Both extractors here use encoding/xml's namespace-aware decoder so
// they work regardless of the prefix the client picked.

package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// SOAP envelope namespaces accepted by ExtractBodyAction. Matches the
// SOAP 1.2 final URI; SOAP 1.1 is unused by ONVIF.
const soap12NS = "http://www.w3.org/2003/05/soap-envelope"

// WSS namespaces. Different ONVIF clients use different versions of the
// 2004/01 schema; the local-name of <Security>/<UsernameToken> children
// is the same across versions, so we accept any namespace.
// (Namespace check intentionally omitted for resilience.)

// UsernameToken is the parsed wsse:UsernameToken from a SOAP header.
type UsernameToken struct {
	Username string
	Password string
	Nonce    string
	Created  string
}

// ExtractBodyAction returns the local-name of the first child element of
// soap:Body. Returns an error if the document does not parse, has no
// Body, or has an empty Body.
func ExtractBodyAction(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	inBody := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("parse SOAP envelope: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !inBody {
			if se.Name.Local == "Body" && se.Name.Space == soap12NS {
				inBody = true
			}
			continue
		}
		// First child of Body.
		return se.Name.Local, nil
	}
	return "", fmt.Errorf("SOAP body action not found")
}

// ExtractUsernameToken returns the WS-Security UsernameToken fields, or
// an error if Username/Password are missing. Nonce and Created are
// optional. Element local-names are matched case-sensitively (per WSS
// schema); namespace is intentionally ignored to tolerate the
// 2004/01-vs-2004/06 schema split that some clients still use.
func ExtractUsernameToken(data []byte) (UsernameToken, error) {
	var t UsernameToken
	dec := xml.NewDecoder(bytes.NewReader(data))

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return t, fmt.Errorf("parse SOAP envelope: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "Username", "Password", "Nonce", "Created":
			var v string
			if err := dec.DecodeElement(&v, &se); err != nil {
				return t, fmt.Errorf("decode %s: %w", se.Name.Local, err)
			}
			switch se.Name.Local {
			case "Username":
				t.Username = v
			case "Password":
				t.Password = v
			case "Nonce":
				t.Nonce = v
			case "Created":
				t.Created = v
			}
		}
	}

	if t.Username == "" {
		return t, fmt.Errorf("username not found in SOAP header")
	}
	if t.Password == "" {
		return t, fmt.Errorf("password not found in SOAP header")
	}
	return t, nil
}
