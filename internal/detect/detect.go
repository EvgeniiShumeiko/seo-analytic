// Package detect ищет на странице счётчики, пиксели, верификации и технологии.
package detect

import (
	"regexp"
	"slices"
	"strings"
)

type Hit struct {
	Name     string   `json:"name"`
	Category string   `json:"category,omitempty"`
	Details  []string `json:"details,omitempty"` // ID счётчика, код верификации или версия
}

// Merge дополняет hits находками из extra, которых ещё нет (имена сравниваются без учёта регистра).
func Merge(hits, extra []Hit) []Hit {
	for _, h := range extra {
		if !slices.ContainsFunc(hits, func(d Hit) bool { return strings.EqualFold(d.Name, h.Name) }) {
			hits = append(hits, h)
		}
	}
	return hits
}

func rx(ps ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(ps))
	for i, p := range ps {
		out[i] = regexp.MustCompile(p)
	}
	return out
}

// matches сообщает, сработала ли хоть одна регулярка, и собирает уникальные значения первой группы захвата.
func matches(s string, res []*regexp.Regexp) (bool, []string) {
	found := false
	var vals []string
	for _, re := range res {
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			found = true
			if len(m) > 1 && m[1] != "" && !slices.Contains(vals, m[1]) {
				vals = append(vals, m[1])
			}
		}
	}
	return found, vals
}
