// Package page извлекает из HTML мета-теги и SEO-данные.
package page

import (
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type Meta struct {
	Key     string `json:"key"`
	Content string `json:"content"`
}

type Page struct {
	Lang, Title, Canonical string
	Metas                  []Meta // в порядке документа
	Hreflang               []Meta
	H1                     []string
	Images, ImagesNoAlt    int
}

func (p *Page) Meta(key string) string {
	for _, m := range p.Metas {
		if m.Key == key {
			return m.Content
		}
	}
	return ""
}

func Parse(body string) (Page, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return Page{}, err
	}
	var p Page
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}
		switch n.Data {
		case "html":
			p.Lang, _ = attr(n, "lang")
		case "title":
			// Берём первый title: внутри <svg> тоже бывают свои.
			if p.Title == "" {
				p.Title = text(n)
			}
		case "meta":
			content, _ := attr(n, "content")
			for _, k := range []string{"name", "property", "http-equiv", "itemprop"} {
				if v, ok := attr(n, k); ok {
					p.Metas = append(p.Metas, Meta{strings.ToLower(v), content})
					break
				}
			}
			if v, ok := attr(n, "charset"); ok {
				p.Metas = append(p.Metas, Meta{"charset", v})
			}
		case "link":
			rel, _ := attr(n, "rel")
			href, _ := attr(n, "href")
			rels := strings.Fields(strings.ToLower(rel))
			if slices.Contains(rels, "canonical") {
				p.Canonical = href
			}
			if hl, ok := attr(n, "hreflang"); ok && slices.Contains(rels, "alternate") {
				p.Hreflang = append(p.Hreflang, Meta{hl, href})
			}
		case "h1":
			p.H1 = append(p.H1, text(n))
		case "img":
			p.Images++
			if alt, ok := attr(n, "alt"); !ok || strings.TrimSpace(alt) == "" {
				p.ImagesNoAlt++
			}
		}
	}
	return p, nil
}

func attr(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

func text(n *html.Node) string {
	var b strings.Builder
	for d := range n.Descendants() {
		if d.Type == html.TextNode {
			b.WriteString(d.Data + " ")
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
