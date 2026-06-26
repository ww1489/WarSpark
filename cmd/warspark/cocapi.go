package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ww1489/WarSpark/pkg/cocapi/cocgen"
)

// cocapi 命令管理 Clash of Clans API 客户端代码的同步与生成。
//
// 用法:
//
//	warspark cocapi syncapi --token=xxx  下载最新 swagger + 重新生成代码(推荐)
//	warspark cocapi fetch --token=xxx    只下载最新 swagger(应用补丁)
//	warspark cocapi generate             只从已有 swagger 生成代码
//
// token 是临时 JWT(约 1 小时有效),在 https://developer.clashofclans.com/#/account 创建,
// 需把当前机器 IP 加入 key 白名单。不要把 token 写入配置文件或提交到仓库。
func newCocapiCommand() *cobra.Command {
	const (
		swaggerPath = "pkg/cocapi/official-swagger.yaml"
		cocapiDir   = "pkg/cocapi"
	)

	cmd := &cobra.Command{
		Use:   "cocapi",
		Short: "Manage Clash of Clans API client (fetch swagger, generate code)",
	}

	// syncapi: fetch + generate 一条龙
	var syncToken string
	syncCmd := &cobra.Command{
		Use:   "syncapi",
		Short: "Fetch latest swagger and regenerate code (recommended)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("下载最新 swagger 到 %s ...\n", swaggerPath)
			if err := cocgen.FetchSwagger(syncToken, swaggerPath); err != nil {
				return fmt.Errorf("fetch: %w", err)
			}
			fmt.Println("✅ swagger 已下载并应用补丁")

			fmt.Printf("从 swagger 生成代码到 %s ...\n", cocapiDir)
			eps, defs, err := cocgen.Generate(swaggerPath, cocapiDir)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}
			fmt.Printf("✅ 生成完成: %d 个端点, %d 个定义\n", eps, defs)
			fmt.Println("\n下一步:gofmt -w pkg/cocapi/ && go test ./pkg/cocapi/...")
			return nil
		},
	}
	syncCmd.Flags().StringVar(&syncToken, "token", "", "Clash of Clans API token (JWT)")
	_ = syncCmd.MarkFlagRequired("token")
	cmd.AddCommand(syncCmd)

	// fetch: 只下载
	var fetchToken string
	fetchCmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch latest swagger only (apply patches)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("下载最新 swagger 到 %s ...\n", swaggerPath)
			if err := cocgen.FetchSwagger(fetchToken, swaggerPath); err != nil {
				return fmt.Errorf("fetch: %w", err)
			}
			fmt.Println("✅ swagger 已下载并应用补丁")
			fmt.Println("\n下一步:warspark cocapi generate")
			return nil
		},
	}
	fetchCmd.Flags().StringVar(&fetchToken, "token", "", "Clash of Clans API token (JWT)")
	_ = fetchCmd.MarkFlagRequired("token")
	cmd.AddCommand(fetchCmd)

	// generate: 只生成
	cmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate code from existing swagger (no download)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("从 %s 生成代码到 %s ...\n", swaggerPath, cocapiDir)
			eps, defs, err := cocgen.Generate(swaggerPath, cocapiDir)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}
			fmt.Printf("✅ 生成完成: %d 个端点, %d 个定义\n", eps, defs)
			fmt.Println("\n下一步:gofmt -w pkg/cocapi/ && go test ./pkg/cocapi/...")
			return nil
		},
	})

	return cmd
}
