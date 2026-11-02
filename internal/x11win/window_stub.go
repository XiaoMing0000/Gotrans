//go:build !linux

package x11win

import "errors"

// errUnsupported 表示当前平台不支持 X11 窗口操作。非 Linux 上所有函数都返回该错误，
// 使 internal/ui/float.go 的悬浮窗拖拽/缩放/置顶/圆角逻辑自动降级
// （applyBorderless 仅在 RemoveDecorationsByTitle/PositionByTitle 都成功后才置 ready）。
var errUnsupported = errors.New("x11win: 当前平台不支持 X11 窗口操作")

func MoveByTitle(title string, x, y int32) error {
	return errUnsupported
}

func ResizeByTitle(title string, w, h uint32) error {
	return errUnsupported
}

func MoveResizeByTitle(title string, x, y int32, w, h uint32) error {
	return errUnsupported
}

func RemoveDecorationsByTitle(title string) error {
	return errUnsupported
}

func SetAlwaysOnTopByTitle(title string, on bool) error {
	return errUnsupported
}

func PositionByTitle(title string) (int32, int32, error) {
	return 0, 0, errUnsupported
}

func SizeByTitle(title string) (uint32, uint32, error) {
	return 0, 0, errUnsupported
}

func PointerPosition() (int32, int32, error) {
	return 0, 0, errUnsupported
}

func SetRoundedCornersByTitle(title string, radius int) error {
	return errUnsupported
}
