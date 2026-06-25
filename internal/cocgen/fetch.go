// fetch.go 从 Clash of Clans 官方 API 下载最新 Swagger 规范,
// 替换 pkg/cocapi/official-swagger.yaml,自动应用已知缺陷补丁。
//
// 通过 cobra 子命令调用:warspark cocapi fetch --token=xxx
// 或独立运行:go run ./cmd/warspark cocapi fetch --token=xxx
package cocgen

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const SwaggerURL = "https://api.clashofclans.com/v1/"

// FetchSwagger 用 token 从官方 API 下载 swagger 规范,应用补丁后写入 outPath。
func FetchSwagger(token, outPath string) error {
	body, err := downloadSwagger(token)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	if !isSwagger(body) {
		return fmt.Errorf("下载的内容不是 swagger 规范(前 100 字节: %q)", string(body[:min(100, len(body))]))
	}
	patched := applyPatches(string(body))
	if err := os.WriteFile(outPath, []byte(patched), 0o644); err != nil {
		return fmt.Errorf("写入 %s: %w", outPath, err)
	}
	return nil
}

func downloadSwagger(token string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, SwaggerURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

func isSwagger(data []byte) bool {
	head := string(data[:min(50, len(data))])
	return contains(head, "swagger:") || contains(head, "openapi:")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// applyPatches 修复官方 swagger.yaml 的已知缺陷。
// 这些是官方文档自身的不准确之处,每次重新下载都会复现,需在此固定修复。
// 修复清单(与 git 历史对应):
//  1. /clanwarleagues/wars/{warTag} 响应类型 ClanWarLeagueGroup → ClanWar
//  2. ClanCapitalRanking 补齐 name/tag/rank/previousRank/clanLevel/members/badgeUrls/location
//  3. ClanBuilderBaseRanking 同上补齐
//  4. ClientError.detail 与 paths 之间补空行(YAML 解析需要)
func applyPatches(s string) string {
	// 1. GetClanWarLeagueWar 响应类型
	s = replace(s,
		"            $ref: '#/definitions/ClanWarLeagueGroup'\n   /goldpass/seasons/current:",
		"            $ref: '#/definitions/ClanWar'\n  /goldpass/seasons/current:")
	// 兼容另一种缩进
	s = replace(s,
		"            $ref: '#/definitions/ClanWarLeagueGroup'\n  /goldpass/seasons/current:",
		"            $ref: '#/definitions/ClanWar'\n  /goldpass/seasons/current:")

	// 2. ClanCapitalRanking 补齐字段
	s = replace(s,
		"  ClanCapitalRanking:\n    type: object\n    properties:\n      clanPoints:\n        type: integer\n      clanCapitalPoints:\n        type: integer\n",
		"  ClanCapitalRanking:\n    type: object\n    properties:\n      clanPoints:\n        type: integer\n      clanCapitalPoints:\n        type: integer\n      tag:\n        type: string\n      name:\n        type: string\n      rank:\n        type: integer\n      previousRank:\n        type: integer\n      clanLevel:\n        type: integer\n      members:\n        type: integer\n      badgeUrls:\n        type: object\n      location:\n        $ref: '#/definitions/Location'\n")

	// 3. ClanBuilderBaseRanking 补齐字段
	s = replace(s,
		"  ClanBuilderBaseRanking:\n    type: object\n    properties:\n      clanPoints:\n        type: integer\n      clanBuilderBasePoints:\n        type: integer\n",
		"  ClanBuilderBaseRanking:\n    type: object\n    properties:\n      clanPoints:\n        type: integer\n      clanBuilderBasePoints:\n        type: integer\n      tag:\n        type: string\n      name:\n        type: string\n      rank:\n        type: integer\n      previousRank:\n        type: integer\n      clanLevel:\n        type: integer\n      members:\n        type: integer\n      badgeUrls:\n        type: object\n      location:\n        $ref: '#/definitions/Location'\n")

	// 4. ClientError.detail 后补空行(YAML 解析需要)
	s = replace(s,
		"      detail:\n        type: object\npaths:",
		"      detail:\n        type: object\n\npaths:")

	return s
}

func replace(s, old, new string) string {
	if !contains(s, old) {
		return s
	}
	return strings.ReplaceAll(s, old, new)
}
