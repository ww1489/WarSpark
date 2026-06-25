package cocapi

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// tagPattern 匹配 CoC tag 格式:# 开头,后接字母数字,长度 3-16。
// 部族 tag 和玩家 tag 共用同一格式。
var tagPattern = regexp.MustCompile(`^#?[A-Z0-9]{3,16}$`)

// NormalizeTag 将 tag 标准化:去空格、转大写、补 # 前缀。
// 不做格式校验,仅做格式规范化。如需校验用 ValidateTag。
func NormalizeTag(tag string) string {
	t := strings.ToUpper(strings.TrimSpace(tag))
	if t == "" {
		return ""
	}
	if !strings.HasPrefix(t, "#") {
		t = "#" + t
	}
	return t
}

// ValidateTag 校验 tag 格式是否合法,返回标准化后的 tag。
func ValidateTag(tag string) (string, error) {
	normalized := NormalizeTag(tag)
	if normalized == "" {
		return "", fmt.Errorf("tag is required")
	}
	if !tagPattern.MatchString(normalized) {
		return "", fmt.Errorf("%w: %s", ErrInvalidTag, normalized)
	}
	return normalized, nil
}

// encodeTagForPath 对 tag 做 URL 路径编码。# 会被编码为 %23。
func encodeTagForPath(tag string) string {
	return url.PathEscape(NormalizeTag(tag))
}
