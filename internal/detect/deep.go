package detect

import (
	"net/http"
	"slices"
	"strings"
	"sync"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"

	"seo-analytic/internal/page"
)

// Разбор вшитой базы правил занимает заметное время — делаем его один раз и только если нужен углублённый анализ.
var wappalyzerClient = sync.OnceValues(wappalyzer.New)

// TechDeep — углублённый анализ: база правил Wappalyzer (тысячи технологий, с версиями и категориями)
// плюс находки простого анализа, которых в ней нет. Часть правил Wappalyzer проверяет JS-объекты
// в браузере — без рендеринга они не срабатывают (так теряется, например, Next.js), простой анализ это закрывает.
// Языки берутся из категории «Programming languages» и из простого анализа.
func TechDeep(body string, header http.Header, metas []page.Meta) ([]Hit, []string, error) {
	w, err := wappalyzerClient()
	if err != nil {
		return nil, nil, err
	}
	hits, langs := []Hit{}, []string{}
	for key, info := range w.FingerprintWithInfo(header, []byte(body)) {
		// Ключ имеет вид "имя" или "имя:версия".
		name, version, _ := strings.Cut(key, ":")
		h := Hit{Name: name, Category: strings.Join(info.Categories, ", ")}
		if version != "" {
			h.Details = []string{version}
		}
		hits = append(hits, h)
		if slices.Contains(info.Categories, "Programming languages") {
			langs = append(langs, name)
		}
	}

	simpleHits, simpleLangs := Tech(body, header, metas)
	hits = Merge(hits, simpleHits)
	for _, l := range simpleLangs {
		if !slices.Contains(langs, l) {
			langs = append(langs, l)
		}
	}
	// Результат приходит map'ой — сортируем, чтобы вывод был стабильным.
	slices.SortFunc(hits, func(a, b Hit) int {
		if c := strings.Compare(a.Category, b.Category); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
	slices.Sort(langs)
	return hits, langs, nil
}
