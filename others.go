//go:build !windows

package main

// 非Windows平台的空实现
func hideConsoleWindow() {
	// 什么都不做
}