package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"github.com/charmbracelet/x/term"

	"seo-analytic/internal/fetch"
	"seo-analytic/internal/report"
)

// askRender спрашивает, повторить ли запрос через Chrome, когда обычный заблокирован защитой от ботов.
// Без терминала спросить некого — возвращает false, решение остаётся за флагом -render.
func askRender(protection string) (bool, error) {
	if !term.IsTerminal(os.Stdin.Fd()) {
		return false, nil
	}
	yes := true
	err := huh.NewConfirm().
		Title(fmt.Sprintf("Сайт защищён от ботов (%s)", protection)).
		Description("Обычный запрос заблокирован. Получить данные через рендеринг в Chrome?\nЭто медленнее (~10 с) и нужен установленный Chrome/Chromium/Edge.").
		Affirmative("Да").
		Negative("Нет").
		Value(&yes).
		Run()
	return yes, err
}

// ask спрашивает домен, разделы отчёта, режим анализа технологий и версию сайта.
// Форма рисуется в stderr, поэтому `seo-analytic -json > report.json` работает и в интерактивном режиме.
func ask() (target string, opts report.Options, devices []string, err error) {
	opts = report.Options{
		Sections: append([]string{}, report.AllSections...), // по умолчанию отмечено всё
		TechMode: report.TechSimple,
	}
	sectionOpts := make([]huh.Option[string], 0, len(report.AllSections))
	for _, s := range report.AllSections {
		sectionOpts = append(sectionOpts, huh.NewOption(report.SectionNames[s], s))
	}
	device := fetch.Desktop

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Домен или URL").
				Placeholder("example.com").
				Value(&target).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("введите домен")
					}
					return nil
				}),
			huh.NewMultiSelect[string]().
				Title("Что показать").
				Description("пробел — отметить, ctrl+a — все/ничего").
				Options(sectionOpts...).
				Height(len(sectionOpts)+2). // +2 строки на заголовок и описание, иначе часть пунктов уходит под прокрутку
				Value(&opts.Sections).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errors.New("отметьте хотя бы один раздел")
					}
					return nil
				}),
		),
		// Вопрос про глубину анализа имеет смысл, только если раздел технологий выбран.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Анализ технологий").
				Options(
					huh.NewOption("Простой — свои сигнатуры, быстро", report.TechSimple),
					huh.NewOption("Углублённый — база Wappalyzer, тысячи технологий", report.TechDeep),
					huh.NewOption("Экспериментальный — + JS-объекты в Chrome (нужен Chrome)", report.TechChrome),
				).
				Value(&opts.TechMode),
		).WithHideFunc(func() bool { return !slices.Contains(opts.Sections, report.Tech) }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Версия сайта").
				Options(
					huh.NewOption("Десктоп", fetch.Desktop),
					huh.NewOption("Мобильная", fetch.Mobile),
					huh.NewOption("Обе", "both"),
				).
				Value(&device),
		),
		// Экспериментальный режим и так рендерит в Chrome — не спрашиваем.
		huh.NewGroup(
			huh.NewConfirm().
				Title("Рендерить страницу в Chrome?").
				Description("увидим DOM после JS и трекеры из GTM; медленнее, нужен Chrome/Chromium/Edge").
				Affirmative("Да").
				Negative("Нет").
				Value(&opts.Render),
		).WithHideFunc(func() bool {
			return opts.TechMode == report.TechChrome && slices.Contains(opts.Sections, report.Tech)
		}),
	).Run()

	devices = []string{device}
	if device == "both" {
		devices = []string{fetch.Desktop, fetch.Mobile}
	}
	return strings.TrimSpace(target), opts, devices, err
}
