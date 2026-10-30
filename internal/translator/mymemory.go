package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// DefaultAPIBase 是 MyMemory 引擎的默认接口地址（基地址，不含 /get）。
const DefaultAPIBase = "https://api.mymemory.translated.net"

// apiBase 保存当前生效的接口地址，翻译时实时读取，便于设置界面动态更新。
var apiBase atomic.Value

func init() {
	apiBase.Store(DefaultAPIBase)
}

// SetAPIBase 更新接口地址；空值或纯空白回退到默认地址。
func SetAPIBase(base string) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = DefaultAPIBase
	}
	apiBase.Store(base)
}

// APIBase 返回当前生效的接口地址。
func APIBase() string { return apiBase.Load().(string) }

// mymemory 使用 MyMemory 免费翻译 API（无需 Key）。
// 文档：https://mymemory.translated.net/doc/spec.php
type mymemory struct {
	client *http.Client
}

func newMyMemory() *mymemory {
	return &mymemory{client: &http.Client{Timeout: 15 * time.Second}}
}

func (m *mymemory) Name() string { return "MyMemory" }

// languages 是 v1 预置的常用语种。
var languages = []Language{
	{Code: "en", Name: "英语"},
	{Code: "zh-CN", Name: "中文（简体）"},
	{Code: "zh-TW", Name: "中文（繁体）"},
	{Code: "ja", Name: "日语"},
	{Code: "ko", Name: "韩语"},
	{Code: "es", Name: "西班牙语"},
	{Code: "fr", Name: "法语"},
	{Code: "de", Name: "德语"},
	{Code: "ru", Name: "俄语"},
	{Code: "pt", Name: "葡萄牙语"},
	{Code: "it", Name: "意大利语"},
	{Code: "ar", Name: "阿拉伯语"},
	{Code: "hi", Name: "印地语"},
	{Code: "vi", Name: "越南语"},
	{Code: "th", Name: "泰语"},
}

func (m *mymemory) Languages() []Language { return languages }

type mymemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	ResponseStatus  int    `json:"responseStatus"`
	ResponseDetails string `json:"responseDetails"`
}

func (m *mymemory) Translate(ctx context.Context, text string, from, to Language) (string, error) {
	q := url.Values{}
	q.Set("q", text)
	q.Set("langpair", from.Code+"|"+to.Code)

	base := strings.TrimRight(apiBase.Load().(string), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/get?"+q.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("构造请求失败: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求翻译服务失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var parsed mymemoryResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if parsed.ResponseStatus != 200 {
		return "", fmt.Errorf("翻译服务返回错误(%d): %s", parsed.ResponseStatus, parsed.ResponseDetails)
	}
	return parsed.ResponseData.TranslatedText, nil
}
