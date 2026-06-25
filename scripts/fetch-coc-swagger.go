//go:build ignore

// fetch-coc-swagger.go 从 Clash of Clans 官方 API 下载最新 Swagger 规范,
// 替换 pkg/cocapi/official-swagger.yaml,然后调用生成器重新生成代码。
//
// 用法:
//
//	go run scripts/fetch-coc-swagger.go <api-token>
//
// 工作流(API 变动时):
//  1. 在 https://developer.clashofclans.com/#/account 确认 token 有效且 IP 已加白
//  2. 运行本脚本,传入 token
//  3. 脚本自动:下载 swagger → 替换 official-swagger.yaml → 重新生成 types/api/spec
//  4. gofmt -w pkg/cocapi/ && go test ./pkg/cocapi/...
//
// 注意:token 是临时 JWT(约 1 小时有效),不要写入代码或配置文件。
// swagger.yaml 本身是静态规范,不含敏感信息,可留在仓库里。
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const swaggerURL = "https://api.clashofclans.com/v1/"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: go run scripts/fetch-coc-swagger.go <api-token>")
		os.Exit(1)
	}
	token := os.Args[1]

	root := findProjectRoot()
	outPath := filepath.Join(root, "pkg", "cocapi", "official-swagger.yaml")

	fmt.Printf("下载 %s ...\n", swaggerURL)
	body, err := downloadSwagger(token)
	if err != nil {
		fail("下载失败: %v", err)
	}
	fmt.Printf("已下载 %d 字节\n", len(body))

	// 基本校验:确认是 swagger 规范
	if !isSwagger(body) {
		fail("下载的内容不是 swagger 规范(前 100 字节: %q)", string(body[:min(100, len(body))]))
	}

	if err := os.WriteFile(outPath, body, 0o644); err != nil {
		fail("写入 %s: %v", outPath, err)
	}
	fmt.Printf("已替换 %s\n", outPath)

	// 应用官方 swagger 已知缺陷的补丁
	patched := applyPatches(string(body))
	if patched != string(body) {
		if err := os.WriteFile(outPath, []byte(patched), 0o644); err != nil {
			fail("写入补丁后 %s: %v", outPath, err)
		}
		fmt.Println("已应用官方 swagger 缺陷补丁(3 处)")
	}

	fmt.Println("\n下一步:运行生成器重新生成代码")
	fmt.Println("  go run scripts/generate-coc-api.go")
	fmt.Println("  gofmt -w pkg/cocapi/")
	fmt.Println("  go test ./pkg/cocapi/...")
}

func downloadSwagger(token string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, swaggerURL, nil)
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

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		fail("getwd: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	fail("未找到 go.mod")
	return ""
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fetch-coc-swagger: "+format+"\n", args...)
	os.Exit(1)
}
