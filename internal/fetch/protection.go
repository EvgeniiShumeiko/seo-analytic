package fetch

import (
	"net/http"
	"strings"
)

// BotProtection возвращает название защиты от ботов, если ответ похож на её страницу блокировки, иначе "".
// Заголовки защиты бывают и в обычных ответах, поэтому смотрим их только вместе с кодом блокировки.
// shortcut: только известные сервисы и явная капча; новые защиты добавлять сюда по мере встречи.
func (r Response) BotProtection() string {
	if r.Status != http.StatusForbidden && r.Status != http.StatusTooManyRequests && r.Status != http.StatusServiceUnavailable {
		return ""
	}
	server := strings.ToLower(r.Header.Get("Server"))
	switch {
	case r.Header.Get("X-Sp-Crid") != "":
		return "ServicePipe"
	case r.Header.Get("Cf-Mitigated") != "" || server == "cloudflare":
		return "Cloudflare"
	case strings.Contains(server, "ddos-guard"):
		return "DDoS-Guard"
	case strings.Contains(server, "qrator"):
		return "Qrator"
	case strings.Contains(strings.ToLower(r.Body), "captcha"):
		return "капча"
	}
	return ""
}
