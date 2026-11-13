// Package fetch скачивает страницу от имени десктопного или мобильного браузера.
package fetch

import (
	"io"
	"net/http"
	"time"

	"golang.org/x/net/html/charset"
)

const (
	Desktop = "desktop"
	Mobile  = "mobile"
)

// UserAgents — User-Agent для каждой версии сайта; используется и в HTTP-запросе, и в Chrome.
var UserAgents = map[string]string{
	Desktop: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36",
	Mobile:  "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1",
}

var client = &http.Client{Timeout: 20 * time.Second}

type Response struct {
	URL    string // итоговый, после редиректов
	Status int
	Header http.Header
	Body   string // HTML в UTF-8
}

// Get запрашивает страницу; тело перекодируется в UTF-8 (актуально для сайтов на windows-1251).
func Get(url, device string) (Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Response{}, err
	}
	// Некоторые сайты отдают заглушку не-браузерным User-Agent.
	req.Header.Set("User-Agent", UserAgents[device])
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	r, err := charset.NewReader(io.LimitReader(resp.Body, 10<<20), resp.Header.Get("Content-Type"))
	if err != nil {
		return Response{}, err
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return Response{}, err
	}
	return Response{
		URL:    resp.Request.URL.String(),
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   string(body),
	}, nil
}
