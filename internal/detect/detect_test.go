package detect

import (
	"net/http"
	"slices"
	"testing"

	"seo-analytic/internal/page"
)

const fixture = `<!doctype html><html><head>
<meta name="generator" content="WordPress 6.4.2">
<meta name="yandex-verification" content="abc123">
<link rel="stylesheet" href="/wp-content/themes/x/style.css">
<script src="/wp-includes/js/jquery/jquery-3.7.1.min.js"></script>
<script async src="https://www.googletagmanager.com/gtag/js?id=G-TEST123"></script>
<script>ym(98765, "init", {}); fbq('init', '555111');</script>
</head><body></body></html>`

func metas(t *testing.T) []page.Meta {
	t.Helper()
	p, err := page.Parse(fixture)
	if err != nil {
		t.Fatal(err)
	}
	return p.Metas
}

func headers() http.Header {
	h := http.Header{}
	h.Set("Server", "nginx")
	h.Set("X-Powered-By", "PHP/8.2.1")
	return h
}

func byName(hits []Hit) map[string][]string {
	m := map[string][]string{}
	for _, h := range hits {
		m[h.Name] = h.Details
	}
	return m
}

func TestTrackers(t *testing.T) {
	got := byName(Trackers(fixture, metas(t)))
	want := map[string]string{
		"Google Analytics 4": "G-TEST123",
		"Яндекс.Метрика":     "98765",
		"Facebook Pixel":     "555111",
		"Яндекс.Вебмастер":   "abc123",
	}
	for name, id := range want {
		if !slices.Contains(got[name], id) {
			t.Errorf("%s: want %s, got %v", name, id, got[name])
		}
	}
	if _, ok := got["VK Pixel"]; ok {
		t.Error("VK Pixel false positive")
	}
}

func TestTech(t *testing.T) {
	hits, langs := Tech(fixture, headers(), metas(t))
	got := byName(hits)
	for _, name := range []string{"nginx", "WordPress 6.4.2", "WordPress", "PHP", "jQuery"} {
		if _, ok := got[name]; !ok {
			t.Errorf("%s not detected, got %v", name, got)
		}
	}
	if !slices.Contains(got["PHP"], "8.2.1") || !slices.Contains(got["jQuery"], "3.7.1") {
		t.Errorf("versions: PHP %v, jQuery %v", got["PHP"], got["jQuery"])
	}
	if !slices.Equal(langs, []string{"PHP"}) {
		t.Errorf("langs = %v", langs)
	}
	if _, ok := got["Next.js"]; ok {
		t.Error("Next.js false positive")
	}
}

func TestTechDeep(t *testing.T) {
	// __NEXT_DATA__ Wappalyzer без рендеринга не видит — должен подхватиться из простого анализа.
	hits, langs, err := TechDeep(fixture+`<script id="__NEXT_DATA__"></script>`, headers(), metas(t))
	if err != nil {
		t.Fatal(err)
	}
	got := byName(hits)
	for _, name := range []string{"WordPress", "PHP", "Nginx", "Next.js"} {
		if _, ok := got[name]; !ok {
			t.Errorf("%s not detected, got %v", name, got)
		}
	}
	if !slices.Contains(langs, "PHP") {
		t.Errorf("langs = %v", langs)
	}
}
