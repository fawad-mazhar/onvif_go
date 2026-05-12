package xml

import (
	"bytes"
	"testing"
)

func TestScrub_Idempotent(t *testing.T) {
	corpus := [][]byte{
		[]byte(`<wsa5:To>http://127.0.0.1:8080/onvif/media_service</wsa5:To>`),
		[]byte(`<tt:Uri>rtsp://10.0.0.5:554/ch0_0.h264</tt:Uri>`),
		[]byte(`<tt:UTCDateTime><tt:Time>00</tt:Time><tt:Date>2026-05-12</tt:Date></tt:UTCDateTime>`),
		[]byte(`<tt:DaylightSavings>false</tt:DaylightSavings>`),
		[]byte(`<wsa5:MessageID>"urn:uuid:abcd"</wsa5:MessageID>`),
		[]byte(`<tt:HwAddress>aa:bb:cc:dd:ee:ff</tt:HwAddress>`),
		[]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Envelope/>"),
	}
	for _, c := range corpus {
		first := Scrub(c)
		second := Scrub(first)
		if !bytes.Equal(first, second) {
			t.Errorf("Scrub not idempotent:\n  in:  %s\n  1st: %s\n  2nd: %s", c, first, second)
		}
	}
}

func TestScrub_HostPort_PreservesPath(t *testing.T) {
	// host:port is normalized but the path is preserved so that future
	// service-routing bugs (e.g. wrong service in a fault's wsa5:To)
	// still appear in the diff.
	in := []byte(`<a>http://127.0.0.1:8080/onvif/media_service</a><b>https://example.test:9999/x</b><c>http://h:80/onvif</c>`)
	out := Scrub(in)
	want := []byte(`<a>http://HOSTSCRUBBED/onvif/media_service</a><b>http://HOSTSCRUBBED/x</b><c>http://HOSTSCRUBBED/onvif</c>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_HostOnly(t *testing.T) {
	// URL with no path renders without the optional capture group.
	in := []byte(`<u>http://h:80</u>`)
	out := Scrub(in)
	want := []byte(`<u>http://HOSTSCRUBBED</u>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_RTSP(t *testing.T) {
	in := []byte(`<u>rtsp://host:554/stream0</u>`)
	out := Scrub(in)
	want := []byte(`<u>rtsp://HOSTSCRUBBED/stream0</u>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_PreservesAttributeURLs(t *testing.T) {
	// xmlns URIs and wsa:Action values live in attributes; they must
	// NOT be scrubbed or we break XML namespace resolution.
	in := []byte(`<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope" xmlns:tt="http://www.onvif.org/ver10/schema"><a>http://host:80/onvif</a></env:Envelope>`)
	out := Scrub(in)
	if !bytes.Contains(out, []byte(`xmlns:env="http://www.w3.org/2003/05/soap-envelope"`)) {
		t.Errorf("xmlns URI was scrubbed: %s", out)
	}
	if !bytes.Contains(out, []byte(`>http://HOSTSCRUBBED/onvif</a>`)) {
		t.Errorf("element-content URL not scrubbed correctly: %s", out)
	}
}

func TestScrub_UTCDateTime(t *testing.T) {
	in := []byte(`<tt:UTCDateTime>
  <tt:Time><tt:Hour>10</tt:Hour></tt:Time>
  <tt:Date>2026-05-12</tt:Date>
</tt:UTCDateTime>`)
	out := Scrub(in)
	want := []byte(`<tt:UTCDateTime>UTCSCRUBBED</tt:UTCDateTime>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_MessageID_Quoted(t *testing.T) {
	in := []byte(`<wsa5:MessageID>"urn:uuid:76931fac-dab2"</wsa5:MessageID>`)
	out := Scrub(in)
	want := []byte(`<wsa5:MessageID>UUIDSCRUBBED</wsa5:MessageID>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_MessageID_Unquoted(t *testing.T) {
	in := []byte(`<wsa:MessageID>urn:uuid:abcd</wsa:MessageID>`)
	out := Scrub(in)
	want := []byte(`<wsa:MessageID>UUIDSCRUBBED</wsa:MessageID>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_HwAddress(t *testing.T) {
	in := []byte(`<tt:HwAddress>aa:bb:cc:dd:ee:ff</tt:HwAddress>`)
	out := Scrub(in)
	want := []byte(`<tt:HwAddress>MACSCRUBBED</tt:HwAddress>`)
	if !bytes.Equal(out, want) {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestScrub_StripsProlog(t *testing.T) {
	in := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Envelope/>")
	out := Scrub(in)
	if bytes.HasPrefix(out, []byte("<?xml")) {
		t.Errorf("prolog still present: %s", out)
	}
	if !bytes.Contains(out, []byte("<Envelope/>")) {
		t.Errorf("lost body: %s", out)
	}
}
