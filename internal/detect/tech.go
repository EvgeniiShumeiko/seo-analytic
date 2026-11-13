package detect

import (
	"net/http"
	"regexp"
	"slices"
	"strings"

	"seo-analytic/internal/page"
)

// Сигнатуры простого анализа. header-регулярки применяются к заголовкам ответа,
// приведённым к виду "ключ: значение" в нижнем регистре, по строке на заголовок.
// lang — язык бэкенда, который технология выдаёт; пусто, если по ней язык не понять.
var techs = []struct {
	name, category, lang string
	header, html         []*regexp.Regexp
}{
	{"WordPress", "CMS", "PHP", nil, rx(`/wp-(?:content|includes)/`)},
	{"1С-Битрикс", "CMS", "PHP", rx(`(?m)^x-powered-cms: bitrix`), rx(`/bitrix/(?:js|templates|cache)/`)},
	{"Joomla", "CMS", "PHP", nil, rx(`/media/jui/`, `/components/com_`)},
	{"Drupal", "CMS", "PHP", rx(`(?m)^x-generator: drupal (\d+)`), rx(`Drupal\.settings`, `/sites/(?:default|all)/files/`)},
	{"OpenCart", "CMS", "PHP", nil, rx(`catalog/view/(?:theme|javascript)/`)},
	{"Tilda", "Конструктор", "", nil, rx(`tildacdn\.com`)},
	{"Wix", "Конструктор", "", nil, rx(`static\.wixstatic\.com`)},
	{"Webflow", "Конструктор", "", nil, rx(`data-wf-page`)},
	{"Shopify", "E-commerce", "Ruby", nil, rx(`cdn\.shopify\.com`)},

	{"Next.js", "Фронтенд", "Node.js", rx(`(?m)^x-powered-by: next\.js ?([\d.]+)?`), rx(`__NEXT_DATA__`, `/_next/static/`)},
	{"Nuxt", "Фронтенд", "Node.js", nil, rx(`__NUXT__`, `/_nuxt/`)},
	{"SvelteKit", "Фронтенд", "Node.js", nil, rx(`__sveltekit`, `data-sveltekit`)},
	{"Remix", "Фронтенд", "Node.js", nil, rx(`__remixContext`)},
	{"Gatsby", "Фронтенд", "", nil, rx(`___gatsby`)},
	{"Astro", "Фронтенд", "", nil, rx(`<astro-island`)},
	{"React", "Фронтенд", "", nil, rx(`data-reactroot`, `react-dom(?:\.production)?(?:\.min)?\.js`)},
	{"Vue.js", "Фронтенд", "", nil, rx(`data-v-[0-9a-f]{8}`, `vue(?:\.runtime)?(?:\.global)?(?:\.prod)?(?:\.min)?\.js`)},
	{"Angular", "Фронтенд", "", nil, rx(`ng-version="([\d.]+)"`)},
	{"jQuery", "JS-библиотека", "", nil, rx(`jquery[-.](\d+\.\d+(?:\.\d+)?)(?:\.min)?\.js`, `jquery(?:\.min)?\.js`)},
	{"Bootstrap", "CSS-фреймворк", "", nil, rx(`bootstrap(?:\.bundle)?(?:\.min)?\.(?:css|js)`)},

	{"PHP", "Бэкенд", "PHP", rx(`(?m)^x-powered-by: php/?([\d.]+)?`, `(?m)^set-cookie: phpsessid=`), nil},
	{"Laravel", "Бэкенд", "PHP", rx(`(?m)^set-cookie: laravel_session=`), nil},
	{"Django", "Бэкенд", "Python", rx(`(?m)^set-cookie: csrftoken=`), rx(`csrfmiddlewaretoken`)},
	{"Ruby on Rails", "Бэкенд", "Ruby", nil, rx(`content="authenticity_token"`)},
	{"ASP.NET", "Бэкенд", "C#", rx(`(?m)^x-aspnet-version: ([\d.]+)`, `(?m)^x-powered-by: asp\.net`, `(?m)^set-cookie: asp\.net_sessionid=`), rx(`__VIEWSTATE`)},
	{"Java (сервлеты)", "Бэкенд", "Java", rx(`(?m)^set-cookie: jsessionid=`), nil},
	{"Express", "Бэкенд", "Node.js", rx(`(?m)^x-powered-by: express`), nil},

	{"Cloudflare", "CDN/хостинг", "", rx(`(?m)^cf-ray:`), nil},
	{"Vercel", "CDN/хостинг", "", rx(`(?m)^x-vercel-id:`), nil},
	{"Netlify", "CDN/хостинг", "", rx(`(?m)^x-nf-request-id:`), nil},
}

// Tech — простой анализ: ~30 своих сигнатур. Возвращает технологии и языки бэкенда, которые из них следуют.
func Tech(body string, header http.Header, metas []page.Meta) ([]Hit, []string) {
	var hb strings.Builder
	for k, vs := range header {
		for _, v := range vs {
			hb.WriteString(strings.ToLower(k + ": " + v + "\n"))
		}
	}
	headers := hb.String()

	hits, langs := []Hit{}, []string{}
	for _, m := range metas {
		if m.Key == "generator" && m.Content != "" {
			hits = append(hits, Hit{Name: m.Content, Category: "Генератор"})
		}
	}
	for _, t := range techs {
		inHeader, v1 := matches(headers, t.header)
		inHTML, v2 := matches(body, t.html)
		if !inHeader && !inHTML {
			continue
		}
		hits = append(hits, Hit{Name: t.name, Category: t.category, Details: append(v1, v2...)})
		if t.lang != "" && !slices.Contains(langs, t.lang) {
			langs = append(langs, t.lang)
		}
	}
	// Заголовок Server добавляем, только если он не повторяет уже найденное (Server: Vercel и сигнатура Vercel).
	if s := header.Get("Server"); s != "" && !slices.ContainsFunc(hits, func(h Hit) bool { return strings.EqualFold(h.Name, s) }) {
		hits = append(hits, Hit{Name: s, Category: "Сервер"})
	}
	return hits, langs
}
