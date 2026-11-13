// Package render открывает страницу в headless Chrome: отдаёт DOM после выполнения JS
// и технологии, найденные по JS-объектам страницы.
package render

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"seo-analytic/internal/detect"
	"seo-analytic/internal/fetch"
)

//go:embed probes.js
var probesJS string

type Result struct {
	Response fetch.Response // статус и заголовки ответа, который получил Chrome; Body — DOM после рендеринга
	Tech     []detect.Hit   // технологии по JS-объектам (window.React, __vue_app__ и т.п.)
}

type probe struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Version  string `json:"version"`
}

// Page рендерит страницу. Chrome, Chromium или Edge ищутся в стандартных путях текущей ОС.
func Page(url, device string) (Result, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.UserAgent(fetch.UserAgents[device]))
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	viewport := chromedp.EmulateViewport(1440, 900)
	if device == fetch.Mobile {
		viewport = chromedp.EmulateViewport(390, 844, chromedp.EmulateScale(3), chromedp.EmulateMobile, chromedp.EmulateTouch, chromedp.EmulatePortrait)
	}
	// Первый вызов запускает браузер — таймаут на него не ставим, иначе он остановит весь браузер.
	if err := chromedp.Do(ctx, viewport); err != nil {
		var notFound *exec.Error
		if errors.As(err, &notFound) {
			return Result{}, fmt.Errorf("не найден Chrome/Chromium/Edge — установите браузер для -render: %w", err)
		}
		return Result{}, fmt.Errorf("запуск Chrome: %w", err)
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, 45*time.Second)
	defer cancelTimeout()
	// Статус и заголовки берём из того же ответа, что и DOM: отдельный HTTP-запрос из Go
	// может получить другой ответ (например, страницу блокировки от защиты от ботов).
	doc, err := chromedp.Run(ctx, chromedp.NavigateResponse(url))
	if err != nil {
		return Result{}, fmt.Errorf("загрузка в Chrome: %w", err)
	}
	// shortcut: фиксированная пауза после load на отложенные скрипты (GTM, ленивые счётчики); на медленных сайтах может не хватить.
	if err := chromedp.Do(ctx, chromedp.Sleep(2*time.Second)); err != nil {
		return Result{}, err
	}
	html, err := chromedp.Run(ctx, chromedp.OuterHTML(chromedp.CSS("html")))
	if err != nil {
		return Result{}, err
	}
	probes, err := chromedp.Run(ctx, chromedp.Evaluate[[]probe](probesJS))
	if err != nil {
		return Result{}, fmt.Errorf("JS-проверки: %w", err)
	}

	hits := make([]detect.Hit, 0, len(probes))
	for _, p := range probes {
		h := detect.Hit{Name: p.Name, Category: p.Category}
		if p.Version != "" {
			h.Details = []string{p.Version}
		}
		hits = append(hits, h)
	}

	// Chrome склеивает повторяющиеся заголовки через "\n" — раскладываем обратно.
	header := http.Header{}
	for k, v := range doc.Headers {
		for _, line := range strings.Split(fmt.Sprint(v), "\n") {
			header.Add(k, line)
		}
	}
	// Set-Cookie Chrome в заголовках ответа не отдаёт, поэтому берём cookie, установленные на странице:
	// для сигнатур (PHPSESSID, laravel_session…) важны только имена.
	cookies, err := chromedp.Call(ctx, network.GetCookies, network.GetCookiesParams{})
	if err != nil {
		return Result{}, fmt.Errorf("cookie: %w", err)
	}
	for _, c := range cookies.Cookies {
		header.Add("Set-Cookie", c.Name+"="+c.Value)
	}
	resp := fetch.Response{URL: doc.URL, Status: int(doc.Status), Header: header, Body: html}
	return Result{Response: resp, Tech: hits}, nil
}
