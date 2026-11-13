package render

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"seo-analytic/internal/fetch"
)

// Тест поднимает страницу, где фреймворки появляются только после выполнения JS,
// и проверяет, что рендер видит и DOM, и JS-объекты. Без Chrome — пропускается.
func TestPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Два Set-Cookie: Chrome склеит их, а мы должны разложить обратно.
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "PHPSESSID=x")
		w.Header().Set("X-Powered-By", "PHP/8.2")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`<html><head><title>t</title></head><body><div id="app"></div><script>
window.jQuery = {fn: {jquery: "3.7.1"}};
const app = document.getElementById("app");
app.__vue_app__ = {version: "3.4.0"};
app.innerHTML = '<h1>Отрендерено</h1>';
</script></body></html>`))
	}))
	defer srv.Close()

	r, err := Page(srv.URL, fetch.Desktop)
	var notFound *exec.Error
	if errors.As(err, &notFound) {
		t.Skip("Chrome не установлен:", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	resp := r.Response
	if !strings.Contains(resp.Body, "<h1>Отрендерено</h1>") {
		t.Errorf("DOM не отрендерен: %s", resp.Body)
	}
	if resp.Status != http.StatusTeapot || resp.Header.Get("X-Powered-By") != "PHP/8.2" {
		t.Errorf("ответ не из Chrome: %d %v", resp.Status, resp.Header)
	}
	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) != 2 {
		t.Errorf("Set-Cookie не разложен: %q", cookies)
	}
	got := map[string][]string{}
	for _, h := range r.Tech {
		got[h.Name] = h.Details
	}
	if !slices.Contains(got["jQuery"], "3.7.1") || !slices.Contains(got["Vue.js"], "3.4.0") {
		t.Errorf("JS-проверки: %v", got)
	}
}
