package cocapi

import (
	"errors"
	"fmt"
)

// 错误哨兵值。API 返回的错误可通过 errors.Is 判断类别。
var (
	// ErrAPINotConfigured 表示客户端未配置 API token。
	ErrAPINotConfigured = errors.New("coc api token is not configured")
	// ErrAPIRequestFailed 表示 HTTP 请求失败(网络错误或非 2xx 状态码)。
	ErrAPIRequestFailed = errors.New("coc api request failed")
	// ErrAPIAccessDenied 表示 401/403,通常是 token 无效或 IP 未加白。
	ErrAPIAccessDenied = errors.New("coc api access denied")
	// ErrNotFound 表示请求的资源(部族/玩家/战争等)不存在。
	ErrNotFound = errors.New("coc api resource not found")
	// ErrRateLimited 表示请求过快被官方限流(429)。
	ErrRateLimited = errors.New("coc api rate limited")
	// ErrAPIResponseInvalid 表示响应 JSON 解析失败。
	ErrAPIResponseInvalid = errors.New("coc api response invalid")
	// ErrInvalidTag 表示传入的 tag 格式错误或路径参数数量不匹配。
	ErrInvalidTag = errors.New("invalid tag or path argument")
)

// APIError 携带 HTTP 状态码和原始消息,用于需要更多上下文的场景。
type APIError struct {
	Code    int
	Message string
	Wrapped error
}

func (e *APIError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("coc api error (HTTP %d): %s: %v", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("coc api error (HTTP %d): %s", e.Code, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Wrapped
}
