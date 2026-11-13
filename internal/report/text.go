package report

import (
	"fmt"
	"io"
	"strings"

	"seo-analytic/internal/detect"
	"seo-analytic/internal/page"
)

// Print выводит отчёт в человекочитаемом виде с предупреждениями (⚠).
func Print(w io.Writer, r Result) {
	how := ""
	if r.Rendered {
		how = ", отрендерено в Chrome"
	}
	fmt.Fprintf(w, "URL: %s (HTTP %d, %s%s)\n", r.URL, r.Status, DeviceNames[r.Device], how)
	if r.BlockedBy != "" {
		fmt.Fprintf(w, "⚠ Запрос заблокирован защитой от ботов (%s): ниже данные страницы блокировки, а не сайта.\n", r.BlockedBy)
		if !r.Rendered {
			fmt.Fprintln(w, "  Получить данные сайта можно через рендеринг в Chrome: добавьте флаг -render.")
		}
	}

	if s := r.SEO; s != nil {
		printSEO(w, s)
	}
	if r.OpenGraph != nil {
		printMetas(w, OpenGraph, r.OpenGraph)
	}
	if r.Twitter != nil {
		printMetas(w, Twitter, r.Twitter)
	}
	if r.OtherMeta != nil {
		printMetas(w, OtherMeta, r.OtherMeta)
	}
	if r.Trackers != nil {
		printHeader(w, Trackers)
		printHits(w, r.Trackers)
	}
	if r.Tech != nil {
		modes := map[string]string{
			TechSimple: "простой анализ",
			TechDeep:   "углублённый анализ",
			TechChrome: "экспериментальный анализ в Chrome",
		}
		mode := modes[r.TechMode]
		fmt.Fprintf(w, "\n== %s (%s) ==\n", SectionNames[Tech], mode)
		printHits(w, r.Tech)
		if len(r.Languages) > 0 {
			fmt.Fprintf(w, "  Языки (вероятно): %s\n", strings.Join(r.Languages, ", "))
		}
	}
}

func printSEO(w io.Writer, s *SEOInfo) {
	printHeader(w, SEO)
	printLine(w, "Title", s.Title, lenWarn(s.Title, 10, 60))
	printLine(w, "Description", s.Description, lenWarn(s.Description, 50, 160))
	printLine(w, "Keywords", s.Keywords, "")

	warn := ""
	if strings.Contains(strings.ToLower(s.Robots), "noindex") {
		warn = "страница закрыта от индексации"
	}
	printLine(w, "Robots", s.Robots, warn)
	printLine(w, "Canonical", s.Canonical, "")

	warn = ""
	if s.Lang == "" {
		warn = "не указан"
	}
	printLine(w, "Lang", s.Lang, warn)

	warn = ""
	if s.Viewport == "" {
		warn = "нет — плохо для мобильных"
	}
	printLine(w, "Viewport", s.Viewport, warn)

	warn = ""
	if len(s.H1) != 1 {
		warn = fmt.Sprintf("найдено %d, рекомендуется ровно один", len(s.H1))
	}
	printLine(w, "H1", strings.Join(s.H1, " | "), warn)

	warn = ""
	if s.ImagesNoAlt > 0 {
		warn = fmt.Sprintf("%d без alt", s.ImagesNoAlt)
	}
	printLine(w, "Изображения", fmt.Sprint(s.Images), warn)

	for _, h := range s.Hreflang {
		printLine(w, "hreflang "+h.Key, h.Content, "")
	}
}

func printHeader(w io.Writer, section string) {
	fmt.Fprintf(w, "\n== %s ==\n", SectionNames[section])
}

func printLine(w io.Writer, label, val, warn string) {
	if val == "" {
		val = "—"
	}
	if warn != "" {
		warn = "  ⚠ " + warn
	}
	fmt.Fprintf(w, "  %-14s %s%s\n", label+":", val, warn)
}

func printMetas(w io.Writer, section string, metas []page.Meta) {
	printHeader(w, section)
	if len(metas) == 0 {
		fmt.Fprintln(w, "  —")
	}
	for _, m := range metas {
		printLine(w, m.Key, m.Content, "")
	}
}

func printHits(w io.Writer, hits []detect.Hit) {
	if len(hits) == 0 {
		fmt.Fprintln(w, "  ничего не найдено")
	}
	for _, h := range hits {
		fmt.Fprintf(w, "  ✓ %s", h.Name)
		if h.Category != "" {
			fmt.Fprintf(w, " (%s)", h.Category)
		}
		if len(h.Details) > 0 {
			fmt.Fprintf(w, "  [%s]", strings.Join(h.Details, ", "))
		}
		fmt.Fprintln(w)
	}
}

func lenWarn(s string, lo, hi int) string {
	n := len([]rune(s))
	if n == 0 {
		return "отсутствует"
	}
	if n < lo || n > hi {
		return fmt.Sprintf("%d симв., рекомендуется %d–%d", n, lo, hi)
	}
	return ""
}
