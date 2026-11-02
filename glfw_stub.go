//go:build !x11

package main

// initGLFWTransparency 在非 X11 后端下为空实现：不抢先初始化 GLFW，
// 让 Fyne 自行完成驱动初始化（例如 Wayland 会话下的 CSD / X11 回退处理）。
func initGLFWTransparency() {}
