package xml

import "testing"

func TestExtractBodyAction_Prefixes(t *testing.T) {
	cases := map[string]string{
		"s-prefix": `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
  <s:Body><tds:GetUsers/></s:Body>
</s:Envelope>`,
		"soap-prefix": `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body><GetCapabilities/></soap:Body>
</soap:Envelope>`,
		"SOAP-ENV-prefix": `<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body><tds:GetServices xmlns:tds="urn:x"/></SOAP-ENV:Body>
</SOAP-ENV:Envelope>`,
		"env-prefix": `<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope">
  <env:Body><tds:GetWsdlUrl xmlns:tds="urn:x"/></env:Body>
</env:Envelope>`,
	}
	want := map[string]string{
		"s-prefix":        "GetUsers",
		"soap-prefix":     "GetCapabilities",
		"SOAP-ENV-prefix": "GetServices",
		"env-prefix":      "GetWsdlUrl",
	}
	for name, body := range cases {
		got, err := ExtractBodyAction([]byte(body))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if got != want[name] {
			t.Errorf("%s: got %q, want %q", name, got, want[name])
		}
	}
}

func TestExtractBodyAction_Errors(t *testing.T) {
	cases := map[string]string{
		"no-body":    `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"/>`,
		"empty-body": `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body/></s:Envelope>`,
		"malformed":  `<s:Envelope`,
	}
	for name, body := range cases {
		_, err := ExtractBodyAction([]byte(body))
		if err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestExtractUsernameToken(t *testing.T) {
	body := `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Header>
    <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
      <wsse:UsernameToken>
        <wsse:Username>admin</wsse:Username>
        <wsse:Password Type="...PasswordDigest">qXdz=</wsse:Password>
        <wsse:Nonce EncodingType="...Base64Binary">LKqI6G/AikKCQrN0zqZFlg==</wsse:Nonce>
        <wsu:Created xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">2026-05-12T11:00:00Z</wsu:Created>
      </wsse:UsernameToken>
    </wsse:Security>
  </s:Header>
  <s:Body><GetUsers/></s:Body>
</s:Envelope>`
	tok, err := ExtractUsernameToken([]byte(body))
	if err != nil {
		t.Fatalf("ExtractUsernameToken: %v", err)
	}
	if tok.Username != "admin" {
		t.Errorf("Username = %q", tok.Username)
	}
	if tok.Password != "qXdz=" {
		t.Errorf("Password = %q", tok.Password)
	}
	if tok.Nonce != "LKqI6G/AikKCQrN0zqZFlg==" {
		t.Errorf("Nonce = %q", tok.Nonce)
	}
	if tok.Created != "2026-05-12T11:00:00Z" {
		t.Errorf("Created = %q", tok.Created)
	}
}

func TestExtractUsernameToken_PrefixAgnostic(t *testing.T) {
	// Non-canonical prefix (`u:` instead of `wsse:`); local names match.
	body := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Header>
    <u:Security xmlns:u="urn:wsse">
      <u:UsernameToken>
        <u:Username>bob</u:Username>
        <u:Password>pw</u:Password>
      </u:UsernameToken>
    </u:Security>
  </s:Header>
  <s:Body><X/></s:Body>
</s:Envelope>`
	tok, err := ExtractUsernameToken([]byte(body))
	if err != nil {
		t.Fatalf("ExtractUsernameToken: %v", err)
	}
	if tok.Username != "bob" || tok.Password != "pw" {
		t.Errorf("got %+v", tok)
	}
}

func TestExtractElement(t *testing.T) {
	cases := []struct {
		name      string
		xml       string
		localName string
		want      string
	}{
		{
			name:      "body-element",
			xml:       `<s:Envelope><s:Body><trt:ProfileToken>Profile0</trt:ProfileToken></s:Body></s:Envelope>`,
			localName: "ProfileToken",
			want:      "Profile0",
		},
		{
			name:      "header-element",
			xml:       `<s:Envelope><s:Header><wsa:Action>http://probe</wsa:Action></s:Header><s:Body/></s:Envelope>`,
			localName: "Action",
			want:      "http://probe",
		},
		{
			name:      "whitespace-trimmed",
			xml:       `<env><ProfileToken>  abc  </ProfileToken></env>`,
			localName: "ProfileToken",
			want:      "abc",
		},
		{
			name:      "case-insensitive-match",
			xml:       `<env><PROFILETOKEN>tok1</PROFILETOKEN></env>`,
			localName: "profiletoken",
			want:      "tok1",
		},
		{
			name:      "first-match-returned",
			xml:       `<env><X>first</X><X>second</X></env>`,
			localName: "X",
			want:      "first",
		},
		{
			name:      "absent-element-returns-empty",
			xml:       `<env><s:Body/></env>`,
			localName: "ProfileToken",
			want:      "",
		},
	}
	for _, tc := range cases {
		got, err := ExtractElement([]byte(tc.xml), tc.localName)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestExtractUsernameToken_Missing(t *testing.T) {
	body := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><X/></s:Body></s:Envelope>`
	_, err := ExtractUsernameToken([]byte(body))
	if err == nil {
		t.Errorf("expected error for missing UsernameToken")
	}
}
