// Package config 负责在用户配置目录下持久化应用配置。
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// 透明度阈值：下限保证窗口不会完全透明到无法操作。
const (
	MinOpacity = 0.2
	MaxOpacity = 1.0
)

// 分隔条默认位置：原文/译文各占一半；主区域占 80%、划词栏占 20%；划词原文/译文各占一半。
const (
	defaultMainSplitOffset            = 0.5
	defaultSelectionSplitOffset       = 0.8
	defaultSelectionSourceSplitOffset = 0.5
)

// CustomEngine 是用户自定义的翻译引擎（Google 兼容接口）。
type CustomEngine struct {
	// Name 为引擎显示名，也是 Provider 标识。
	Name string `json:"name"`
	// BaseURL 为完整接口地址；空值回退到 Google 默认地址。
	BaseURL string `json:"base_url"`
	// APIKey 为鉴权用 Bearer 令牌。
	APIKey string `json:"api_key"`
}

// Config 是持久化的应用配置。
type Config struct {
	SelectionEnabled    bool    `json:"selection_enabled"`
	ShowSelectionSource bool    `json:"show_selection_source"`
	WindowOpacity       float64 `json:"window_opacity"`
	FontOpacity         float64 `json:"font_opacity"`
	// Provider 为翻译引擎名称（mymemory / google / 自定义引擎名），空值表示默认 mymemory。
	Provider string `json:"provider"`
	// GoogleAPIKey 为 Google 翻译 v2 的 API Key，仅 provider=google 时使用。
	GoogleAPIKey string `json:"google_api_key"`
	// GoogleAPIBaseURL 为 Google 接口地址（完整端点）；空值表示使用默认地址。
	GoogleAPIBaseURL string `json:"google_api_base_url"`
	// APIBaseURL 为 MyMemory 接口地址（基地址，不含 /get 查询路径）；空值表示使用引擎默认地址。
	APIBaseURL string `json:"api_base_url"`
	// MainSplitOffset 为原文/译文区域水平分隔条位置（0-1）。
	MainSplitOffset float64 `json:"main_split_offset"`
	// SelectionSplitOffset 为主区域/划词区域垂直分隔条位置（0-1）。
	SelectionSplitOffset float64 `json:"selection_split_offset"`
	// SelectionSourceSplitOffset 为划词原文/译文区域水平分隔条位置（0-1）。
	SelectionSourceSplitOffset float64 `json:"selection_source_split_offset"`
	// CustomEngines 为用户自定义引擎列表（Google 兼容接口）。
	CustomEngines []CustomEngine `json:"custom_engines"`
}

// Default 返回默认配置。
func Default() Config {
	return Config{
		SelectionEnabled:           true,
		ShowSelectionSource:        true,
		WindowOpacity:              1.0,
		FontOpacity:                1.0,
		Provider:                   "mymemory",
		GoogleAPIKey:               "",
		GoogleAPIBaseURL:           "",
		APIBaseURL:                 "",
		MainSplitOffset:            defaultMainSplitOffset,
		SelectionSplitOffset:       defaultSelectionSplitOffset,
		SelectionSourceSplitOffset: defaultSelectionSourceSplitOffset,
	}
}

// Path 返回配置文件路径，如 ~/.config/gotrans/config.json。
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gotrans", "config.json"), nil
}

// Load 读取配置；文件不存在或损坏时返回默认值。
func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Default(), err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Default(), err
	}
	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	cfg.WindowOpacity = ClampOpacity(cfg.WindowOpacity)
	cfg.FontOpacity = ClampOpacity(cfg.FontOpacity)
	cfg.MainSplitOffset = clampSplit(cfg.MainSplitOffset, defaultMainSplitOffset)
	cfg.SelectionSplitOffset = clampSplit(cfg.SelectionSplitOffset, defaultSelectionSplitOffset)
	cfg.SelectionSourceSplitOffset = clampSplit(cfg.SelectionSourceSplitOffset, defaultSelectionSourceSplitOffset)
	return cfg, nil
}

// Save 将配置写入用户配置目录（目录不存在时自动创建）。
func (c Config) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	c.WindowOpacity = ClampOpacity(c.WindowOpacity)
	c.FontOpacity = ClampOpacity(c.FontOpacity)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// ClampOpacity 把透明度限制在 [MinOpacity, MaxOpacity] 内。
func ClampOpacity(v float64) float64 {
	if v < MinOpacity {
		return MinOpacity
	}
	if v > MaxOpacity {
		return MaxOpacity
	}
	return v
}

// clampSplit 把分隔条位置限制在 (0,1] 内，非法值回退到默认位置。
func clampSplit(v, def float64) float64 {
	if v <= 0 || v > 1 {
		return def
	}
	return v
}
