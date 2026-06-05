# WarSpark 项目文档说明与完成状态

## 1. 文档目的

本文档用于记录 WarSpark 项目需要维护的核心文档、每份文档的用途、优先级、依赖关系和完成状态。它是项目文档的总控表，帮助项目经理、产品经理和开发负责人判断下一步应该补齐哪些资料。

WarSpark 当前处于产品定义和 MVP 边界收敛阶段，文档工作的重点不是追求数量，而是先把以下问题讲清楚：

- 首版到底做什么，不做什么。
- 用户从哪里进入，怎么完成核心流程。
- 每个页面展示什么、有哪些状态和操作。
- 找阵、阵型库、视频、战争数据之间的数据关系。
- 外部内容和数据来源有哪些边界。
- 开发阶段如何拆分，如何验收。

## 2. 状态定义

| 状态 | 含义 |
| --- | --- |
| 已完成 | 已写入仓库，可以作为后续文档或开发依据；仍可迭代。 |
| 待编写 | 尚未创建，是近期需要补齐的文档。 |
| 后续补充 | 当前阶段不阻塞 MVP，但在进入开发或运营前需要补齐。 |
| 模板既有 | 从 GinSpark 后端模板继承而来，暂时保留，后续可按 WarSpark 产品重写。 |

## 3. 当前文档总览

| 文档 | 状态 | 优先级 | 作用 | 依赖 |
| --- | --- | --- | --- | --- |
| `docs/product/warspark-prd.md` | 已完成 | P0 | WarSpark 总版 PRD，定义战前情报站定位、产品主线、MVP、阶段规划和竞品参考。 | 无 |
| `docs/project/documentation-status.md` | 已完成 | P0 | 记录项目需要哪些文档、每份文档的说明和完成情况。 | `warspark-prd.md` |
| `docs/product/mvp-scope.md` | 已完成 | P0 | 明确 MVP 做什么、不做什么、验收边界和推迟能力。 | `warspark-prd.md` |
| `docs/product/user-flows.md` | 已完成 | P0 | 记录核心用户流程：截图找阵、战争目标找阵、找阵沉淀阵型库、复盘沉淀资料。 | `warspark-prd.md`, `mvp-scope.md` |
| `docs/product/page-spec.md` | 已完成 | P0 | 页面级产品规格，定义页面模块、字段、按钮、空状态、错误状态。 | `user-flows.md` |
| `docs/product/content-data-policy.md` | 已完成 | P0 | 定义外部内容和数据使用边界，包括用户上传、公开页面、OpenLayout、YouTube 时间戳、官方 API、授权导入。 | `warspark-prd.md` |
| `docs/architecture/data-model.md` | 已完成 | P0 | 定义核心数据模型和实体关系，如阵型、图片、视频、匹配记录、搜索任务、战争快照。 | `mvp-scope.md`, `page-spec.md` |
| `docs/project/implementation-roadmap.md` | 已完成 | P0 | 把 MVP 拆成开发阶段、交付物、验收标准和风险。 | `mvp-scope.md`, `data-model.md` |
| `docs/architecture/system-architecture.md` | 待编写 | P1 | 定义后端模块、图片处理、缓存、异步任务、外部 API、存储和部署边界。 | `data-model.md`, `implementation-roadmap.md` |
| `docs/architecture/features/image-search.md` | 待编写 | P1 | 定义截图找阵实现逻辑，包括上传任务、处理状态、候选匹配、低置信度和无结果。 | `system-architecture.md`, `mvp-scope.md`, `data-model.md` |
| `docs/architecture/features/layout-library.md` | 待编写 | P1 | 定义轻量阵型库实现逻辑，包括阵型入库、列表筛选、详情页、OpenLayout 状态。 | `system-architecture.md`, `page-spec.md`, `data-model.md` |
| `docs/architecture/features/video-association.md` | 待编写 | P1 | 定义相关攻击视频和防守回放关联逻辑，包括时间戳、关联类型、失效状态。 | `system-architecture.md`, `content-data-policy.md`, `data-model.md` |
| `docs/architecture/features/war-data.md` | 待编写 | P1 | 定义战争数据接入逻辑，包括官方 API、缓存、错误状态、目标上下文绑定。 | `system-architecture.md`, `data-model.md` |
| `docs/architecture/features/review-and-ingestion.md` | 待编写 | P1 | 定义内容入库和审核逻辑，包括公开来源、授权导入、质量状态、链接状态。 | `content-data-policy.md`, `data-model.md` |
| `docs/api/api-contract.md` | 后续补充 | P1 | 定义前后端 API 契约，包括上传截图、查询结果、阵型库、视频、战争数据。 | `page-spec.md`, `data-model.md`, `features/*` |
| `docs/qa/acceptance-checklist.md` | 后续补充 | P1 | 定义功能验收清单，确保开发结果和产品目标一致。 | `page-spec.md`, `api-contract.md` |
| `docs/ops/content-operations.md` | 后续补充 | P1 | 定义阵型图片、OpenLayout、视频时间戳、失效链接和审核状态的运营流程。 | `content-data-policy.md`, `review-and-ingestion.md` |
| `docs/design.md` | 模板既有 | P2 | GinSpark 模板架构设计说明，当前主要用于理解后端模板结构。 | 无 |
| `docs/template-spec.md` | 模板既有 | P2 | GinSpark 模板规格说明，后续可替换为 WarSpark 工程规范。 | 无 |
| `docs/AI_DEVELOPMENT.md` | 模板既有 | P2 | AI 辅助开发说明，后续可按 WarSpark 工作流更新。 | 无 |
| `docs/swagger.*` / `docs/docs.go` | 模板既有 | P2 | 当前后端模板生成的 Swagger 文档，后续随 API 开发更新。 | 后端接口实现 |

## 4. 第一批必须完成的文档

第一批文档用于防止项目范围发散，是开始功能开发前最该补齐的部分。

### 4.1 MVP 范围文档

路径：`docs/product/mvp-scope.md`

需要回答：

- MVP 的核心目标是什么。
- 截图找阵、轻量阵型库、相关视频、防守回放、战争数据预留分别做到什么程度。
- 哪些功能明确不做：登录、收藏、付费、完整 War Report、多端 App、完整视频帧索引、自动配兵识别、AI 打法推荐。
- 每个模块的验收标准是什么。

完成标准：

- 能让开发者直接判断一个需求是否属于 MVP。
- 能让产品决策时避免临时加范围。
- 能成为开发排期和验收清单的输入。

### 4.2 用户流程文档

路径：`docs/product/user-flows.md`

需要回答：

- 用户如何从首页进入找阵。
- 用户如何从战争目标进入找阵。
- 找阵结果如何进入阵型详情。
- 找阵结果如何沉淀为阵型库资产。
- 战后如何把攻击记录和防守结果沉淀到资料库。

完成标准：

- 每条流程都有开始、关键步骤、结束状态。
- 明确正常流程、低置信度流程、无结果流程、外部链接失效流程。
- 后续页面规格可以直接引用流程编号。

### 4.3 页面规格文档

路径：`docs/product/page-spec.md`

需要回答：

- 作战台、找阵页、找阵结果页、阵型库列表、阵型详情、战争情报页分别有哪些模块。
- 每个模块需要哪些字段。
- 每个页面有哪些用户操作。
- 空状态、加载状态、错误状态如何展示。

完成标准：

- 前端可以据此拆页面。
- 后端可以据此确认接口字段。
- QA 可以据此提炼页面验收用例。

### 4.4 内容与数据边界文档

路径：`docs/product/content-data-policy.md`

需要回答：

- 哪些数据可以使用：用户上传、公开阵型页面、OpenLayout、官方 CoC API、用户授权导入、人工整理。
- 哪些数据不默认使用：私有 Discord / Telegram 抓取、登录绕过、付费内容绕过、YouTube 视频下载、游戏客户端自动化。
- 阵型图片、链接、视频时间戳如何记录来源和审核状态。

完成标准：

- 数据采集和内容运营有明确边界。
- 后续开发不会默认实现高风险采集路径。
- 能支撑阵型库的来源、置信度和审核字段设计。

## 5. 第二批开发前文档

第二批文档用于把产品规格转成工程可执行方案。

### 5.1 数据模型文档

路径：`docs/architecture/data-model.md`

核心实体建议：

- `base_layouts`：阵型主表。
- `layout_images`：阵型图片。
- `layout_links`：OpenLayout 和其他复制链接。
- `videos`：YouTube 视频记录。
- `layout_video_matches`：阵型和视频的关联。
- `image_search_jobs`：截图找阵任务。
- `image_search_results`：找阵结果。
- `war_snapshots`：战争数据快照。
- `war_members`：战争成员状态。

完成标准：

- 明确每个实体的职责。
- 明确实体之间的关系。
- 明确哪些字段属于 MVP，哪些字段预留。

### 5.2 实施路线图

路径：`docs/project/implementation-roadmap.md`

建议阶段：

- Phase 0：产品文档和数据模型。
- Phase 1：轻量阵型库和人工样本数据。
- Phase 2：截图上传和找阵结果页。
- Phase 3：相关视频和防守回放关联。
- Phase 4：战争数据基础接入。
- Phase 5：复盘和资料沉淀。

完成标准：

- 每个阶段都有明确交付物。
- 每个阶段都有验收标准。
- 每个阶段列出主要风险和不做事项。

## 6. 开发前技术设计文档

### 6.1 系统架构文档

路径：`docs/architecture/system-architecture.md`

用于说明后端模块、图片上传处理、对象存储、数据库、缓存、异步任务、外部 API 和部署方式。

完成时机：

- 在所有功能技术设计文档之前完成。
- 作为后续 `features/*` 文档的上层架构约束。

### 6.2 截图找阵技术设计

路径：`docs/architecture/features/image-search.md`

用于说明截图找阵的实现逻辑，包括上传、任务状态、TH 识别、候选匹配、低置信度、无结果、失败重试和结果沉淀。

完成时机：

- 在 `system-architecture.md` 之后完成。
- 在实现截图上传、找阵任务和找阵结果页之前完成。

### 6.3 阵型库技术设计

路径：`docs/architecture/features/layout-library.md`

用于说明轻量阵型库的实现逻辑，包括阵型入库、审核状态、OpenLayout 状态、列表筛选、详情页和找阵结果关联。

完成时机：

- 在 `system-architecture.md` 之后完成。
- 在实现阵型库列表页和阵型详情页之前完成。

### 6.4 视频关联技术设计

路径：`docs/architecture/features/video-association.md`

用于说明相关攻击视频和防守回放的关联逻辑，包括 YouTube 时间戳、精确关联、相似参考、同 TH 参考、视频失效和无视频状态。

完成时机：

- 在 `image-search.md` 和 `layout-library.md` 之后完成。
- 在实现找阵结果页的视频分组之前完成。

### 6.5 战争数据技术设计

路径：`docs/architecture/features/war-data.md`

用于说明战争数据接入逻辑，包括 Clash of Clans 官方 API、缓存、错误状态、战争快照、成员数据和目标上下文绑定。

完成时机：

- 可以在 MVP 找阵和阵型库技术设计之后完成。
- 在实现 V1 战争情报页之前完成。

### 6.6 入库与审核技术设计

路径：`docs/architecture/features/review-and-ingestion.md`

用于说明内容入库、来源记录、审核状态、质量状态、链接状态、公开页面整理和授权导入的技术逻辑。

完成时机：

- 在 `content-data-policy.md` 和 `data-model.md` 之后完成。
- 在实现内容导入、审核或运营流程之前完成。

## 7. 后续接口、验收与运营文档

### 7.1 API 契约文档

路径：`docs/api/api-contract.md`

用于定义前后端接口，不依赖口头约定推进开发。

完成时机：

- 在 `system-architecture.md` 和核心 `features/*` 文档完成后编写。
- 不应早于功能技术设计，否则接口容易反复改。

### 7.2 验收清单

路径：`docs/qa/acceptance-checklist.md`

用于在每个阶段结束时检查功能是否真正达到产品要求。

完成时机：

- 在 `api-contract.md` 和页面规格稳定后编写。

### 7.3 内容运营文档

路径：`docs/ops/content-operations.md`

用于定义阵型、图片、链接、视频、时间戳、审核状态和失效链接处理流程。

完成时机：

- 在 `review-and-ingestion.md` 完成后编写。
- 在开始规模化整理阵型和视频内容前完成。

## 8. 推荐推进顺序

当前推荐顺序：

1. 完成 `docs/product/mvp-scope.md`。
2. 完成 `docs/product/user-flows.md`。
3. 完成 `docs/product/page-spec.md`。
4. 完成 `docs/product/content-data-policy.md`。
5. 完成 `docs/architecture/data-model.md`。
6. 完成 `docs/project/implementation-roadmap.md`。
7. 完成 `docs/architecture/system-architecture.md`。
8. 按实现顺序完成 `docs/architecture/features/image-search.md`、`layout-library.md`、`video-association.md`、`war-data.md`、`review-and-ingestion.md`。
9. 完成 `docs/api/api-contract.md`。
10. 完成 `docs/qa/acceptance-checklist.md`。
11. 完成 `docs/ops/content-operations.md`。

判断标准：

- 如果还在讨论“做不做某个功能”，优先更新 `mvp-scope.md`。
- 如果还在讨论“用户怎么走”，优先更新 `user-flows.md`。
- 如果还在讨论“页面放什么”，优先更新 `page-spec.md`。
- 如果还在讨论“数据怎么存”，优先更新 `data-model.md`。
- 如果还在讨论“先开发哪块”，优先更新 `implementation-roadmap.md`。
- 如果还在讨论“系统整体怎么搭”，优先更新 `system-architecture.md`。
- 如果还在讨论“某个功能内部怎么实现”，优先更新对应的 `docs/architecture/features/*.md`。
- 如果还在讨论“前后端怎么对接”，优先更新 `api-contract.md`。

## 9. 当前完成情况

截至 2026-06-05：

- 已完成：WarSpark 总版 PRD。
- 已完成：项目文档说明与完成状态。
- 已完成：MVP 范围文档。
- 已完成：用户流程文档。
- 已完成：页面规格文档。
- 已完成：内容与数据边界文档。
- 已完成：数据模型文档。
- 已完成：实施路线图。
- 待编写：系统架构。
- 待编写：功能技术设计文档组。
- 后续补充：API 契约、验收清单、内容运营流程。
