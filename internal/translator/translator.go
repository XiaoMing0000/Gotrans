// Package translator 提供可插拔的翻译引擎抽象。
package translator

import (
	"context"
	"os"
)

// Language 表示一种语言。Code 为 RFC3066 语言代码（如 "zh-CN"）。
type Language struct {
	Code string
	Name string
}

// Provider 是翻译引擎的抽象接口。新增引擎只需实现该接口并在 New 中注册。
type Provider interface {
	// Name 返回引擎名称，用于状态展示与日志。
	Name() string
	// Translate 将 text 从 from 翻译到 to。
	Translate(ctx context.Context, text string, from, to Language) (string, error)
	// Languages 返回该引擎支持的语种列表。
	Languages() []Language
}

// New 返回当前生效的翻译引擎。
// 通过环境变量 GOTRANS_PROVIDER 选择，默认使用免费的 MyMemory。
func New() Provider {
	switch os.Getenv("GOTRANS_PROVIDER") {
	case "mymemory", "":
		return newMyMemory()
	default:
		// 未知引擎回退到默认实现，避免影响启动。
		return newMyMemory()
	}
}
