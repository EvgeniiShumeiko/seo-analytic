package page

import "testing"

func TestParse(t *testing.T) {
	p, err := Parse(`<html lang="ru"><head><title> Тест </title>
<meta name="Description" content="Описание"><meta property="og:title" content="OG">
<link rel="canonical" href="https://example.com/"><link rel="alternate" hreflang="en" href="/en"></head>
<body><h1>Заголовок</h1><svg><title>иконка</title></svg><img src="a.png"><img src="b.png" alt="b"></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Тест" || p.Lang != "ru" || p.Canonical != "https://example.com/" {
		t.Errorf("bad page: %+v", p)
	}
	if p.Meta("description") != "Описание" || p.Meta("og:title") != "OG" || len(p.Hreflang) != 1 {
		t.Errorf("bad meta: %+v", p)
	}
	if p.Images != 2 || p.ImagesNoAlt != 1 || len(p.H1) != 1 {
		t.Errorf("bad img/h1: %+v", p)
	}
}
