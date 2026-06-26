# CoC API 剩余端点接入设计:都城系统 + 战争扩展 + 工具类

> 日期:2026-06-26
> 模块:github.com/ww1489/WarSpark
> 状态:待实施
> 端点使用:23 → 35 / 35(100%)

## 1. 背景与目标

### 1.1 现状

`pkg/cocapi` 共封装 35 个端点，已接入 23 个（Plan 1: 3 端点 + Plan 2: 17 端点 + 既有 war: 3 端点）。剩余 12 个端点未接入业务层。本设计将此 12 端点全部接入，达到 `pkg/cocapi` 100% 覆盖率。

### 1.2 目标

- 补齐 cocapi 剩余端点的 HTTP API 暴露
- 遵循已有项目模式：domain 类型 → cache → service → controller → routes
- 缓存策略差异化：静态元数据长 TTL，动态数据短 TTL 或不缓存
- `make check` 全绿

### 1.3 范围

涵盖 12 个 cocapi 端点，分 3 个模块：

| 模块 | 端点 | 说明 |
|------|------|------|
| Capital(都城) | 5 | 都城突袭赛季 + 都城联赛 + 建筑大师联赛 |
| War 扩展 | 2 | 部落战日志 + CWL 单场详情 |
| Utility(工具) | 5 | 黄金通行证 + 部落搜索 + 定位详情 + 令牌验证 + 联赛组 |

---

## 2. 模块 1: Capital 都城系统

### 2.1 类型

`internal/domain/capital/types.go`

```go
type CapitalRaidSeason struct {
    AttackLog               CapitalRaidSeasonAttackLog
    CapitalTotalLoot        int
    DefenseLog              CapitalRaidSeasonDefenseLog
    DefensiveReward         int
    EndTime                 string
    EnemyDistrictsDestroyed int
    Members                 CapitalRaidSeasonMembers
    OffensiveReward         int
    RaidsCompleted          int
    StartTime               string
    State                   string
    TotalAttacks            int
}

type CapitalRaidSeasonAttackLog struct {
    Items []CapitalRaidSeasonAttack `json:"items"`
}

type CapitalRaidSeasonAttack struct {
    Attacker      CapitalRaidSeasonAttacker
    DestructionPercentage int
    DistrictCount         int
    Districts             []CapitalRaidSeasonDistrict
    Stars                 int
}

type CapitalRaidSeasonAttacker struct {
    Tag  string `json:"tag"`
    Name string `json:"name"`
}

type CapitalRaidSeasonDistrict struct {
    DestructionPercentage int    `json:"destructionPercentage"`
    DistrictHallLevel     int    `json:"districtHallLevel"`
    ID                    int    `json:"id"`
    Name                  string `json:"name"`
    Stars                 int    `json:"stars"`
}

type CapitalRaidSeasonDefenseLog struct {
    Items []CapitalRaidSeasonDefense
}

type CapitalRaidSeasonDefense struct {
    Defender      CapitalRaidSeasonDefender
    DestructionPercentage int
    DistrictCount         int
    Districts             []CapitalRaidSeasonDistrict
    Stars                 int
}

type CapitalRaidSeasonDefender struct {
    Tag  string `json:"tag"`
    Name string `json:"name"`
}

type CapitalRaidSeasonMembers struct {
    Items []CapitalRaidSeasonMember
}

type CapitalRaidSeasonMember struct {
    Tag                 string `json:"tag"`
    Name                string `json:"name"`
    AttackLimit         int    `json:"attackLimit"`
    Attacks             int    `json:"attacks"`
    BonusAttackLimit    int    `json:"bonusAttackLimit"`
    CapitalResourcesLooted int `json:"capitalResourcesLooted"`
}

type CapitalRaidSeasonListResponse struct {
    Items  []CapitalRaidSeason `json:"items"`
    Paging Paging              `json:"paging"`
}

// ---- 都城联赛 ----

type CapitalLeague struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type CapitalLeagueListResponse struct {
    Items  []CapitalLeague `json:"items"`
    Paging Paging          `json:"paging"`
}

// ---- 建筑大师联赛 ----

type BuilderBaseLeague struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type BuilderBaseLeagueListResponse struct {
    Items  []BuilderBaseLeague `json:"items"`
    Paging Paging              `json:"paging"`
}

type Paging struct {
    Cursors PagingCursors `json:"cursors"`
}

type PagingCursors struct {
    After  string `json:"after,omitempty"`
    Before string `json:"before,omitempty"`
}
```

`CapitalLeague.Name` / `BuilderBaseLeague.Name` 在 cocapi 原始类型中是 `JsonLocalizedName`，domain 层转换为 `string`（调用 `.String()`）。

### 2.2 缓存

`internal/infra/redis/capital_cache.go`

- CapitalRaidSeason: 5min TTL（突袭数据动态变化）
- CapitalLeague: 30min TTL（联赛元数据）
- CapitalLeague detail: 30min TTL
- BuilderBaseLeague: 30min TTL
- BuilderBaseLeague detail: 30min TTL

Get/Set 模式同 `ranking_cache.go`、`league_cache.go`、`label_cache.go`。Key 命名规则：

```
wsp:capital:raid_seasons:{clanTag}
wsp:capital:leagues
wsp:capital:league:{id}
wsp:builder:leagues
wsp:builder:league:{id}
```

### 2.3 Service

`internal/service/capital_service.go`

- `CapitalService` 持有 `*cocapi.Client` + `CapitalCache` + TTL
- 5 个方法，每个方法实现：tag 校验 → cache 查询 → cocapi 调用 → cache 设置 → 返回
- 错误映射 `mapCocapiCapitalError` 复用 `wardomain` 错误码

### 2.4 Controller

`internal/controller/capital_controller.go`

- `CapitalController` 持有 `*service.CapitalService`
- 6 个 handler（含 failCapital 辅助函数）
- 查询参数 `limit/after/before` 通过 `utils.BindQuery` 绑定
- `clanTag` 路径参数直接传原始值（service 层有 NormalizeClanTag）

### 2.5 Endpoints

| HTTP | 路径 | handler | cocapi |
|------|------|---------|--------|
| GET | /api/v1/clans/:tag/capital-raid-seasons | GetClanCapitalRaidSeasons | GetCapitalRaidSeasons |
| GET | /api/v1/capital-leagues | GetCapitalLeagues | GetCapitalLeagues |
| GET | /api/v1/capital-leagues/:id | GetCapitalLeague | GetCapitalLeague |
| GET | /api/v1/builder-base-leagues | GetBuilderBaseLeagues | GetBuilderBaseLeagues |
| GET | /api/v1/builder-base-leagues/:id | GetBuilderBaseLeague | GetBuilderBaseLeague |

---

## 3. 模块 2: War 扩展

### 3.1 策略

扩展现有 `internal/service/war_service.go` 和 `internal/controller/war_controller.go`。不新建 domain 模块，将 WarLogEntry 类型加入 `internal/domain/war/`。

由于 `internal/infra/coc.Client`（adapter）未包装 GetClanWarLog 和 GetClanWarLeagueWar，直接在 `WarService` 中新增字段注入 `*cocapi.Client`，不走 adapter 转换，直接返回 cocapi 原始类型。

### 3.2 类型

`internal/domain/war/war.go` 新增：

```go
type WarLogEntry struct {
    AttacksPerMember int         `json:"attacksPerMember"`
    BattleModifier   string      `json:"battleModifier,omitempty"`
    Clan             WarLogClan  `json:"clan"`
    EndTime          string      `json:"endTime"`
    Opponent         WarLogClan  `json:"opponent"`
    Result           string      `json:"result"`
    TeamSize         int         `json:"teamSize"`
}

type WarLogClan struct {
    BadgeURLs any          `json:"badgeUrls"`
    ClanLevel int          `json:"clanLevel"`
    Stars     int          `json:"stars"`
    Tag       string       `json:"tag"`
    Name      string       `json:"name"`
    DestructionPercentage float64 `json:"destructionPercentage"`
    Attacks   int          `json:"attacks"`
    ExpEarned int          `json:"expEarned"`
}
```

### 3.3 Service

`WarService` 新增字段 `capi *cocapi.Client`。构造函数 `NewWarService` 签名不变（variadic options 已存在，追加字段通过 `WarServiceOptions` 传入）。

新增方法：
- `GetWarLog(ctx, clanTag, limit, after, before)` → 缓存查询 → cocapi 调用 → 缓存设置
- `GetCWLWar(ctx, warTag)` → 缓存查询 → cocapi 调用 → 缓存设置

### 3.4 Controller

`WarController` 新增 handler：
- `GetWarLog` — 解析 `:tag` + 查询参数 `limit/after/before`
- `GetCWLWar` — 解析 `:war_tag`

### 3.5 缓存

扩展 `WarCache` interface + `internal/infra/redis/war_cache.go`：

```
wsp:war:log:{clanTag}
wsp:war:cwl_war:{warTag}
```

TTL: 5min（与现有部落缓存一致）

### 3.6 Endpoints

| HTTP | 路径 | handler | cocapi |
|------|------|---------|--------|
| GET | /api/v1/clans/:tag/war-log | GetWarLog | GetClanWarLog |
| GET | /api/v1/war/cwl/wars/:war_tag | GetCWLWar | GetClanWarLeagueWar |

---

## 4. 模块 3: Utility 工具类

### 4.1 类型

`internal/domain/utility/types.go`

```go
type GoldPassSeason struct {
    EndTime   string `json:"endTime"`
    StartTime string `json:"startTime"`
}

type ClanSearchResult struct {
    Items  []ClanSearchClan `json:"items"`
    Paging Paging           `json:"paging"`
}

type ClanSearchClan struct {
    Tag, Name, Description, Type, WarFrequency  string
    ClanLevel, ClanPoints, ClanBuilderBasePoints, ClanCapitalPoints  int
    Members, RequiredTrophies, RequiredTownhallLevel, RequiredBuilderBaseTrophies  int
    WarWins, WarLosses, WarTies, WarWinStreak  int
    IsFamilyFriendly, IsWarLogPublic  bool
    BadgeURLs                            any
    CapitalLeague                        CapitalLeague
    ChatLanguage                         Language
    ClanCapital                          ClanCapital
    Labels                               LabelList
    Location                             Location
    MemberList                           ClanMemberList
    RequiredLeagueTier                   LeagueTier
    WarLeague                            WarLeague
}

type LocationDetail struct {
    CountryCode    string `json:"countryCode,omitempty"`
    ID             int    `json:"id"`
    IsCountry      bool   `json:"isCountry"`
    LocalizedName  string `json:"localizedName"`
    Name           string `json:"name"`
}

type VerifyTokenRequest struct {
    Token string `json:"token" binding:"required"`
}

type VerifyTokenResponse struct {
    Status string `json:"status"`
    Tag    string `json:"tag"`
    Token  string `json:"token"`
}

type PlayerLeagueGroup struct {
    AttackLogs  []LeagueBattleLogEntry `json:"attackLogs"`
    DefenseLogs []LeagueBattleLogEntry `json:"defenseLogs"`
    Members     []LeagueGroupMember    `json:"members"`
}

type LeagueBattleLogEntry struct {
    CreationTime          string `json:"creationTime"`
    DestructionPercentage int    `json:"destructionPercentage"`
    OpponentName          string `json:"opponentName"`
    OpponentPlayerTag     string `json:"opponentPlayerTag"`
    Stars                 int    `json:"stars"`
    Trophies              int    `json:"trophies"`
}

type LeagueGroupMember struct {
    AttackLoseCount  int    `json:"attackLoseCount"`
    AttackWinCount   int    `json:"attackWinCount"`
    DefenseLoseCount int    `json:"defenseLoseCount"`
    DefenseWinCount  int    `json:"defenseWinCount"`
    LeagueTrophies   int    `json:"leagueTrophies"`
    ClanName         string `json:"clanName"`
    ClanTag          string `json:"clanTag"`
    PlayerName       string `json:"playerName"`
    PlayerTag        string `json:"playerTag"`
}

type ClanSearchParams struct {
    Name          string `form:"name"`
    WarFrequency  string `form:"warFrequency"`
    LocationID    int    `form:"locationId"`
    MinMembers    int    `form:"minMembers"`
    MaxMembers    int    `form:"maxMembers"`
    MinClanPoints int    `form:"minClanPoints"`
    MinClanLevel  int    `form:"minClanLevel"`
    Limit         int    `form:"limit"`
    After         string `form:"after"`
    Before        string `form:"before"`
    LabelIds      string `form:"labelIds"`
}
```

### 4.2 缓存

`internal/infra/redis/utility_cache.go`

| 端点 | 缓存 | TTL |
|------|------|:---:|
| GoldPass | GetGoldPass/SetGoldPass | 1h |
| SearchClans | 不缓存 | — |
| LocationDetail | 不缓存 | — |
| VerifyToken | 不缓存 | — |
| LeagueGroup | 不缓存 | — |

### 4.3 Service

`internal/service/utility_service.go`

- `UtilityService` 持有 `*cocapi.Client` + `UtilityCache` + GoldPass TTL
- GoldPass: cache → cocapi → cache 标准模式
- SearchClans: 参数透传，直接调用 cocapi
- LocationDetail: 直接调用 cocapi
- VerifyToken: POST body 透传，直接调用 cocapi
- LeagueGroup: 参数透传，直接调用 cocapi（需要 playerTag 查询参数 + 路径上的 leagueGroupTag + seasonId）

### 4.4 Controller

`internal/controller/utility_controller.go`

- `UtilityController` 持有 `*service.UtilityService`
- 6 个 handler（含 failUtility 辅助函数）
- `BindJSON` 用于 VerifyToken 的 body 绑定
- `BindQuery` 用于 SearchClans 的查询参数绑定
- LeagueGroup 需要路径参数 `:league_group_tag` + 查询参数 `seasonId` + 可选 `playerTag`

### 4.5 Endpoints

| HTTP | 路径 | handler | cocapi |
|------|------|---------|--------|
| GET | /api/v1/gold-pass/current | GetCurrentGoldPass | GetCurrentGoldPassSeason |
| GET | /api/v1/clans | SearchClans | SearchClans |
| GET | /api/v1/locations/:id | GetLocation | GetLocation |
| POST | /api/v1/players/:tag/verify-token | VerifyPlayerToken | VerifyToken |
| GET | /api/v1/players/:tag/league-group | GetPlayerLeagueGroup | GetLeagueGroup |

> `GET /api/v1/clans/:tag` 已存在（部落详情），`GET /api/v1/clans` 无冲突，Gin 区分精确路径与参数化路径。

### 4.6 LeagueGroup 入参说明

GetLeagueGroup 需要：
- 路径参数 `:tag`（玩家 tag，用于关联上下文）
- 路径参数（leagueGroupTag — 从何处获取？需要决定）

Bug: GetLeagueGroup 需要 `leagueGroupTag` 和 `leagueSeasonId` 两个路径参数 + 可选的 `playerTag` 查询参数。设计为：

`GET /api/v1/players/:tag/league-group?seasonId={seasonId}&playerTag={optional}`

但实际上 cocapi 的 GetLeagueGroup 签名是 `GetLeagueGroup(ctx, leagueGroupTag, leagueSeasonId, query)`。这里的 `leagueGroupTag` 从哪里来？

现有路由中没有可以获取 `leagueGroupTag` 的前置端点。这属于 Plan 3（CWL 情报增强）的前置数据。**该端点暂时标记为 stub** — 返回 `utils.Fail(10007, "league group not available, requires CWL round data")`，待 Plan 3 实现后补齐。

---

## 5. Routes 注册

`internal/api/v1/routes.go` 中：

- Capital + Utility：使用已有的 `cocapiClient`（raw cocapi）
- War 扩展：将 `cocapiClient` 传入 `WarServiceOptions`

```go
// 已有 cocapiClient (raw)
cocapiClient := cocapi.New(cocapi.Config{...})

// Capital
capitalService := service.NewCapitalService(cocapiClient, ...)
capitalController := controller.NewCapitalController(capitalService)
api.GET("/clans/:tag/capital-raid-seasons", capitalController.GetClanCapitalRaidSeasons)
api.GET("/capital-leagues", capitalController.GetCapitalLeagues)
api.GET("/capital-leagues/:id", capitalController.GetCapitalLeague)
api.GET("/builder-base-leagues", capitalController.GetBuilderBaseLeagues)
api.GET("/builder-base-leagues/:id", capitalController.GetBuilderBaseLeague)

// War 扩展 (注入 cocapiClient)
warService := service.NewWarService(warAPIClient, warRepository, service.WarServiceOptions{
    Cache: warCache,
    CocapiClient: cocapiClient,
    ...
})
api.GET("/clans/:tag/war-log", warController.GetWarLog)
api.GET("/war/cwl/wars/:war_tag", warController.GetCWLWar)

// Utility
utilityService := service.NewUtilityService(cocapiClient, ...)
utilityController := controller.NewUtilityController(utilityService)
api.GET("/gold-pass/current", utilityController.GetCurrentGoldPass)
api.GET("/clans", utilityController.SearchClans)
api.GET("/locations/:id", utilityController.GetLocation)
api.POST("/players/:tag/verify-token", utilityController.VerifyPlayerToken)
api.GET("/players/:tag/league-group", utilityController.GetPlayerLeagueGroup)
```

---

## 6. 错误码映射

所有新端点复用 `wardomain` 现有错误码：

| cocapi | wardomain |
|--------|-----------|
| ErrAPINotConfigured | ErrorAPINotConfigured |
| ErrAPIAccessDenied | ErrorAPIAccessDenied |
| ErrNotFound | ErrorWarNotFound → mapErrorWithNotFound |
| ErrInvalidTag | ErrorInvalidTag |
| ErrAPIResponseInvalid | ErrorAPIResponseInvalid |
| ErrAPIRequestFailed | ErrorAPIRequestFailed |
| ErrRateLimited | ErrorAPIRequestFailed |

War 扩展复用 war_controller 已有的 `failWar` 辅助函数。Capital 和 Utility 各自实现 `failCapital` / `failUtility`。

---

## 7. 测试策略

### 7.1 Service 单元测试

使用共享的 `cocapi_fake_test.go`（已有 `fakeCocapiClient` 覆盖 17 个方法，需要追加 12 个新方法）。

每个 service 文件对应测试文件：
- `capital_service_test.go`
- `war_service_ext_test.go`（追加到已有 war_service_test.go）
- `utility_service_test.go`

### 7.2 Controller 单元测试

- `capital_controller_test.go`
- `war_controller_ext_test.go`（追加）
- `utility_controller_test.go`

### 7.3 验证

`make check` 全绿（gofmt → go vet → go test → go build）。

---

## 8. 不做事项

- 不新建数据库表（无 migration）
- 不修改 infracoc adapter（新端点直接走 raw cocapi）
- LeagueGroup 端点返回 stub 错误（依赖 Plan 3 CWL 数据）
- 不修改已有测试

---

## 9. 文件清单

### 新建文件

| 文件 | 模块 |
|------|------|
| `internal/domain/capital/types.go` | Capital |
| `internal/infra/redis/capital_cache.go` | Capital |
| `internal/service/capital_service.go` | Capital |
| `internal/service/capital_service_test.go` | Capital |
| `internal/controller/capital_controller.go` | Capital |
| `internal/controller/capital_controller_test.go` | Capital |
| `internal/domain/utility/types.go` | Utility |
| `internal/infra/redis/utility_cache.go` | Utility |
| `internal/service/utility_service.go` | Utility |
| `internal/service/utility_service_test.go` | Utility |
| `internal/controller/utility_controller.go` | Utility |
| `internal/controller/utility_controller_test.go` | Utility |

### 修改文件

| 文件 | 变更 |
|------|------|
| `internal/domain/war/war.go` | 追加 WarLogEntry + WarLogClan 类型 |
| `internal/infra/redis/war_cache.go` | 追加 GetWarLog/SetWarLog + GetCWLWar/SetCWLWar |
| `internal/service/war_service.go` | 追加 CocapiClient 字段 + GetWarLog/GetCWLWar 方法 |
| `internal/service/war_service_test.go` | 追加测试 |
| `internal/controller/war_controller.go` | 追加 GetWarLog/GetCWLWar handler |
| `internal/controller/war_controller_test.go` | 追加测试 |
| `internal/service/cocapi_fake_test.go` | 追加 12 个新 fake 方法 |
| `internal/api/v1/routes.go` | 注册 12 条新路由 |
