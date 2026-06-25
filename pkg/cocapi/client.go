// Package cocapi 提供与 Clash of Clans 官方 API v1 交互的轻量级 Go 客户端。
//
// 使用方式:
//
//	client := cocapi.New(cocapi.Config{
//	    BaseURL:  "https://api.clashofclans.com/v1",
//	    APIToken: "your-token",
//	    Timeout:  10 * time.Second,
//	})
//	clan, err := client.GetClan(ctx, "#2PP")
//
// 自动处理:
//   - Bearer token 鉴权 (Authorization 头)
//   - Tag 中的 # 号 URL 编码 (# → %23)
//   - JSON 序列化/反序列化
//   - HTTP 错误映射到类型化错误
//
// spec.yaml 是端点定义的权威来源。添加/修改端点后执行:
//
//	go run ./cmd/warspark cocapi generate
//
// 重新生成 api.go。
package cocapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config 是 CoC API 客户端的配置。
type Config struct {
	BaseURL  string
	APIToken string
	Timeout  time.Duration
}

// Client 是与 Clash of Clans API 交互的 HTTP 客户端。
// 零值不可用，请使用 New 构造。
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// New 创建一个新的 CoC API 客户端。
// 如果 cfg.Timeout <= 0，默认使用 10 秒。
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.clashofclans.com/v1"
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
		apiToken: cfg.APIToken,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do 发送 HTTP 请求并解析 JSON 响应到 result 中。
// path 中的 {param} 占位符会被 args 对应的值替换（自动 URL 编码）。
// 调用方不应处理 result 的零值——当 err != nil 时 result 无效。
func (c *Client) Do(ctx context.Context, method, path string, body any, result any, args ...string) error {
	if c.apiToken == "" {
		return ErrAPINotConfigured
	}

	resolvedPath, err := resolvePath(path, args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTag, err)
	}

	var requestBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		requestBody = bytes.NewReader(data)
	}

	endpoint := c.baseURL + resolvedPath
	req, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	if err := checkStatusCode(resp.StatusCode); err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("%w: %w", ErrAPIResponseInvalid, err)
	}
	return nil
}

// Get 发送 GET 请求。
func (c *Client) Get(ctx context.Context, path string, result any, args ...string) error {
	return c.Do(ctx, http.MethodGet, path, nil, result, args...)
}

// GetWithQuery 发送带查询参数的 GET 请求。
func (c *Client) GetWithQuery(ctx context.Context, path string, query map[string]string, result any, args ...string) error {
	if c.apiToken == "" {
		return ErrAPINotConfigured
	}

	resolvedPath, err := resolvePath(path, args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTag, err)
	}

	endpoint := c.baseURL + resolvedPath
	if len(query) > 0 {
		params := url.Values{}
		for k, v := range query {
			params.Set(k, v)
		}
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	if err := checkStatusCode(resp.StatusCode); err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("%w: %w", ErrAPIResponseInvalid, err)
	}
	return nil
}

// Post 发送 POST 请求。
func (c *Client) Post(ctx context.Context, path string, body any, result any, args ...string) error {
	return c.Do(ctx, http.MethodPost, path, body, result, args...)
}

// resolvePath 用 args 替换 path 中的 {param} 占位符，自动 URL 编码。
func resolvePath(path string, args []string) (string, error) {
	if len(args) == 0 {
		return path, nil
	}
	result := path
	for _, arg := range args {
		encoded := url.PathEscape(arg)
		idx := strings.Index(result, "{}")
		if idx < 0 {
			return "", fmt.Errorf("not enough placeholders in path %q for %d args", path, len(args))
		}
		result = result[:idx] + encoded + result[idx+2:]
	}
	// 用正则查找 {param} 风格的占位符（如 {clanTag}、{playerTag}）
	// 如果还有未被替换的，也是错误
	if strings.Contains(result, "{") && strings.Contains(result, "}") {
		return "", fmt.Errorf("not enough arguments for path %q", path)
	}
	return result, nil
}

// checkStatusCode 将 HTTP 状态码映射到类型化错误。
func checkStatusCode(code int) error {
	switch {
	case code == http.StatusForbidden || code == http.StatusUnauthorized:
		return ErrAPIAccessDenied
	case code == http.StatusNotFound:
		return ErrNotFound
	case code == http.StatusTooManyRequests:
		return ErrRateLimited
	case code >= 200 && code < 300:
		return nil
	default:
		return fmt.Errorf("%w: HTTP %d", ErrAPIRequestFailed, code)
	}
}
