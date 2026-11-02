//go:build linux && x11

package main

import "github.com/go-gl/glfw/v3.4/glfw"

// initGLFWTransparency 提前初始化 GLFW（幂等，Fyne 之后会再次调用并直接复用），
// 以便在创建窗口前请求带 alpha 通道的透明帧缓冲：窗口背景可单独做透明度，
// 而文字透明度走主题、不受窗口透明度影响。若仍走 X11 整窗属性
// （_NET_WM_WINDOW_OPACITY），文字会随整窗一起淡化，无法独立配置。
func initGLFWTransparency() {
	// X11 下把窗口的 WM_CLASS（类名/实例名）固定为 "gotrans"。若不设置，
	// GLFW 会退回用窗口标题当作 WM_CLASS（标题含空格与非 ASCII 字符），
	// GNOME 的 dash 面板与 Super+Tab 切换器就无法把运行中的窗口匹配到
	// build/gotrans.desktop，于是显示为默认的齿轮图标。
	glfw.WindowHintString(glfw.X11ClassName, "gotrans")
	glfw.WindowHintString(glfw.X11InstanceName, "gotrans")

	if glfw.Init() == nil {
		glfw.WindowHint(glfw.AlphaBits, 8)
		glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	}
}
