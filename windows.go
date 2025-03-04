//go:build windows

package main

import "syscall"

func hideConsoleWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// 隐藏窗口
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	if hwnd, _, _ := getConsoleWindow.Call(); hwnd != 0 {
		user32 := syscall.NewLazyDLL("user32.dll")
		showWindow := user32.NewProc("ShowWindow")
		showWindow.Call(hwnd, 0)
	}

	// 释放控制台
	freeConsole := kernel32.NewProc("FreeConsole")
	freeConsole.Call()
}