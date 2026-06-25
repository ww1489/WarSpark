//go:build ignore

// fetch-coc-swagger.go 是 cocgen.FetchSwagger 的薄包装,保留 go run 独立运行方式。
// 推荐使用 cobra 子命令:warspark cocapi fetch --token=xxx
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ww1489/WarSpark/internal/cocgen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: go run scripts/fetch-coc-swagger.go <api-token>")
		os.Exit(1)
	}
	token := os.Args[1]

	root := findProjectRoot()
	outPath := filepath.Join(root, "pkg", "cocapi", "official-swagger.yaml")

	fmt.Printf("下载 %s ...\n", cocgen.SwaggerURL)
	if err := cocgen.FetchSwagger(token, outPath); err != nil {
		fmt.Fprintf(os.Stderr, "fetch-coc-swagger: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已替换 %s(已应用补丁)\n", outPath)
	fmt.Println("\n下一步:warspark cocapi generate 或 go run scripts/generate-coc-api.go")
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
		os.Exit(1)
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
	fmt.Fprintln(os.Stderr, "未找到 go.mod")
	os.Exit(1)
	return ""
}
