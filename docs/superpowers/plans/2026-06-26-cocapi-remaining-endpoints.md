# CoC API 剩余端点 接入计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `pkg/cocapi` 剩余的 12 个端点全部接入 HTTP API，达到 35/35(100%) 覆盖率。

**Architecture:** 分 3 个模块：Capital（5 端点，新建 domain/cache/service/controller）、War 扩展（2 端点，扩建已有 war domain/service/controller）、Utility（5 端点，新建 domain/cache/service/controller）。所有新端点复用 `cocapiClient`（raw cocapi，不走 adapter）。

**Tech Stack:** Go 1.26, Gin, Redis, cocapi

## Global Constraints

- 代码不加注释，swagger 注释放保留
- 响应格式 `{"code":0,"message":"ok","data":{}}`，用 `utils.OK()` / `utils.Fail()`
- 错误用 `wardomain.Error`，controller 通过 `StatusForCode` 做 HTTP 映射
- `make check` 必须全绿（fmt → vet → test → build）
- 不入库，不涉及 migration
- TTL：联赛类 30min，突袭赛季 5min，战争类 5min，GoldPass 1h，其余不缓存
- tag 统一在 service 层通过 `NormalizeClanTag` 规范化
- 测试共享 `cocapi_fake_test.go`（需追加 12 个新 fake 方法）

---

### Task 1: 扩展 fakeCocapiClient（共享依赖）

**Files:**
- Modify: `internal/service/cocapi_fake_test.go`

**Interfaces:**
- Consumes: cocapi 现有 fake 模式
- Produces: fake 实现供 Task 2-4 测试使用

- [ ] **Step 1: 追加 12 个 fake 方法**

在 `fakeCocapiClient` 结构体上添加以下方法；每个方法返回 `struct{}{}` 空结构体或带一个有效字段的占位数据（测试只验证调用成功，不验证数据内容）：

```go
func (f *fakeCocapiClient) GetBuilderBaseLeagues(ctx context.Context, query cocapi.QueryGetBuilderBaseLeagues) (cocapi.BuilderBaseLeagueListResponse, error) {
	return cocapi.BuilderBaseLeagueListResponse{}, nil
}

func (f *fakeCocapiClient) GetBuilderBaseLeague(ctx context.Context, leagueId string) (cocapi.BuilderBaseLeague, error) {
	return cocapi.BuilderBaseLeague{}, nil
}

func (f *fakeCocapiClient) GetCapitalLeagues(ctx context.Context, query cocapi.QueryGetCapitalLeagues) (cocapi.CapitalLeagueListResponse, error) {
	return cocapi.CapitalLeagueListResponse{}, nil
}

func (f *fakeCocapiClient) GetCapitalLeague(ctx context.Context, leagueId string) (cocapi.CapitalLeague, error) {
	return cocapi.CapitalLeague{}, nil
}

func (f *fakeCocapiClient) SearchClans(ctx context.Context, query cocapi.QuerySearchClans) (cocapi.ClanListResponse, error) {
	return cocapi.ClanListResponse{}, nil
}

func (f *fakeCocapiClient) GetCapitalRaidSeasons(ctx context.Context, clanTag string, query cocapi.QueryGetCapitalRaidSeasons) (cocapi.ClanCapitalRaidSeasonsResponse, error) {
	return cocapi.ClanCapitalRaidSeasonsResponse{}, nil
}

func (f *fakeCocapiClient) GetClanWarLog(ctx context.Context, clanTag string, query cocapi.QueryGetClanWarLog) (cocapi.ClanWarLogResponse, error) {
	return cocapi.ClanWarLogResponse{}, nil
}

func (f *fakeCocapiClient) GetClanWarLeagueWar(ctx context.Context, warTag string) (cocapi.ClanWar, error) {
	return cocapi.ClanWar{}, nil
}

func (f *fakeCocapiClient) GetCurrentGoldPassSeason(ctx context.Context) (cocapi.GoldPassSeason, error) {
	return cocapi.GoldPassSeason{}, nil
}

func (f *fakeCocapiClient) GetLocation(ctx context.Context, locationId string) (cocapi.Location, error) {
	return cocapi.Location{}, nil
}

func (f *fakeCocapiClient) VerifyToken(ctx context.Context, playerTag string, body cocapi.VerifyTokenRequest) (cocapi.VerifyTokenResponse, error) {
	return cocapi.VerifyTokenResponse{}, nil
}

func (f *fakeCocapiClient) GetLeagueGroup(ctx context.Context, leagueGroupTag string, leagueSeasonId string, query cocapi.QueryGetLeagueGroup) (cocapi.LeagueGroup, error) {
	return cocapi.LeagueGroup{}, nil
}
```

### Task 2: War 扩展 — 战争日志 + CWL 单场详情

**Files:**
- Modify: `internal/domain/war/war.go`
- Modify: `internal/infra/redis/war_cache.go`
- Modify: `internal/service/war_service.go`
- Modify: `internal/service/war_service_test.go`
- Modify: `internal/controller/war_controller.go`
- Modify: `internal/controller/war_controller_test.go`

**Interfaces:**
- Consumes: `cocapi.Client` → `warCacheInterface`(扩展)
- Produces: `WarService.GetWarLog(ctx, clanTag, limit, after, before)`, `WarService.GetCWLWar(ctx, warTag)`, 响应透传 cocapi 原始类型

- [ ] **Step 1: 追加 WarLogEntry 类型到 war domain**

`internal/domain/war/war.go` 追加：

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
	BadgeURLs            any     `json:"badgeUrls"`
	ClanLevel            int     `json:"clanLevel"`
	Stars                int     `json:"stars"`
	Tag                  string  `json:"tag"`
	Name                 string  `json:"name"`
	DestructionPercentage float64 `json:"destructionPercentage"`
	Attacks              int     `json:"attacks"`
	ExpEarned            int     `json:"expEarned"`
}
```

- [ ] **Step 2: 扩展 WarCache interface**

在 `internal/service/war_service.go` 的 `WarCache` interface 中追加：

```go
GetWarLog(ctx context.Context, clanTag string) (cocapi.ClanWarLogResponse, bool, error)
SetWarLog(ctx context.Context, clanTag string, resp cocapi.ClanWarLogResponse, ttl time.Duration) error
GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, bool, error)
SetCWLWar(ctx context.Context, warTag string, resp cocapi.ClanWar, ttl time.Duration) error
```

- [ ] **Step 3: 扩展 WarServiceOptions 加 CocapiClient**

在 `internal/service/war_service.go` 的 `WarServiceOptions` struct 追加：

```go
CocapiClient *cocapi.Client
```

在 `WarService` struct 追加字段：

```go
capi *cocapi.Client
```

在 `NewWarService` 函数中赋值：

```go
if option.CocapiClient != nil {
	s.capi = option.CocapiClient
}
```

- [ ] **Step 4: 新增 GetWarLog 方法**

在 `internal/service/war_service.go` 追加：

```go
func (s *WarService) GetWarLog(ctx context.Context, clanTag string, limit int, after, before string) (cocapi.ClanWarLogResponse, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return cocapi.ClanWarLogResponse{}, err
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetWarLog(ctx, normalizedTag)
		if err == nil && ok {
			return resp, nil
		}
	}
	resp, err := s.capi.GetClanWarLog(ctx, normalizedTag, cocapi.QueryGetClanWarLog{
		Limit:  limit,
		After:  after,
		Before: before,
	})
	if err != nil {
		return cocapi.ClanWarLogResponse{}, mapCocapiErrorToWarError(err)
	}
	if s.cache != nil {
		_ = s.cache.SetWarLog(ctx, normalizedTag, resp, s.currentWarCacheTTL)
	}
	return resp, nil
}
```

- [ ] **Step 5: 新增 GetCWLWar 方法**

```go
func (s *WarService) GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, error) {
	normalizedTag, err := NormalizeClanTag(warTag)
	if err != nil {
		return cocapi.ClanWar{}, err
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetCWLWar(ctx, normalizedTag)
		if err == nil && ok {
			return resp, nil
		}
	}
	resp, err := s.capi.GetClanWarLeagueWar(ctx, normalizedTag)
	if err != nil {
		return cocapi.ClanWar{}, mapCocapiErrorToWarError(err)
	}
	if s.cache != nil {
		_ = s.cache.SetCWLWar(ctx, normalizedTag, resp, s.cwlGroupCacheTTL)
	}
	return resp, nil
}
```

- [ ] **Step 6: 追加 mapCocapiErrorToWarError**

`internal/service/war_service.go` 追加。复用 wardomain 错误码：

```go
func mapCocapiErrorToWarError(err error) error {
	if errors.Is(err, cocapi.ErrNotFound) {
		return wardomain.NewError(wardomain.ErrorWarNotFound, err.Error())
	}
	if errors.Is(err, cocapi.ErrInvalidTag) {
		return wardomain.NewError(wardomain.ErrorInvalidTag, err.Error())
	}
	if errors.Is(err, cocapi.ErrAPINotConfigured) {
		return wardomain.NewError(wardomain.ErrorAPINotConfigured, err.Error())
	}
	if errors.Is(err, cocapi.ErrAPIAccessDenied) {
		return wardomain.NewError(wardomain.ErrorAPIAccessDenied, err.Error())
	}
	if errors.Is(err, cocapi.ErrAPIResponseInvalid) {
		return wardomain.NewError(wardomain.ErrorAPIResponseInvalid, err.Error())
	}
	return wardomain.NewError(wardomain.ErrorAPIRequestFailed, err.Error())
}
```

需要在 `war_service.go` 的 imports 中追加 `errors` 和 `cocapi`。

- [ ] **Step 7: 扩展 Redis war_cache 实现**

`internal/infra/redis/war_cache.go` 追加 4 个方法（Get/Set WarLog + Get/Set CWLWar），JSON 序列化/反序列化，key 规则：

```go
func (c *WarCache) GetWarLog(ctx context.Context, clanTag string) (cocapi.ClanWarLogResponse, bool, error)
func (c *WarCache) SetWarLog(ctx context.Context, clanTag string, resp cocapi.ClanWarLogResponse, ttl time.Duration) error
func (c *WarCache) GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, bool, error)
func (c *WarCache) SetCWLWar(ctx context.Context, warTag string, resp cocapi.ClanWar, ttl time.Duration) error
```

Key 格式：`wsp:war:log:{clanTag}`、`wsp:war:cwl_war:{warTag}`

- [ ] **Step 8: 追加 controller handler — GetWarLog + GetCWLWar**

`internal/controller/war_controller.go` 追加：

```go
func (ctl *WarController) GetWarLog(c *gin.Context) {
	tag := c.Param("tag")
	var query struct {
		Limit  int    `form:"limit"`
		After  string `form:"after"`
		Before string `form:"before"`
	}
	if err := utils.BindQuery(c, &query); err != nil {
		return
	}
	resp, err := ctl.service.GetWarLog(c.Request.Context(), tag, query.Limit, query.After, query.Before)
	if err != nil {
		ctl.failWar(c, err)
		return
	}
	utils.OK(c, resp)
}

func (ctl *WarController) GetCWLWar(c *gin.Context) {
	warTag := c.Param("war_tag")
	resp, err := ctl.service.GetCWLWar(c.Request.Context(), warTag)
	if err != nil {
		ctl.failWar(c, err)
		return
	}
	utils.OK(c, resp)
}
```

- [ ] **Step 9: 追加测试**

`internal/service/war_service_test.go` 追加 TestGetWarLog / TestGetCWLWar（覆盖缓存命中、缓存未命中、无效 tag）。
`internal/controller/war_controller_test.go` 追加 TestGetWarLog / TestGetCWLWar（覆盖 200、400、500）。

- [ ] **Step 10: 运行测试验证**

```bash
go test -buildvcs=false -run "TestGetWarLog|TestGetCWLWar" ./internal/service/ ./internal/controller/ -v
```
Expected: PASS

- [ ] **Step 11: Commit**

```bash
git add internal/domain/war/ internal/infra/redis/war_cache.go internal/service/war_service.go internal/service/war_service_test.go internal/controller/war_controller.go internal/controller/war_controller_test.go
git commit -m "feat: add war log + CWL war detail endpoints"
```

---

### Task 3: Capital 模块 — 都城突袭赛季 + 都城联赛 + 建筑大师联赛

**Files:**
- Create: `internal/domain/capital/types.go`
- Create: `internal/infra/redis/capital_cache.go`
- Create: `internal/service/capital_service.go`
- Create: `internal/service/capital_service_test.go`
- Create: `internal/controller/capital_controller.go`
- Create: `internal/controller/capital_controller_test.go`

**Interfaces:**
- Consumes: `*cocapi.Client`, `CapitalCache`
- Produces: `CapitalService` 5 方法, `CapitalController` 5 handler

- [ ] **Step 1: 创建 domain types**

`internal/domain/capital/types.go`：

```go
package capital

import "encoding/json"

type CapitalRaidSeason struct {
	AttackLog               json.RawMessage          `json:"attackLog"`
	CapitalTotalLoot        int                      `json:"capitalTotalLoot"`
	DefenseLog              json.RawMessage          `json:"defenseLog"`
	DefensiveReward         int                      `json:"defensiveReward"`
	EndTime                 string                   `json:"endTime"`
	EnemyDistrictsDestroyed int                      `json:"enemyDistrictsDestroyed"`
	Members                 json.RawMessage          `json:"members"`
	OffensiveReward         int                      `json:"offensiveReward"`
	RaidsCompleted          int                      `json:"raidsCompleted"`
	StartTime               string                   `json:"startTime"`
	State                   string                   `json:"state"`
	TotalAttacks            int                      `json:"totalAttacks"`
}

type CapitalRaidSeasonListResponse struct {
	Items  []CapitalRaidSeason `json:"items"`
	Paging Paging              `json:"paging"`
}

type CapitalLeague struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CapitalLeagueListResponse struct {
	Items  []CapitalLeague `json:"items"`
	Paging Paging          `json:"paging"`
}

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

- [ ] **Step 2: 创建 Redis cache**

`internal/infra/redis/capital_cache.go`。模式同 `label_cache.go`（8 对 Get/Set 方法）。

Key 规则：
```
wsp:capital:raid_seasons:{clanTag}      — CapitalRaidSeason, 5min
wsp:capital:leagues                     — CapitalLeague list, 30min
wsp:capital:league:{id}                 — CapitalLeague detail, 30min
wsp:builder:leagues                     — BuilderBaseLeague list, 30min
wsp:builder:league:{id}                 — BuilderBaseLeague detail, 30min
```

TTL 通过构造函数传入 `time.Duration`（5 对方法公用两个 TTL 值：`raidSeasonTTL` 和 `leagueTTL`）。

- [ ] **Step 3: 创建 CapitalService**

`internal/service/capital_service.go`。模式同 `label_service.go`。

```go
type CapitalService struct {
	api           *cocapi.Client
	cache         *infraredis.CapitalCache
	raidSeasonTTL time.Duration
	leagueTTL     time.Duration
}
```

5 个方法，每个方法：tag 校验 → cache hit check → cocapi call → cache set → return。
错误映射通过 `mapCocapiCapitalError` 函数，逻辑同 `mapCocapiLabelError`。

- [ ] **Step 4: 创建 CapitalService 测试**

`internal/service/capital_service_test.go`。模式同 `label_service_test.go`。
- 5 个测试函数
- 每个测试覆盖：正常调用、缓存命中、无效 tag

- [ ] **Step 5: 创建 CapitalController**

`internal/controller/capital_controller.go`。模式同 `label_controller.go`。

```go
type CapitalController struct {
	service *service.CapitalService
}
```

6 个 handler（5 端点 + failCapital 辅助函数）。

- [ ] **Step 6: 创建 CapitalController 测试**

`internal/controller/capital_controller_test.go`。模式同 `label_controller_test.go`。
- 5 个测试函数
- 每个测试覆盖 200 + 400/500 错误路径

- [ ] **Step 7: 运行测试验证**

```bash
go test -buildvcs=false ./internal/service/ -run "Test.*Capital" -v
go test -buildvcs=false ./internal/controller/ -run "Test.*Capital" -v
```
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/domain/capital/ internal/infra/redis/capital_cache.go internal/service/capital_service.go internal/service/capital_service_test.go internal/controller/capital_controller.go internal/controller/capital_controller_test.go
git commit -m "feat: add capital module - raid seasons, capital leagues, builder base leagues"
```

---

### Task 4: Utility 模块 — GoldPass + SearchClans + Location + VerifyToken + LeagueGroup(stub)

**Files:**
- Create: `internal/domain/utility/types.go`
- Create: `internal/infra/redis/utility_cache.go`
- Create: `internal/service/utility_service.go`
- Create: `internal/service/utility_service_test.go`
- Create: `internal/controller/utility_controller.go`
- Create: `internal/controller/utility_controller_test.go`

**Interfaces:**
- Consumes: `*cocapi.Client`, `UtilityCache`
- Produces: `UtilityService` 5 方法, `UtilityController` 5 handler

- [ ] **Step 1: 创建 domain types**

`internal/domain/utility/types.go`：

```go
package utility

type GoldPassSeason struct {
	EndTime   string `json:"endTime"`
	StartTime string `json:"startTime"`
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

- [ ] **Step 2: 创建 Redis cache（仅 GoldPass 需缓存）**

`internal/infra/redis/utility_cache.go`。仅 1 对 Get/Set 方法，key `wsp:goldpass:current`，TTL 1h。
其余 4 个端点不缓存。

- [ ] **Step 3: 创建 UtilityService**

`internal/service/utility_service.go`：

```go
type UtilityService struct {
	api          *cocapi.Client
	cache        *infraredis.UtilityCache
	goldPassTTL  time.Duration
}
```

5 个方法：
- `GetCurrentGoldPassSeason(ctx)` → cache → cocapi → cache → return domain.utility.GoldPassSeason
- `SearchClans(ctx, params)` → 直接 cocapi → return cocapi.ClanListResponse（透传，不建 domain 类型）
- `GetLocation(ctx, locationID)` → 直接 cocapi → return Location（cocapi struct，不建 domain 类型）
- `VerifyPlayerToken(ctx, playerTag, token)` → 直接 cocapi → return cocapi.VerifyTokenResponse
- `GetPlayerLeagueGroup(ctx, playerTag, seasonID)` → stub → return utils.Fail

错误映射 `mapCocapiUtilityError` 复用 wardomain 错误码。

- [ ] **Step 4: 创建 UtilityService 测试**

`internal/service/utility_service_test.go`：
- TestGetCurrentGoldPassSeason
- TestSearchClans
- TestGetLocation
- TestVerifyPlayerToken
- TestGetPlayerLeagueGroup（stub 验证返回错误）

- [ ] **Step 5: 创建 UtilityController**

`internal/controller/utility_controller.go`：

5 handler：
- `GetCurrentGoldPass` — GET, 无参
- `SearchClans` — GET, BindQuery → ClanSearchParams
- `GetLocation` — GET, `c.Param("id")`
- `VerifyPlayerToken` — POST, `BindJSON` → VerifyTokenRequest
- `GetPlayerLeagueGroup` — GET, 返回 stub 错误

- [ ] **Step 6: 创建 UtilityController 测试**

`internal/controller/utility_controller_test.go`：
- 5 个测试函数，覆盖 200 + 400/500

- [ ] **Step 7: 运行测试验证**

```bash
go test -buildvcs=false ./internal/service/ -run "Test.*[Gg]old|[Ss]earch|[Ll]ocation|[Vv]erify|[Ll]eague[Gg]roup|Utility" -v
go test -buildvcs=false ./internal/controller/ -run "Test.*[Gg]old|[Ss]earch|[Ll]ocation|[Vv]erify|[Ll]eague[Gg]roup|Utility" -v
```
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/domain/utility/ internal/infra/redis/utility_cache.go internal/service/utility_service.go internal/service/utility_service_test.go internal/controller/utility_controller.go internal/controller/utility_controller_test.go
git commit -m "feat: add utility module - gold pass, search, location, verify, league group"
```

---

### Task 5: 路由注册 + make check

**Files:**
- Modify: `internal/api/v1/routes.go`
- Modify: `internal/service/cocapi_fake_test.go`（补充 Task 1 可能遗漏的 import）

- [ ] **Step 1: 注册 Capital 路由**

在 `internal/api/v1/routes.go` 中 capital 路由组，使用已有的 `cocapiClient`：

```go
capitalService := service.NewCapitalService(cocapiClient, infraredis.NewCapitalCache(runtimeConfig.Redis), 5*time.Minute, 30*time.Minute)
capitalController := controller.NewCapitalController(capitalService)

api.GET("/clans/:tag/capital-raid-seasons", capitalController.GetClanCapitalRaidSeasons)
api.GET("/capital-leagues", capitalController.GetCapitalLeagues)
api.GET("/capital-leagues/:id", capitalController.GetCapitalLeague)
api.GET("/builder-base-leagues", capitalController.GetBuilderBaseLeagues)
api.GET("/builder-base-leagues/:id", capitalController.GetBuilderBaseLeague)
```

- [ ] **Step 2: 注册 War 扩展路由**

修改 `WarService` 构造，传入 cocapiClient：

```go
warService := service.NewWarService(warAPIClient, warRepository, service.WarServiceOptions{
	Cache:              warCache,
	CurrentWarCacheTTL: runtimeConfig.Config.CoC.CurrentWarCacheTTL,
	CWLGroupCacheTTL:   runtimeConfig.Config.CoC.CWLGroupCacheTTL,
	CocapiClient:       cocapiClient,
})

api.GET("/clans/:tag/war-log", warController.GetWarLog)
api.GET("/war/cwl/wars/:war_tag", warController.GetCWLWar)
```

> **注意：** 路由 `/war/cwl/wars/:war_tag` 的 `:war_tag` 必须与 `/war/cwl` 不同前缀模式 — `:war_tag` 匹配一层路径，不会与已有 `/war/cwl` 冲突。

- [ ] **Step 3: 注册 Utility 路由**

```go
utilityService := service.NewUtilityService(cocapiClient, infraredis.NewUtilityCache(runtimeConfig.Redis), time.Hour)
utilityController := controller.NewUtilityController(utilityService)

api.GET("/gold-pass/current", utilityController.GetCurrentGoldPass)
api.GET("/clans", utilityController.SearchClans)
api.GET("/locations/:id", utilityController.GetLocation)
api.POST("/players/:tag/verify-token", utilityController.VerifyPlayerToken)
api.GET("/players/:tag/league-group", utilityController.GetPlayerLeagueGroup)
```

- [ ] **Step 4: 补全 import 路径**

确认 routes.go 的 imports 包含新增的 `infraredis` 引用。

- [ ] **Step 5: make check**

```bash
make check
```
Expected: gofmt ok → go vet ok → test PASS → build ok

如果各 Task 1-4 的 commit 已包含全部新文件，routes.go 修改后整体验证。

- [ ] **Step 6: 最终 Commit**

```bash
git add internal/api/v1/routes.go
git commit -m "feat: register 12 remaining cocapi endpoints in v1 routes"
```
