// seo-analytic: печатает мета-теги, SEO-сводку, Open Graph, трекеры и технологии страницы.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"

	"seo-analytic/internal/detect"
	"seo-analytic/internal/fetch"
	"seo-analytic/internal/render"
	"seo-analytic/internal/report"
)

func main() {
	// Консоль настраиваем до любого вывода, а закрываем в одной точке: os.Exit пропускает defer.
	closeConsole := setupConsole()
	code := run()
	closeConsole(code)
	os.Exit(code)
}

// run выполняет программу и возвращает код завершения: 0 — успех, 1 — ошибка, 2 — неверные аргументы, 130 — прервано.
func run() int {
	mobile := flag.Bool("mobile", false, "запрашивать страницу как мобильный браузер (iPhone Safari)")
	asJSON := flag.Bool("json", false, "вывести отчёт в JSON")
	techMode := flag.String("tech", report.TechSimple,
		"анализ технологий: simple — свои сигнатуры, deep — база Wappalyzer, chrome — экспериментальный: deep + JS-объекты в Chrome (включает -render)")
	renderJS := flag.Bool("render", false, "рендерить страницу в headless Chrome: виден DOM после JS и трекеры из GTM (нужен Chrome/Chromium/Edge)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: seo-analytic [-mobile] [-json] [-render] [-tech simple|deep|chrome] [url]\nбез url — интерактивный режим")
		flag.PrintDefaults()
	}
	flag.Parse()

	switch *techMode {
	case report.TechSimple, report.TechDeep, report.TechChrome:
	default:
		fmt.Fprintf(os.Stderr, "неизвестный режим -tech %q, допустимо: simple, deep, chrome\n", *techMode)
		return 2
	}

	var target string
	opts := report.Options{Sections: report.AllSections, TechMode: *techMode, Render: *renderJS}
	devices := []string{fetch.Desktop}
	switch flag.NArg() {
	case 0:
		var err error
		target, opts, devices, err = ask()
		if errors.Is(err, huh.ErrUserAborted) {
			return 130
		}
		if err != nil {
			return fail(err)
		}
	case 1:
		target = flag.Arg(0)
		if *mobile {
			devices = []string{fetch.Mobile}
		}
	default:
		flag.Usage()
		return 2
	}
	if !strings.Contains(target, "://") {
		target = "https://" + target
	}
	// JS-объекты существуют только в браузере.
	if opts.TechMode == report.TechChrome && slices.Contains(opts.Sections, report.Tech) {
		opts.Render = true
	}

	var results []report.Result
	asked := false // про Chrome спрашиваем один раз, даже если проверяем обе версии сайта
	for _, device := range devices {
		resp, jsTech, err := load(target, device, opts.Render)
		if err != nil {
			return fail(fmt.Errorf("%s: %w", report.DeviceNames[device], err))
		}

		if protection := resp.BotProtection(); protection != "" && !opts.Render && !asked {
			asked = true
			yes, err := askRender(protection)
			if errors.Is(err, huh.ErrUserAborted) {
				return 130
			}
			if err != nil {
				return fail(err)
			}
			if yes {
				opts.Render = true
				if resp, jsTech, err = load(target, device, true); err != nil {
					return fail(fmt.Errorf("%s: %w", report.DeviceNames[device], err))
				}
			} else if *asJSON {
				// В текстовом отчёте это предупреждение и так есть, а JSON читает программа — подсказываем человеку в stderr.
				fmt.Fprintf(os.Stderr, "⚠ сайт защищён от ботов (%s): обычный запрос заблокирован; данные сайта можно получить с флагом -render\n", protection)
			}
		}

		r, err := report.Build(device, resp, jsTech, opts)
		if err != nil {
			return fail(err)
		}
		results = append(results, r)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(results); err != nil {
			return fail(err)
		}
		return 0
	}
	for i, r := range results {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("─", 60))
		}
		report.Print(os.Stdout, r)
	}
	return 0
}

// load скачивает страницу обычным HTTP-запросом или рендерит в Chrome.
func load(target, device string, useChrome bool) (fetch.Response, []detect.Hit, error) {
	if useChrome {
		r, err := render.Page(target, device)
		return r.Response, r.Tech, err
	}
	resp, err := fetch.Get(target, device)
	return resp, nil, err
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, "error:", err)
	return 1
}
