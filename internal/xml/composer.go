// Package xml: template composer for list-shaped ONVIF responses.
//
// Many ONVIF operations (GetProfiles, GetVideoSources, GetEventProperties,
// ...) return a list whose SOAP shape is:
//
//   <Header> ... item_1 ... item_2 ... item_N ... <Footer>
//
// The C reference ships pre-split templates per op, e.g.:
//
//   media_service_files/GetProfiles_header.xml
//   media_service_files/GetProfiles_middle.xml   (one item, with %...% placeholders)
//   media_service_files/GetProfiles_footer.xml
//   media_service_files/GetProfiles_none.xml     (empty-list case)
//
// Compose walks the three files and stitches the result.

package xml

import (
	"path/filepath"
	"strings"
)

// Compose stitches header + middle*N + footer in-memory. Each string is
// treated as already-rendered XML; callers are responsible for
// substituting placeholders before passing the items in.
func Compose(header string, items []string, footer string) string {
	var sb strings.Builder
	sb.Grow(len(header) + len(footer) + 128*len(items))
	sb.WriteString(header)
	for _, it := range items {
		sb.WriteString(it)
	}
	sb.WriteString(footer)
	return sb.String()
}

// ComposeFromFiles reads the three template files and substitutes:
//   - globals in every file
//   - perItem[i] in the middle file, once per item
//
// If len(perItem) == 0 and nonePath is non-empty, nonePath is rendered
// with globals and returned as the full response; otherwise an empty
// list still produces header+footer (no middle).
func ComposeFromFiles(
	headerPath, middlePath, footerPath, nonePath string,
	globals map[string]string,
	perItem []map[string]string,
) (string, error) {

	if len(perItem) == 0 && nonePath != "" {
		return ProcessTemplate(nonePath, globals)
	}

	header, err := ProcessTemplate(headerPath, globals)
	if err != nil {
		return "", err
	}

	items := make([]string, 0, len(perItem))
	for _, sub := range perItem {
		// Merge globals + per-item; per-item wins on collision.
		merged := make(map[string]string, len(globals)+len(sub))
		for k, v := range globals {
			merged[k] = v
		}
		for k, v := range sub {
			merged[k] = v
		}
		s, err := ProcessTemplate(middlePath, merged)
		if err != nil {
			return "", err
		}
		items = append(items, s)
	}

	footer, err := ProcessTemplate(footerPath, globals)
	if err != nil {
		return "", err
	}
	return Compose(header, items, footer), nil
}

// ServiceTemplatePath returns the canonical on-disk path for a service
// template under service_files/<service>/<name>.xml.
// ServiceTemplateDir is configurable by tests.
var ServiceTemplateDir = "service_files"

func ServiceTemplatePath(service, name string) string {
	return filepath.Join(ServiceTemplateDir, service, name)
}
