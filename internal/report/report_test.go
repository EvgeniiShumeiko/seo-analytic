package report

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"seo-analytic/internal/fetch"
)

// Невыбранный раздел пропадает из JSON, выбранный пустой — остаётся как [].
func TestBuildSections(t *testing.T) {
	resp := fetch.Response{URL: "u", Status: 200, Header: http.Header{},
		Body: `<html><head><meta property="og:title" content="x"></head></html>`}
	r, err := Build(fetch.Desktop, resp, nil, Options{Sections: []string{Twitter}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"twitter":[]`) || strings.Contains(s, `"seo"`) || strings.Contains(s, `"open_graph"`) {
		t.Errorf("unexpected JSON: %s", s)
	}
}
