//go:build !linux

package selection

import "errors"

// Watcher 在非 Linux 平台上为空壳，仅用于保持与 Linux 版相同的类型签名，
// 使 internal/ui/app.go 与 watchSelection 无需平台分支即可编译。
type Watcher struct{}

// New 在非 Linux 平台直接返回错误，app.go 据此显示「划词翻译不可用」。
func New() (*Watcher, error) {
	return nil, errors.New("划词翻译仅支持 Linux/X11，当前平台不可用")
}

// Close 无操作。
func (*Watcher) Close() {}

// ReadPrimary 在非 Linux 平台不可用。
func (*Watcher) ReadPrimary() (string, error) {
	return "", errors.New("划词翻译仅支持 Linux/X11，当前平台不可用")
}
