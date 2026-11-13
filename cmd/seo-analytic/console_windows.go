//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const cpUTF8 = 65001

// В x/sys/windows нет обёртки для GetConsoleProcessList — вызываем её из kernel32 напрямую.
var procGetConsoleProcessList = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleProcessList")

// setupConsole готовит консоль Windows и возвращает функцию, которую нужно вызвать перед выходом.
//
// Консоль по умолчанию работает в кодировке 866, а программа пишет в UTF-8 — без переключения русский
// текст превращается в кракозябры. При запуске двойным кликом Windows создаёт окно только для программы
// и закрывает его сразу после выхода — тогда перед закрытием ждём Enter, чтобы отчёт можно было прочитать.
func setupConsole() func(code int) {
	oldCP, err := windows.GetConsoleOutputCP()
	switched := err == nil && windows.SetConsoleOutputCP(cpUTF8) == nil
	ownWindow := consoleProcessCount() == 1

	return func(code int) {
		// 130 — пользователь сам прервал программу, держать окно незачем.
		if ownWindow && code != 130 {
			fmt.Fprint(os.Stderr, "\nНажмите Enter, чтобы закрыть окно...")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
		// Возвращаем кодировку, иначе она останется UTF-8 в консоли, из которой нас запустили.
		if switched {
			windows.SetConsoleOutputCP(oldCP)
		}
	}
}

// consoleProcessCount — сколько процессов подключено к нашей консоли. Если запустили из cmd или
// PowerShell, их минимум два (оболочка и мы); если двойным кликом — только мы. При ошибке возвращает 0.
func consoleProcessCount() int {
	ids := make([]uint32, 2) // больше двух знать не нужно — важно лишь, одни ли мы
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&ids[0])), uintptr(len(ids)))
	return int(n)
}
