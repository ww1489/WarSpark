//go:build ignore

// generate-coc-api.go 是 cocgen.Generate 的薄包装,保留 go run 独立运行方式。
// 推荐使用 cobra 子命令:warspark cocapi generate
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ww1489/WarSpark/internal/cocgen"
)

func main() {
	root := findProjectRoot()
	swaggerPath := filepath.Join(root, "pkg", "cocapi", "official-swagger.yaml")
	dir := filepath.Join(root, "pkg", "cocapi")

	eps, defs, err := cocgen.Generate(swaggerPath, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate-coc-api: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("生成完成: %d 个端点, %d 个定义\n", eps, defs)
	fmt.Println("下一步:gofmt -w pkg/cocapi/ && go test ./pkg/cocapi/...")
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
