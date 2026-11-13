#!/usr/bin/env sh
# Собирает seo-analytic под все поддерживаемые системы в папку dist/.
# Запуск из корня проекта: ./scripts/build.sh
set -eu

cd "$(dirname "$0")/.."
mkdir -p dist

for target in linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64; do
	os=${target%/*}
	arch=${target#*/}
	out="dist/seo-analytic-$os-$arch"
	[ "$os" = windows ] && out="$out.exe"
	echo "→ $out"
	# CGO не нужен — бинарник статический и запускается без установленного Go.
	# -trimpath убирает локальные пути из бинарника, -s -w — отладочную информацию (минус ~25% размера).
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w" -o "$out" ./cmd/seo-analytic
done

# Контрольные суммы, чтобы скачавший мог проверить файл.
(cd dist && shasum -a 256 seo-analytic-* > SHA256SUMS 2>/dev/null || sha256sum seo-analytic-* > SHA256SUMS)
echo "готово: dist/"
