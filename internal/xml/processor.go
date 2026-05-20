// Package xml provides XML template processing utilities for ONVIF services.
package xml

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

// gzipRC wraps a gzip.Reader and its underlying file so both are closed together.
type gzipRC struct {
	*gzip.Reader
	f *os.File
}

func (g gzipRC) Close() error {
	rerr := g.Reader.Close()
	ferr := g.f.Close()
	if rerr != nil {
		return rerr
	}
	return ferr
}

// openTemplate opens filename for reading, transparently decompressing gzip
// content when filename ends with ".gz". If filename is not found and does not
// already end with ".gz", openTemplate automatically retries with
// filename+".gz" so callers can be format-agnostic.
func openTemplate(filename string) (io.ReadCloser, error) {
	f, err := os.Open(filename)
	if err != nil {
		if !strings.HasSuffix(filename, ".gz") {
			gzf, gerr := os.Open(filename + ".gz")
			if gerr == nil {
				gz, gerr2 := gzip.NewReader(gzf)
				if gerr2 == nil {
					return gzipRC{gz, gzf}, nil
				}
				_ = gzf.Close()
			}
		}
		return nil, err
	}
	if strings.HasSuffix(filename, ".gz") {
		gz, gerr := gzip.NewReader(f)
		if gerr != nil {
			_ = f.Close()
			return nil, gerr
		}
		return gzipRC{gz, f}, nil
	}
	return f, nil
}

// ProcessTemplate reads an XML template file and replaces placeholders with
// provided values. Files ending in ".gz" are transparently decompressed; plain
// ".xml" paths also fall back to ".xml.gz" if the plain file is not found.
func ProcessTemplate(filename string, replacements map[string]string) (string, error) {
	rc, err := openTemplate(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %v", err)
	}
	defer func() { _ = rc.Close() }()

	var result strings.Builder
	scanner := bufio.NewScanner(rc)

	for scanner.Scan() {
		line := scanner.Text()

		// Replace all placeholders in the line
		for placeholder, replacement := range replacements {
			line = strings.ReplaceAll(line, placeholder, replacement)
		}

		// Replicate C cat() semantics: trim each line, skip empty lines, then
		// concatenate — elements starting with '<' are joined directly, all
		// other text gets a single leading space (matching the C reference output
		// so that c14n comparison against captured C fixtures passes).
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "<") {
			result.WriteByte(' ')
		}
		result.WriteString(line)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading template file: %v", err)
	}

	return result.String(), nil
}

// FileExists reports whether filename (or filename+".gz") exists on disk.
// The ".gz" fallback mirrors the openTemplate fallback, so callers using
// FileExists as a pre-check remain consistent with ProcessTemplate.
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	if !strings.HasSuffix(filename, ".gz") {
		_, err := os.Stat(filename + ".gz")
		return err == nil
	}
	return false
}
