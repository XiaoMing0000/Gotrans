package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

// logo.png 应用/窗口图标：由 assets/logo.svg 用 rsvg-convert 转出（librsvg），
// 以 go:embed 内嵌进二进制，保证单文件交付。
//
// 注意不要直接内嵌 SVG 作为窗口图标：Fyne 的 SetIcon 会走其内置 SVG 光栅化器
// （github.com/fyne-io/oksvg），它不支持该 logo 用到的 <style> 类选择器与
// fill-rule="evenodd"，会把图标渲染成全透明，导致任务栏/切换器看不到 logo。
//
//go:embed assets/logo.png
var logoBytes []byte

// appIcon 返回内嵌的应用图标资源。Fyne 在 SetIcon 时按 PNG 直接解码，
// 不经过 SVG 光栅化，因此能正确显示。
func appIcon() fyne.Resource {
	return fyne.NewStaticResource("logo.png", logoBytes)
}
