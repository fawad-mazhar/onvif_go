package tests

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/server"
	xmlpkg "github.com/fawad-mazhar/onvif-go/internal/xml"
	"github.com/ucarion/c14n"
)

// fixturesDir is repo-relative.
const fixturesDir = "fixtures"

// TestGoldenDiff is the parity regression gate (Phase 0a).
//
// For each captured fixture (test/fixtures/<service>/<Op>.{request,response}.xml)
// it posts the request to an in-process Go server built against the
// canonical fixture config, scrubs both the captured C response and the
// Go response, canonicalizes both via c14n, and byte-compares.
//
// Regressions are not yet treated as failures: Phase 1-5 will gradually
// bring Go's output into byte-identical parity. Until then, this test
// LOGS pass/fail per op and only FAILS if the pass count drops below
// the baseline stored at test/fixtures/.baseline.
//
// To refresh the baseline after a legitimate improvement:
//
//	GOLDEN_UPDATE_BASELINE=1 go test ./test/ -run GoldenDiff
func TestGoldenDiff(t *testing.T) {
	cfg, err := config.LoadConfig(filepath.Join(fixturesDir, "config", "server.conf"))
	if err != nil {
		t.Fatalf("load fixture config: %v", err)
	}

	// Point xml library at the repo-root service_files/ regardless of CWD.
	cwd, _ := os.Getwd()
	xmlpkg.GenericTemplateDir = filepath.Join(cwd, "..", "service_files", "generic")
	xmlpkg.ServiceTemplateDir = filepath.Join(cwd, "..", "service_files")

	// Stand up the Go server.
	r := server.BuildRouter(cfg)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Walk the fixture tree.
	var fixtures []string
	err = filepath.WalkDir(fixturesDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".response.xml") {
			fixtures = append(fixtures, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk fixtures: %v", err)
	}
	sort.Strings(fixtures)

	if len(fixtures) == 0 {
		t.Fatalf("no fixtures found under %s", fixturesDir)
	}

	passing := map[string]bool{}
	var logLines []string
	for _, respPath := range fixtures {
		reqPath := strings.TrimSuffix(respPath, ".response.xml") + ".request.xml"
		rel := strings.TrimPrefix(respPath, fixturesDir+"/")
		rel = strings.TrimSuffix(rel, ".response.xml")

		service := filepath.Dir(rel)
		op := filepath.Base(rel)

		req, err := os.ReadFile(reqPath)
		if err != nil {
			logLines = append(logLines, fmt.Sprintf("  [skip] %-45s request missing", rel))
			continue
		}
		expected, err := os.ReadFile(respPath)
		if err != nil {
			logLines = append(logLines, fmt.Sprintf("  [skip] %-45s response missing", rel))
			continue
		}

		url := ts.URL + "/onvif/" + service
		httpReq, err := http.NewRequest("POST", url, bytes.NewReader(req))
		if err != nil {
			logLines = append(logLines, fmt.Sprintf("  [skip] %-45s req build: %v", rel, err))
			continue
		}
		httpReq.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			logLines = append(logLines, fmt.Sprintf("  [FAIL] %-45s Go HTTP: %v", rel, err))
			continue
		}
		actual, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		expCanon, expErr := canonForm(expected)
		actCanon, actErr := canonForm(actual)
		if expErr != nil || actErr != nil {
			logLines = append(logLines, fmt.Sprintf("  [FAIL] %-45s c14n err exp=%v act=%v", rel, expErr, actErr))
			continue
		}
		if bytes.Equal(expCanon, actCanon) {
			logLines = append(logLines, fmt.Sprintf("  [PASS] %-45s", rel))
			passing[service+"/"+op] = true
		} else {
			logLines = append(logLines, fmt.Sprintf("  [diff] %-45s (%d vs %d bytes canonical)", rel, len(expCanon), len(actCanon)))
		}
	}

	for _, l := range logLines {
		t.Log(l)
	}
	t.Logf("summary: %d/%d fixtures passing golden diff", len(passing), len(fixtures))

	// Enforce the baseline (no regression).
	baselinePath := filepath.Join(fixturesDir, ".baseline")
	if os.Getenv("GOLDEN_UPDATE_BASELINE") == "1" {
		if err := writeBaseline(baselinePath, passing); err != nil {
			t.Fatalf("write baseline: %v", err)
		}
		t.Logf("baseline updated: %d passing ops written to %s", len(passing), baselinePath)
		return
	}

	required, err := readBaseline(baselinePath)
	if os.IsNotExist(err) {
		t.Logf("no baseline file; creating one with the current %d passing ops", len(passing))
		if err := writeBaseline(baselinePath, passing); err != nil {
			t.Fatalf("write initial baseline: %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var missing []string
	for op := range required {
		if !passing[op] {
			missing = append(missing, op)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("regression: %d op(s) that previously passed now fail:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

func canonForm(data []byte) ([]byte, error) {
	scrubbed := xmlpkg.Scrub(data)
	dec := xml.NewDecoder(bytes.NewReader(scrubbed))
	return c14n.Canonicalize(dec)
}

func writeBaseline(path string, passing map[string]bool) error {
	ops := make([]string, 0, len(passing))
	for k := range passing {
		ops = append(ops, k)
	}
	sort.Strings(ops)
	return os.WriteFile(path, []byte(strings.Join(ops, "\n")+"\n"), 0o600)
}

func readBaseline(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out[line] = true
		}
	}
	return out, nil
}
