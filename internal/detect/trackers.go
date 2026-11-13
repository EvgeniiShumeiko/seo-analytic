package detect

import (
	"regexp"

	"seo-analytic/internal/page"
)

// Регулярки с группой захвата ставим первыми — они дают ID.
var trackers = []struct {
	name string
	res  []*regexp.Regexp
}{
	{"Google Analytics 4", rx(`gtag\(\s*['"]config['"]\s*,\s*['"](G-[A-Z0-9]+)`, `googletagmanager\.com/gtag/js\?id=(G-[A-Z0-9]+)`)},
	{"Google Universal Analytics", rx(`['"](UA-\d+-\d+)['"]`, `google-analytics\.com/(?:analytics|ga)\.js`)},
	{"Google Tag Manager", rx(`['"=](GTM-[A-Z0-9]+)`, `googletagmanager\.com/gtm\.js`)},
	{"Google Ads", rx(`['"=](AW-\d+)`, `googleadservices\.com`)},
	{"Яндекс.Метрика", rx(`ym\(\s*(\d+)\s*,\s*['"]init`, `mc\.yandex\.ru/watch/(\d+)`, `yaCounter(\d+)`, `mc\.yandex\.ru/metrika`)},
	{"Facebook Pixel", rx(`fbq\(\s*['"]init['"]\s*,\s*['"](\d+)`, `facebook\.com/tr\?id=(\d+)`, `connect\.facebook\.net/[^"']*/fbevents\.js`)},
	{"VK Pixel", rx(`VK\.Retargeting\.Init\(\s*['"]([\w-]+)`, `vk\.com/rtrg\?p=([\w-]+)`, `vk\.com/js/api/openapi\.js`)},
	{"Top.Mail.ru / VK Ads", rx(`_tmr\.push\(\s*\{[^}]*id:\s*['"](\d+)`, `top-fwz1\.mail\.ru/counter\?id=(\d+)`, `top-fwz1\.mail\.ru/js/code\.js`)},
	{"TikTok Pixel", rx(`ttq\.load\(\s*['"](\w+)`, `analytics\.tiktok\.com`)},
	{"Microsoft Clarity", rx(`["']clarity["']\s*,\s*["']script["']\s*,\s*["'](\w+)`, `clarity\.ms/tag/`)},
	{"Bing UET", rx(`bat\.bing\.com`)},
	{"Hotjar", rx(`hjid\s*:\s*(\d+)`, `static\.hotjar\.com`)},
	{"LinkedIn Insight", rx(`_linkedin_partner_id\s*=\s*['"]?(\d+)`, `snap\.licdn\.com`)},
	{"Pinterest Tag", rx(`pintrk\(\s*['"]load['"]\s*,\s*['"](\d+)`, `s\.pinimg\.com/ct/core\.js`)},
}

// Verifications — meta-теги подтверждения прав на сайт: имя тега → сервис.
var Verifications = map[string]string{
	"google-site-verification":     "Google Search Console",
	"yandex-verification":          "Яндекс.Вебмастер",
	"msvalidate.01":                "Bing Webmaster",
	"wmail-verification":           "Mail.ru Webmaster",
	"facebook-domain-verification": "Facebook Domain Verification",
	"p:domain_verify":              "Pinterest Verification",
}

// Trackers ищет счётчики в исходном HTML и верификации в meta-тегах.
// shortcut: трекеры, подгружаемые через GTM или JS после рендера, не видны — для них нужен headless-браузер.
func Trackers(body string, metas []page.Meta) []Hit {
	hits := []Hit{}
	for _, t := range trackers {
		if found, ids := matches(body, t.res); found {
			hits = append(hits, Hit{Name: t.name, Details: ids})
		}
	}
	for _, m := range metas {
		if name, ok := Verifications[m.Key]; ok {
			hits = append(hits, Hit{Name: name, Details: []string{m.Content}})
		}
	}
	return hits
}
