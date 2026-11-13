@echo off
rem Собирает seo-analytic под все поддерживаемые системы в папку dist\.
rem Запуск: scripts\build.bat из командной строки или двойной клик по файлу.
setlocal EnableDelayedExpansion
rem UTF-8, чтобы русские сообщения не превратились в кракозябры.
chcp 65001 >nul

rem Переходим в корень проекта, откуда бы ни запустили скрипт.
cd /d "%~dp0.."

where go >nul 2>&1 || (
    echo Go не найден. Установите Go 1.26+ с https://go.dev/dl/ и перезапустите консоль.
    goto :fail
)

if not exist dist mkdir dist

rem CGO не нужен: бинарник статический и запускается без установленного Go.
set CGO_ENABLED=0
rem Цели записаны через дефис: символ / в списке for у cmd ненадёжен.
for %%T in (linux-amd64 linux-arm64 windows-amd64 windows-arm64 darwin-amd64 darwin-arm64) do (
    for /f "tokens=1,2 delims=-" %%A in ("%%T") do (
        set "GOOS=%%A"
        set "GOARCH=%%B"
        set "OUT=dist\seo-analytic-%%A-%%B"
        if "%%A"=="windows" set "OUT=!OUT!.exe"
        echo -^> !OUT!
        rem -trimpath убирает локальные пути из бинарника, -s -w — отладочную информацию.
        go build -trimpath -ldflags "-s -w" -o "!OUT!" ./cmd/seo-analytic || goto :fail
    )
)

rem Контрольные суммы в формате sha256sum. certutil печатает три строки,
rem строка с хешем — единственная без двоеточия.
(
    for %%F in (dist\seo-analytic-*) do (
        for /f "delims=" %%H in ('certutil -hashfile "%%F" SHA256 ^| findstr /v ":"') do echo %%H  %%~nxF
    )
) > dist\SHA256SUMS

echo готово: dist\
call :pause_if_doubleclick
exit /b 0

:fail
echo Сборка не удалась.
call :pause_if_doubleclick
exit /b 1

rem При запуске двойным кликом окно закрылось бы сразу — ждём нажатия клавиши.
:pause_if_doubleclick
echo %cmdcmdline% | find /i "%~nx0" >nul && pause
exit /b 0
