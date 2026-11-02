// Package translator 提供可插拔的翻译引擎抽象。
package translator

import (
	"context"
	"sync/atomic"
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

// providerRef 统一包装 Provider，作为 atomic.Value 的存储类型。
// 不同引擎的具体类型（*mymemory、*google）各不相同，直接存入 atomic.Value
// 会触发「store of inconsistently typed value」panic，故用固定类型包装。
type providerRef struct {
	p Provider
}

// current 保存当前生效的引擎，翻译时实时读取，便于设置界面动态切换。
var current atomic.Value

func init() {
	current.Store(providerRef{newMyMemory()})
}

// New 根据 provider 名称构造引擎：默认或空值为 MyMemory，google 为 Google 翻译。
func New(provider, googleAPIKey string) Provider {
	switch provider {
	case "google":
		return newGoogle(googleAPIKey)
	case "mymemory", "":
		return newMyMemory()
	default:
		// 未知引擎回退到默认实现，避免影响启动。
		return newMyMemory()
	}
}

// NewCustom 构造一个自定义的 Google 兼容引擎：名称 + 接口地址 + Bearer 令牌。
func NewCustom(name, baseURL, apiKey string) Provider {
	return newGoogleCustom(name, baseURL, apiKey)
}

// Current 返回当前生效的引擎。
func Current() Provider { return current.Load().(providerRef).p }

// SetProvider 切换当前引擎（供设置界面调用）。
func SetProvider(p Provider) { current.Store(providerRef{p}) }
