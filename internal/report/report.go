// Package report собирает отчёт по странице из выбранных разделов.
package report

import (
	"slices"
	"strings"

	"seo-analytic/internal/detect"
	"seo-analytic/internal/fetch"
	"seo-analytic/internal/page"
)

// Разделы отчёта.
const (
	SEO       = "seo"
	OpenGraph = "og"
	Twitter   = "twitter"
	OtherMeta = "meta"
	Trackers  = "trackers"
	Tech      = "tech"
)

var AllSections = []string{SEO, OpenGraph, Twitter, OtherMeta, Trackers, Tech}

var SectionNames = map[string]string{
	SEO:       "SEO",
	OpenGraph: "Open Graph",
	Twitter:   "Twitter Card",
	OtherMeta: "Прочие meta",
	Trackers:  "Аналитика, пиксели, верификация",
	Tech:      "Технологии и языки",
}

// Режимы анализа технологий.
const (
	TechSimple = "simple"
	TechDeep   = "deep"
	TechChrome = "chrome" // экспериментальный: deep + JS-объекты в Chrome, требует рендеринга
)

var DeviceNames = map[string]string{fetch.Desktop: "десктоп", fetch.Mobile: "мобильный"}

type Options struct {
	Sections []string
	TechMode string // TechSimple, TechDeep или TechChrome
	Render   bool   // страница отрендерена в Chrome
}

type SEOInfo struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Keywords    string      `json:"keywords"`
	Robots      string      `json:"robots"`
	Canonical   string      `json:"canonical"`
	Lang        string      `json:"lang"`
	Viewport    string      `json:"viewport"`
	H1          []string    `json:"h1"`
	Images      int         `json:"images"`
	ImagesNoAlt int         `json:"images_no_alt"`
	Hreflang    []page.Meta `json:"hreflang"`
}

// Result — отчёт по одной версии страницы. Невыбранные разделы остаются nil
// и пропадают из JSON (omitzero), а выбранные, но пустые выводятся как [].
type Result struct {
	URL       string       `json:"url"`
	Device    string       `json:"device"`
	Status    int          `json:"status"`
	Rendered  bool         `json:"rendered"`
	BlockedBy string       `json:"blocked_by,omitempty"` // получена страница блокировки защиты от ботов, а не сам сайт
	SEO       *SEOInfo     `json:"seo,omitzero"`
	OpenGraph []page.Meta  `json:"open_graph,omitzero"`
	Twitter   []page.Meta  `json:"twitter,omitzero"`
	OtherMeta []page.Meta  `json:"other_meta,omitzero"`
	Trackers  []detect.Hit `json:"trackers,omitzero"`
	TechMode  string       `json:"tech_mode,omitempty"`
	Tech      []detect.Hit `json:"technologies,omitzero"`
	Languages []string     `json:"languages,omitzero"`
}

// Build собирает отчёт. jsTech — технологии, найденные по JS-объектам в Chrome (только для TechChrome).
func Build(device string, resp fetch.Response, jsTech []detect.Hit, opts Options) (Result, error) {
	p, err := page.Parse(resp.Body)
	if err != nil {
		return Result{}, err
	}
	r := Result{URL: resp.URL, Device: device, Status: resp.Status, Rendered: opts.Render, BlockedBy: resp.BotProtection()}
	want := func(s string) bool { return slices.Contains(opts.Sections, s) }

	if want(SEO) {
		r.SEO = &SEOInfo{
			Title:       p.Title,
			Description: p.Meta("description"),
			Keywords:    p.Meta("keywords"),
			Robots:      p.Meta("robots"),
			Canonical:   p.Canonical,
			Lang:        p.Lang,
			Viewport:    p.Meta("viewport"),
			H1:          append([]string{}, p.H1...),
			Images:      p.Images,
			ImagesNoAlt: p.ImagesNoAlt,
			Hreflang:    append([]page.Meta{}, p.Hreflang...),
		}
	}
	if want(OpenGraph) {
		r.OpenGraph = filterMetas(p.Metas, isOG)
	}
	if want(Twitter) {
		r.Twitter = filterMetas(p.Metas, isTwitter)
	}
	if want(OtherMeta) {
		r.OtherMeta = filterMetas(p.Metas, isOther)
	}
	if want(Trackers) {
		r.Trackers = detect.Trackers(resp.Body, p.Metas)
	}
	if want(Tech) {
		r.TechMode = opts.TechMode
		switch opts.TechMode {
		case TechDeep, TechChrome:
			r.Tech, r.Languages, err = detect.TechDeep(resp.Body, resp.Header, p.Metas)
			if err != nil {
				return Result{}, err
			}
			// JS-находки ставим первыми: они точнее и чаще содержат версию.
			r.Tech = detect.Merge(append([]detect.Hit{}, jsTech...), r.Tech)
		default:
			r.Tech, r.Languages = detect.Tech(resp.Body, resp.Header, p.Metas)
		}
	}
	return r, nil
}

func isOG(k string) bool      { return strings.HasPrefix(k, "og:") }
func isTwitter(k string) bool { return strings.HasPrefix(k, "twitter:") }

// isOther — всё, что не попало в SEO, OG, Twitter и верификации.
func isOther(k string) bool {
	_, verification := detect.Verifications[k]
	seoKeys := []string{"description", "keywords", "robots", "viewport", "generator"}
	return !verification && !isOG(k) && !isTwitter(k) && !slices.Contains(seoKeys, k)
}

func filterMetas(metas []page.Meta, keep func(string) bool) []page.Meta {
	out := []page.Meta{}
	for _, m := range metas {
		if keep(m.Key) {
			out = append(out, m)
		}
	}
	return out
}
