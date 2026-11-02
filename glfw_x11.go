//go:build linux && x11

package main

import "github.com/go-gl/glfw/v3.4/glfw"

// initGLFWTransparency 提前初始化 GLFW（幂等，Fyne 之后会再次调用并直接复用），
// 以便在创建窗口前请求带 alpha 通道的透明帧缓冲：窗口背景可单独做透明度，
// 而文字透明度走主题、不受窗口透明度影响。若仍走 X11 整窗属性
// （_NET_WM_WINDOW_OPACITY），文字会随整窗一起淡化，无法独立配置。
func initGLFWTransparency() {
	if glfw.Init() == nil {
		glfw.WindowHint(glfw.AlphaBits, 8)
		glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	}
}
