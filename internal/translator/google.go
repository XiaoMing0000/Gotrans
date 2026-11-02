package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// DefaultGoogleAPIBase 是 Google Cloud Translation v2 接口地址（完整端点）。
const DefaultGoogleAPIBase = "https://translation.googleapis.com/language/translate/v2"

// googleAPIBase 保存当前生效的 Google 接口地址，翻译时实时读取，便于设置界面动态更新。
var googleAPIBase atomic.Value

func init() {
	googleAPIBase.Store(DefaultGoogleAPIBase)
}

// SetGoogleAPIBase 更新 Google 接口地址；空值或纯空白回退到默认地址。
func SetGoogleAPIBase(base string) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = DefaultGoogleAPIBase
	}
	googleAPIBase.Store(base)
}

// GoogleAPIBase 返回当前生效的 Google 接口地址。
func GoogleAPIBase() string { return googleAPIBase.Load().(string) }

// google 使用 Google Cloud Translation v2（需要 API Key）。
// 文档：https://cloud.google.com/translate/docs/reference/rest/v2/translate
type google struct {
	client  *http.Client
	apiKey  string
	name    string
	baseURL string // 空值表示使用全局 googleAPIBase
}

func newGoogle(apiKey string) *google {
	return &google{
		client: &http.Client{Timeout: 15 * time.Second},
		apiKey: strings.TrimSpace(apiKey),
		name:   "Google",
	}
}

// newGoogleCustom 构造一个自定义的 Google 兼容引擎（独立名称、接口地址与 Key）。
func newGoogleCustom(name, baseURL, apiKey string) *google {
	return &google{
		client:  &http.Client{Timeout: 15 * time.Second},
		apiKey:  strings.TrimSpace(apiKey),
		name:    strings.TrimSpace(name),
		baseURL: strings.TrimSpace(baseURL),
	}
}

func (g *google) Name() string { return g.name }

func (g *google) Languages() []Language { return languages }

// googleAPIError 兼容两种错误返回格式：Google 官方为对象（code/message），
// MTranServer 等自建兼容服务为字符串（如 "Unauthorized"）。
type googleAPIError struct {
	Code    int
	Message string
}

func (e *googleAPIError) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Message = s
		return nil
	}
	var obj struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	e.Code = obj.Code
	e.Message = obj.Message
	return nil
}

type googleResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText string `json:"translatedText"`
		} `json:"translations"`
	} `json:"data"`
	Error *googleAPIError `json:"error"`
}

func (g *google) Translate(ctx context.Context, text string, from, to Language) (string, error) {
	if g.apiKey == "" {
		return "", fmt.Errorf("未配置 Google API Key，请在设置中填写")
	}

	payload := map[string]string{
		"q":      text,
		"source": from.Code,
		"target": to.Code,
		"format": "text",
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("构造请求失败: %w", err)
	}

	base := g.baseURL
	if base == "" {
		base = googleAPIBase.Load().(string)
	}
	base = strings.TrimRight(base, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("构造请求失败: %w", err)
	}
	// MTranServer 等自建 Google 兼容服务通过 Bearer 头鉴权（JSON 请求体）；
	// 官方 Google v2 也接受 JSON 请求体。密钥不放在 URL 或表单中，避免泄露。
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求翻译服务失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var parsed googleResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if parsed.Error != nil {
		if parsed.Error.Code != 0 {
			return "", fmt.Errorf("Google 翻译返回错误(%d): %s", parsed.Error.Code, parsed.Error.Message)
		}
		return "", fmt.Errorf("Google 翻译返回错误: %s", parsed.Error.Message)
	}
	if len(parsed.Data.Translations) == 0 {
		return "", fmt.Errorf("Google 翻译未返回译文")
	}
	return parsed.Data.Translations[0].TranslatedText, nil
}
