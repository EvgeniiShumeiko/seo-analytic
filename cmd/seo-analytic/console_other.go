//go:build !windows

package main

// setupConsole на Linux и macOS ничего не делает: терминалы там работают в UTF-8,
// а программы не запускают двойным кликом в отдельном окне.
func setupConsole() func(code int) { return func(int) {} }
