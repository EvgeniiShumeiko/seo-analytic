package fetch

import (
	"net/http"
	"testing"
)

func TestBotProtection(t *testing.T) {
	h := func(kv ...string) http.Header {
		hd := http.Header{}
		for i := 0; i < len(kv); i += 2 {
			hd.Set(kv[i], kv[i+1])
		}
		return hd
	}
	cases := []struct {
		name string
		resp Response
		want string
	}{
		{"servicepipe", Response{Status: 403, Header: h("X-Sp-Crid", "1:1")}, "ServicePipe"},
		{"servicepipe header on normal page", Response{Status: 200, Header: h("X-Sp-Crid", "1:1")}, ""},
		{"cloudflare challenge", Response{Status: 403, Header: h("Cf-Mitigated", "challenge")}, "Cloudflare"},
		{"ddos-guard", Response{Status: 503, Header: h("Server", "ddos-guard")}, "DDoS-Guard"},
		{"captcha", Response{Status: 429, Header: h(), Body: "<div class=g-recaptcha>"}, "капча"},
		{"plain 403", Response{Status: 403, Header: h("Server", "nginx"), Body: "Forbidden"}, ""},
	}
	for _, c := range cases {
		if got := c.resp.BotProtection(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
