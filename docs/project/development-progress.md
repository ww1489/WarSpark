# WarSpark 开发进度文档

> 本文是 WarSpark 代码实现的权威快照,与 [implementation-roadmap.md](/E:/Users/ww/Desktop/project/codex/project/WarSpark/docs/project/implementation-roadmap.md) 分工:roadmap 定义阶段目标与验收标准,本文记录代码实际落地情况。每次代码合入后更新本文,并同步 roadmap 的"当前状态"章节。
>
> 代码基线:GinSpark Go REST 模板 + WarSpark 业务层(布局库 / 截图找阵 / 战争数据 / 内部后台)。

## 1. 快照信息

| 项 | 值 |
| --- | --- |
| 生成日期 | 2026-06-25 |
| 分支 | dev_v1_local |
| 最近提交 | 3c636a7 Migrate war module to use pkg/cocapi via adapter |
| 编译 | `go build ./...` 通过 |
| 测试 | `go test ./...` 全部通过(14 个测试包 ok,0 失败) |
| 模块路径 | github.com/ww1489/WarSpark |
| Go 版本 | 见 go.mod |

## 2. Phase 进度总表

| 阶段 | 状态 | 代码落地说明 |
| --- | --- | --- |
| Phase 0 产品文档与数据模型 | 完成 | PRD/MVP/流程/页面/数据政策/路线图/架构/功能技术设计/API 契约/验收清单/运营流程全部成稿 |
| Phase 1 基础数据模型、布局库、内部后台架构 | 完成 | 12 张表迁移落库;布局列表/详情;admin 草稿/链接/视频关联/审核/审计 |
| Phase 2 官方战争 API、截图上传、自动找阵结果页 | 完成 | CoC API 真实接入 + Redis 缓存;上传校验 + 异步 worker + 结果/重试接口;worker 已在 bootstrap 启动;cocapi 独立包重构(06-25):官方 swagger 生成 35 端点客户端 + adapter 模式接入 |
| Phase 3 相关视频和防守回放关联 | 部分 | 数据表 + 公开视频接口 + admin 关联已有;缺独立 video 模块、失效反馈、可访问性检查 |
| Phase 4 战争情报增强和复盘入口 | 未开始 | 历史快照浏览/目标筛选排序/复盘草稿入口/后台缓存刷新均未实现 |
| Phase 5 复盘和资料沉淀 | 未开始 | 复盘页/攻击防守记录/表现统计均未实现 |

## 3. 数据层落地

迁移文件:[migrations/000001_create_warspark_core.up.sql](/E:/Users/ww/Desktop/project/codex/project/WarSpark/migrations/000001_create_warspark_core.up.sql),一次性建立 12 张表,完整覆盖 MVP 数据模型与入库审核。

| 表 | 职责 | 对应文档实体 |
| --- | --- | --- |
| `base_layouts` | 阵型主表(TH/类型/风格/来源/审核/质量/可见性) | 阵型 |
| `layout_images` | 阵型图片(主图/上传图/候选图/参考图,角色+审核+质量) | 阵型图片 |
| `layout_links` | 阵型链接(official_open_layout/source_page/backup + 状态 + 检查时间) | 阵型链接 |
| `videos` | YouTube 视频(youtube_video_id 唯一,标题/频道/发布时间/审核/可见性) | 视频 |
| `layout_video_matches` | 阵型与视频关联(时间戳/分组 attack_video-defense_replay/精度 exact-similar-same_th/星数/摧毁率/置信度) | 阵型视频关联 |
| `image_search_jobs` | 截图找阵任务(状态/识别 TH/截图质量/建筑数/IP 哈希/保留期/目标上下文/错误码) | 搜索任务 |
| `image_search_results` | 找阵结果(任务/阵型/排名/匹配等级/置信度/原因) | 搜索结果 |
| `war_snapshots` | 战争快照(部族 tag/对手/状态/规模/双方星数与摧毁/获取时间/API 错误) | 战争快照 |
| `war_members` | 战争成员(阵营/位置/玩家 tag/TH/已用攻击/最佳防守星数与摧毁) | 战争成员 |
| `war_targets` | 战争目标(成员/找阵任务/位置/名称/TH,绑定找阵上下文) | 战争目标 |
| `import_batches` | 导入批次(来源/状态/总行数/成功失败行数/错误摘要) | 入库批次 |
| `admin_audit_logs` | 后台操作日志(管理员/资源类型/资源 ID/动作/前后快照) | 审计 |

索引与外键已就位(公开筛选索引、状态索引、级联删除)。`war_targets` 和 `import_batches` 已超出 MVP data-model 建议但已落表,为 Phase 2 目标上下文绑定和后续入库预留。

## 4. API 实现矩阵

路由集中装配在 [internal/api/v1/routes.go](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/api/v1/routes.go)。契约见 [docs/api/api-contract.md](/E:/Users/ww/Desktop/project/codex/project/WarSpark/docs/api/api-contract.md)。

### 4.1 已实现接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /health, /api/v1/health, /api/v1/ready | 健康检查与就绪 |
| GET | /api/v1 | 服务信息 |
| GET | /api/v1/layouts | 阵型列表(TH/类型/风格/来源/审核/质量/链接状态筛选 + 分页) |
| GET | /api/v1/layouts/:layout_id | 阵型详情(图片/链接/进攻视频/防守回放/相似阵型) |
| GET | /api/v1/layouts/:layout_id/videos | 阵型相关视频(match_group/match_type 筛选) |
| POST | /api/v1/image-search/jobs | 上传截图建找阵任务(multipart,含目标上下文) |
| GET | /api/v1/image-search/jobs/:job_id | 查询任务状态 |
| GET | /api/v1/image-search/jobs/:job_id/results | 查询找阵结果 |
| POST | /api/v1/image-search/jobs/:job_id/retry | 重试任务 |
| GET | /api/v1/war/current | 当前战争(?clan_tag=,CoC API + 缓存 + 快照落库) |
| GET | /api/v1/war/cwl | CWL 基础结构(?clan_tag=&season=) |
| GET | /api/v1/war/snapshots/:war_snapshot_id/members | 战争成员(?side=clan/opponent + 分页) |
| GET | /uploads, /swagger/*any | 静态图片与 Swagger |

Admin 接口(JWT 鉴权,authmw.Required):

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /api/v1/admin/review-queue | 复核队列(resource_type/审核/质量筛选 + 分页) |
| GET | /api/v1/admin/audit-logs | 后台操作日志 |
| POST | /api/v1/admin/layouts | 创建阵型草稿 |
| POST | /api/v1/admin/layouts/:layout_id/links | 添加阵型链接 |
| POST | /api/v1/admin/layouts/:layout_id/video-matches | 添加视频关联 |
| PATCH | /api/v1/admin/review/:resource_type/:resource_id | 更新审核/质量/可见性 |
| PATCH | /api/v1/admin/layout-links/:link_id | 更新链接状态/检查时间 |

### 4.2 契约已规划但代码未实现

| 方法 | 路径 | 缺失原因 |
| --- | --- | --- |
| POST | /api/v1/videos/:video_id/report-broken | 视频失效反馈接口未实现 |
| GET | /api/v1/clans/:clan_tag | 部族概览未实现(当前战争接口已间接覆盖部族维度) |
| GET | /api/v1/players/:player_tag | 玩家概览未实现 |
| POST | /api/v1/war/targets/:target_id/image-search-jobs | 战争目标找阵独立入口未实现(war_targets 表已建,无 endpoint) |

## 5. 功能模块实现深度

### 5.1 布局库(layout library)— 完成

- 文件:[domain/layout](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/domain/layout/layout.go)、[repository](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/repository/layout_repository.go)、[service](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/service/layout_service.go)、[controller](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/controller/layout_controller.go)、[admin controller](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/controller/admin_layout_controller.go)
- 已实现:列表筛选、详情聚合(图片/链接/进攻视频/防守回放/相似阵型)、admin 草稿创建、链接添加、视频关联、审核状态更新、链接状态更新、复核队列、审计日志写入。
- 缺口:无独立阵型图片管理接口(图片经找阵上传或 admin 草稿间接产生);无相似阵型算法说明(当前按同 TH/类型查询)。

### 5.2 截图找阵(image search)— 完成(匹配算法为占位)

- 文件:[service](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/service/image_search_service.go)、[worker](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/worker/image_search_worker.go)、[repository](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/repository/image_search_repository.go)、[domain](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/domain/imagesearch/imagesearch.go)、[rate limiter](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/infra/redis/image_search_rate_limiter.go)
- 已实现的上传约束:格式 jpg/jpeg/png/webp、单图 10MB、最短边 512px、EXIF 通过解码后 PNG 重编码消除、超过 4096px 用 ApproxBiLinear 归一化、IP 哈希(SHA-256)后 Redis 限流。完全符合 mvp-scope 上传限制。
- 异步链路:worker 在 [bootstrap.go](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/app/bootstrap.go) 以 5s 轮询启动,`ClaimNextJob` 原子认领,处理失败转 `failed`,无 pending 时静默。
- 状态机:created / processing / matched / low_confidence / no_result / failed,支持 retry。
- 关键占位:**匹配算法仅按 detected TH(或目标上下文 expected_th)做候选过滤,取前 5 条,无图像特征相似度计算**。符合 MVP"自动候选匹配、不承诺 100% 精确"定位,但 roadmap 风险点"找阵算法不成熟导致体验不可用"尚未真正解决,见第 9 节 P0 待办。

### 5.3 战争数据(war data)— 完成

- 文件:[service](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/service/war_service.go)、[coc client](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/infra/coc/client.go)、[repository](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/repository/war_repository.go)、[cache](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/infra/redis/war_cache.go)、[domain](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/domain/war/war.go)
- 已实现:真实 CoC 官方 API 接入(`/clans/{tag}/currentwar`、`/clans/{tag}/currentwar/leaguegroup`),Bearer token 鉴权,401/403/404 错误映射(33001 未配置/33002 请求失败/33003 无当前战争/33004 tag 格式)。
- 缓存:current_war 2min、CWL 5min(Redis,可配置),未命中才请求官方 API。
- 快照落库:每次拉取 current war 生成 snapshot + members + targets,计算每位成员最佳防守结果(星数+摧毁率),对手成员自动生成 war_targets。
- tag 校验:`NormalizeClanTag` 统一加 `#` 并校验 `^#?[A-Z0-9]{3,16}$`。
- cocapi 独立包(2026-06-25 重构):`pkg/cocapi/` 从官方 Swagger 2.0 规范生成 35 端点客户端(types.go/api.go 生成,client.go/errors.go/tag.go 手写);`internal/cocgen/` 代码生成器 + `warspark cocapi syncapi/fetch/generate` 三子命令;`internal/infra/coc/client.go` 改写为 adapter,包装 `cocapi.Client` 实现 `service.WarAPIClient` 接口,转换 `cocapi.ClanWar`→`wardomain.CurrentWar`、`cocapi.ClanWarLeagueGroup`→`wardomain.CWLGroup`,错误经 `mapError` 映射为 `wardomain.Error`。service/controller/domain 层通过接口解耦,不感知 cocapi 包。

### 5.4 视频关联(video association)— 部分

- 数据表 `videos`、`layout_video_matches` 已建;公开 `GET /layouts/:id/videos` 按 match_group/match_type 筛选已实现;admin `AddVideoMatch` 已实现(upsert video + 关联)。
- 缺口:无独立 video 模块(domain/service/repository/controller),视频 CRUD 分散在 layout admin;`POST /videos/:id/report-broken` 未实现;视频可访问性自动检查 worker 未实现;视频卡片完整字段(频道/发布时间/星数/摧毁率)在表已存但公开返回结构需对照契约核对。

### 5.5 入库与审核(review and ingestion)— 部分

- 数据表 `import_batches`、`admin_audit_logs` 已建;admin 复核队列、审计日志、审核状态更新已实现。
- 缺口:无导入批次管理接口(创建/查询/重试);无授权导入 worker;无公开页面整理流程;批量入库链路未打通。当前审核入口只覆盖人工后台,自动化产出的低置信/异常样本进入复核队列的写入路径需确认。

### 5.6 战争情报增强与复盘(Phase 4/5)— 未开始

- 历史战争快照浏览、CWL 轮次视图、目标筛选与风险排序、复盘草稿入口、后台缓存刷新管理、攻击/防守记录展示、阵型/视频表现统计、关键资产标记均未实现。

## 6. 架构与运行链路

分层:`cmd -> app -> api -> controller -> service -> repository -> domain`,infra 提供 coc/logger/mysql/redis,worker 独立异步,pkg 提供 jwt/snowflake/cocapi。

启动流程([bootstrap.go](/E:/Users/ww/Desktop/project/codex/project/WarSpark/internal/app/bootstrap.go)):

1. 加载配置(优先级:defaults -> YAML -> `WARSPARK_*` 环境变量 -> CLI)
2. 初始化 logger / MySQL / Redis(带 startup timeout)
3. 构造 JWT tokenManager
4. 装配 Gin 中间件(requestid / recovery / access log / CORS)
5. `v1.SetupRoutes` 注册路由(内部构造 controller/service/repository)
6. 构造 imageSearchService + worker 并 `Start(appCtx)`(5s 轮询)
7. `runHTTPServer`(Unix 用 endless 优雅重启,Windows 用标准 server)

技术债:**依赖装配已统一**(2026-06-25)— `bootstrap.go` 构造 imageSearchService 后注入 `SetupRoutes` 并共享给 worker,`routes.go` 不再重复构造。

## 7. 质量状态

- 编译:`go build ./...` 通过(2026-06-25)。
- 测试:`go test ./...` 全绿(14 个测试包 ok)。有测试的包:api/v1、config、controller、infra/coc、infra/logger、infra/mysql、middleware/auth、middleware/requestid、service、utils、worker、pkg/cocapi、pkg/jwt、pkg/snowflake。
- 格式:`gofmt -l .` 无输出,全部文件符合格式(2026-06-25)。
- 覆盖缺口(无测试文件):`cmd/warspark`、`internal/app`、`internal/cocgen`、`internal/domain/*`、`internal/infra/redis`、`internal/repository`、`middleware/cors|logging|recovery`。
- 持久层风险:`internal/repository`(layout 31KB + image_search 14KB + war 4KB)无任何测试,SQL 正确性依赖人工与运行时验证,是最高价值补测点。
- CI:[.github/workflows/ci.yml](/E:/Users/ww/Desktop/project/codex/project/WarSpark/.github/workflows/ci.yml)。
- Swagger:`docs/swagger.*` 与 `docs/docs.go` 仍是模板生成内容,未随业务接口更新(main.go 描述仍是"Reusable Go REST API backend template")。

## 8. 配置与运行

- 配置文件:`configs/config.dev.yaml`(含 coc 段)、`configs/config.prod.yaml`、`configs/config.example.yaml`。
- CoC 配置:`base_url`/`timeout`/`current_war_cache_ttl`/`cwl_group_cache_ttl` 在 `Validate()` 强制非空;`api_token` 不强制,运行时为空则返回 33001。生产应通过 `WARSPARK_COC_API_TOKEN` 注入。
- 运行:`go run ./cmd/warspark server --config=configs/config.dev.yaml` 或 `make run`。
- 迁移:`make migrate-up` / `make migrate-down STEPS=1` / `make migrate-version`。
- 日志:默认输出到 `logs/warspark-dev.log`(已存在运行痕迹)。
- cocapi 代码生成:`go run ./cmd/warspark cocapi syncapi --token=xxx`(下载 swagger + 生成代码)、`fetch --token=xxx`(仅下载)、`generate`(仅生成)。

## 9. 待办清单(后续开发优先级)

### P0 — MVP 可用性核心

1. 找阵匹配算法:当前仅按 TH 过滤候选。需接入图像特征(感知哈希 pHash / CLIP embedding / 结构相似度)计算置信度,产出真正的 match_level 与 confidence_score,并区分 matched / low_confidence / no_result。这是 MVP 风险点 #1,直接决定首屏体验。
2. 战争目标找阵入口:实现 `POST /api/v1/war/targets/:target_id/image-search-jobs`,把 war_targets 与 image_search_jobs 打通,完成"从战争情报目标进入找阵"的完整闭环。
3. repository 层集成测试:用 Docker MySQL 或 sqlmock 覆盖 layout/image_search/war 持久层,锁定 SQL 与状态机正确性。

### P1 — Phase 3 补齐

4. 视频模块独立化:抽出 video domain/service/repository/controller,补齐 `POST /videos/:id/report-broken`,并实现视频可访问性检查 worker(定期探活 YouTube,失效转 broken)。
5. 视频卡片字段对齐契约:核对 `GET /layouts/:id/videos` 返回是否含 channel_name/published_at/stars/destruction_percent/youtube_url(含时间戳跳转)。
6. 入库与授权导入:实现 import_batches 管理接口与授权导入 worker,打通公开来源/授权入库到复核队列的写入路径。

### P2 — 工程质量与 Phase 4 准备

7. Swagger 同步:用 swag 注解更新业务接口,替换模板描述。
8. 部族/玩家概览:按需实现 `GET /clans/:tag`、`GET /players/:tag`(可由现有 CoC client 扩展)。
9. Phase 4 前置:历史战争快照列表/详情接口、CWL 轮次视图、目标筛选排序、后台缓存刷新管理。

## 10. 技术债

- repository 无测试(见第 7 节)。
- Swagger 与 API 契约脱节。
- image_search 匹配为占位,`buildings_detected` 字段未真正计算(当前仅按尺寸分 good/acceptable)。

## 11. 风险与不做事项(继承 mvp-scope)

不做:登录/注册/个人中心、收藏/订阅/付费/会员、用户侧投稿审核、移动 App、完整 War Report 多端工具、完整视频帧索引、自动识别配兵/流派/打法推荐、AI 总结打法、私有 Discord/Telegram 抓取、登录绕过与付费绕过、YouTube 视频下载、游戏客户端自动化、公开 SEO 运营后台。

风险:找阵算法不成熟(已用低置信标记缓解,但 P0 待办 #1 未完成前体验不可用);官方 API 缓存与权限致战争数据不完整(已用缓存 TTL + 错误码 + 找阵独立可运行缓解);上传接口被匿名滥用(IP 限流已上线)。

## 12. 建议开发顺序

1. P0-1 找阵匹配算法(感知哈希起步,可迭代到 embedding)。
2. P0-2 战争目标找阵入口,打通战争情报到找阵闭环。
3. P1-4 视频模块独立化 + report-broken + 可访问性检查。
4. P0-3 repository 测试(依赖装配已统一,见第 6 节)。
5. P1-6 入库与授权导入,支撑运营流程。
6. Phase 4 战争情报增强(历史快照/目标排序/复盘草稿入口)。
7. Phase 5 复盘与资料沉淀。

## 13. 维护说明

- 每次合入业务代码后更新第 1、2、4、5、9 节。
- 新增接口同步更新第 4 节矩阵与 [api-contract.md](/E:/Users/ww/Desktop/project/codex/project/WarSpark/docs/api/api-contract.md)。
- 阶段验收时同步 roadmap"当前状态"与本文 Phase 总表。
