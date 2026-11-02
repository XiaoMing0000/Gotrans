package main

import (
	fyneapp "fyne.io/fyne/v2/app"

	"gotrans/internal/ui"
)

// version 在发布构建时由 GoReleaser 通过 -ldflags "-X main.version=..." 注入，
// 本地 / 开发构建为 "dev"。启动时同步给 ui.Version，用于在设置窗口展示。
var version = "dev"

func main() {
	// 仅 Linux/X11 后端需要透明帧缓冲（窗口背景单独做透明度）；实现见 glfw_x11.go，
	// 其余平台为空实现（glfw_stub.go），避免抢先初始化 GLFW 破坏 Fyne 的 X11/CSD 处理。
	initGLFWTransparency()

	ui.Version = version

	a := fyneapp.New()
	w := ui.Build(a)
	w.ShowAndRun()
}
