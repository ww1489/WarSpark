# 排名榜 + 联赛元数据 + 标签 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 17 个 HTTP 接口覆盖排名榜(6)、联赛元数据(9)、标签(2)三大模块，接入 cocapi 21 个端点。纯缓存、不入库、不走 adapter。

**Architecture:** 3 个独立 service 直接持有 `*cocapi.Client`，无 adapter 层。Service 内部做 cocapi→domain 类型转换。Redis 缓存按模块隔离。Controller 标准模式。所有 service 共享 `mapCocapiRankingError` 辅助函数进行错误映射。

**Tech Stack:** Go 1.26, Gin, Redis (go-redis), cocapi (本仓库 pkg/cocapi), 项目的 utils.OK/Fail 响应模式。

## Global Constraints

- 使用中文回复(AGENTS.md 规则)。
- 代码不加注释(AGENTS.md 规则)，但 Swagger 注释放保留。
- 响应格式:`{"code": 0, "message": "ok", "data": {}}`，用 `utils.OK()` / `utils.Fail()`。
- 错误用 `wardomain.Error` 带稳定 code，controller 通过 `StatusForCode` 做 HTTP 映射。
- `make check` 必须全绿(fmt + vet + test + build)。
- 不走 adapter(纯透传),service 直接持有 `*cocapi.Client`。
- 排名榜 Redis TTL = 10min, 联赛 TTL = 30min, 标签 TTL = 1h。
- 不入库，不涉及 migration。
- cocapi 所有分页查询参数统一:`Limit int` / `After string` / `Before string`(omitempty)。
- 新错误码加到 `internal/domain/war/war.go`。
- 每个 service 内部用 `mapCocapiRankingError(err, notFoundCode)` 将 cocapi sentinel errors 映射为 `wardomain.Error`。

## Spec 偏离说明

设计文档(`docs/superpowers/specs/2026-06-26-cocapi-features-design.md`):
1. ❌ `GetBuilderBaseLeagues` / `GetBuilderBaseLeague` 不暴露(对战争情报无价值,后续需要再加)。
2. ❌ `GetLeagueGroup`(传奇联赛分组)不暴露(冗余,战争情报页用 CWL 分组即可)。
3. ❌ `GetLocation` 单地点详情不暴露(列表已含全部字段)。
4. 合计偏差:3 端点不暴露,实际新增 18 / 21 个端点。接口数 17 不变。

---

### Task 1:Domain 类型 + 错误码扩展

**Files:**
- Create: `internal/domain/ranking/types.go`
- Create: `internal/domain/league/types.go`
- Create: `internal/domain/label/types.go`
- Modify: `internal/domain/war/war.go:8-18`(错误码常量块)

**Interfaces:**
- Produces: 3 个 domain 包的类型定义;`wardomain.ErrorLeagueNotFound`、`wardomain.ErrorLocationNotFound` 常量。

- [ ] **Step 1: 扩展 war.go 错误码**

在 `internal/domain/war/war.go` 的 `ErrorPlayerNotFound` 行之后追加:

```go
	ErrorLeagueNotFound    = "league_not_found"
	ErrorLocationNotFound  = "location_not_found"
```

- [ ] **Step 2: 创建 ranking domain 包**

创建 `internal/domain/ranking/types.go`:

```go
package ranking

type Paging struct {
	Cursors Cursors `json:"paging,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

type Location struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	CountryCode   string `json:"countryCode,omitempty"`
	IsCountry     bool   `json:"isCountry,omitempty"`
	LocalizedName string `json:"localizedName,omitempty"`
}

type LocationListResponse struct {
	Items  []Location `json:"items"`
	Paging Paging     `json:"paging"`
}

type ClanRankingEntry struct {
	Tag          string    `json:"tag"`
	Name         string    `json:"name"`
	ClanLevel    int       `json:"clanLevel"`
	ClanPoints   int       `json:"clanPoints"`
	Members      int       `json:"members"`
	Rank         int       `json:"rank"`
	PreviousRank int       `json:"previousRank,omitempty"`
	BadgeURLs    any       `json:"badgeUrls,omitempty"`
	Location     *Location `json:"location,omitempty"`
}

type ClanRankingListResponse struct {
	Items  []ClanRankingEntry `json:"items"`
	Paging Paging             `json:"paging"`
}

type PlayerRankingEntry struct {
	Tag          string   `json:"tag"`
	Name         string   `json:"name"`
	ExpLevel     int      `json:"expLevel"`
	Trophies     int      `json:"trophies"`
	Rank         int      `json:"rank"`
	PreviousRank int      `json:"previousRank,omitempty"`
	AttackWins   int      `json:"attackWins,omitempty"`
	DefenseWins  int      `json:"defenseWins,omitempty"`
	Clan         *ClanRef `json:"clan,omitempty"`
}

type ClanRef struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	BadgeURLs any    `json:"badgeUrls,omitempty"`
}

type PlayerRankingListResponse struct {
	Items  []PlayerRankingEntry `json:"items"`
	Paging Paging               `json:"paging"`
}

type ClanCapitalRankingEntry struct {
	Tag               string    `json:"tag"`
	Name              string    `json:"name"`
	ClanLevel         int       `json:"clanLevel"`
	ClanCapitalPoints int       `json:"clanCapitalPoints"`
	Rank              int       `json:"rank"`
	PreviousRank      int       `json:"previousRank,omitempty"`
	Members           int       `json:"members"`
	BadgeURLs         any       `json:"badgeUrls,omitempty"`
	Location          *Location `json:"location,omitempty"`
}

type ClanCapitalRankingListResponse struct {
	Items  []ClanCapitalRankingEntry `json:"items"`
	Paging Paging                    `json:"paging"`
}

type ClanBuilderBaseRankingEntry struct {
	Tag                   string    `json:"tag"`
	Name                  string    `json:"name"`
	ClanLevel             int       `json:"clanLevel"`
	ClanBuilderBasePoints int       `json:"clanBuilderBasePoints"`
	Rank                  int       `json:"rank"`
	PreviousRank          int       `json:"previousRank,omitempty"`
	Members               int       `json:"members"`
	BadgeURLs             any       `json:"badgeUrls,omitempty"`
	Location              *Location `json:"location,omitempty"`
}

type ClanBuilderBaseRankingListResponse struct {
	Items  []ClanBuilderBaseRankingEntry `json:"items"`
	Paging Paging                        `json:"paging"`
}

type PlayerBuilderBaseRankingEntry struct {
	Tag                 string   `json:"tag"`
	Name                string   `json:"name"`
	ExpLevel            int      `json:"expLevel"`
	BuilderBaseTrophies int      `json:"builderBaseTrophies"`
	Rank                int      `json:"rank"`
	PreviousRank        int      `json:"previousRank,omitempty"`
	Clan                *ClanRef `json:"clan,omitempty"`
}

type PlayerBuilderBaseRankingListResponse struct {
	Items  []PlayerBuilderBaseRankingEntry `json:"items"`
	Paging Paging                          `json:"paging"`
}
```

- [ ] **Step 3: 创建 league domain 包**

创建 `internal/domain/league/types.go`:

```go
package league

type Paging struct {
	Cursors Cursors `json:"paging,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

type League struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconURLs any    `json:"iconUrls,omitempty"`
}

type LeagueListResponse struct {
	Items  []League `json:"items"`
	Paging Paging   `json:"paging"`
}

type LeagueSeason struct {
	ID string `json:"id"`
}

type LeagueSeasonListResponse struct {
	Items  []LeagueSeason `json:"items"`
	Paging Paging         `json:"paging"`
}

type LeagueSeasonRankingEntry struct {
	Tag          string   `json:"tag"`
	Name         string   `json:"name"`
	ExpLevel     int      `json:"expLevel"`
	Trophies     int      `json:"trophies"`
	Rank         int      `json:"rank"`
	PreviousRank int      `json:"previousRank,omitempty"`
	AttackWins   int      `json:"attackWins,omitempty"`
	DefenseWins  int      `json:"defenseWins,omitempty"`
	Clan         *ClanRef `json:"clan,omitempty"`
}

type ClanRef struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	BadgeURLs any    `json:"badgeUrls,omitempty"`
}

type LeagueSeasonRankingListResponse struct {
	Items  []LeagueSeasonRankingEntry `json:"items"`
	Paging Paging                     `json:"paging"`
}

type LeagueTier struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconURLs any    `json:"iconUrls,omitempty"`
}

type LeagueTierListResponse struct {
	Items  []LeagueTier `json:"items"`
	Paging Paging       `json:"paging"`
}

type WarLeague struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WarLeagueListResponse struct {
	Items  []WarLeague `json:"items"`
	Paging Paging      `json:"paging"`
}

type LeagueSeasonResult struct {
	LeagueSeasonID int64 `json:"leagueSeasonId"`
	LeagueTierID   int   `json:"leagueTierId,omitempty"`
	Trophies       int   `json:"trophies"`
	Placement      int   `json:"placement,omitempty"`
	AttackWins     int   `json:"attackWins,omitempty"`
	AttackLosses   int   `json:"attackLosses,omitempty"`
	AttackStars    int   `json:"attackStars,omitempty"`
	DefenseWins    int   `json:"defenseWins,omitempty"`
	DefenseLosses  int   `json:"defenseLosses,omitempty"`
	DefenseStars   int   `json:"defenseStars,omitempty"`
	MaxBattles     int   `json:"maxBattles,omitempty"`
}

type LeagueSeasonResultListResponse struct {
	Items  []LeagueSeasonResult `json:"items"`
	Paging Paging               `json:"paging"`
}
```

- [ ] **Step 4: 创建 label domain 包**

创建 `internal/domain/label/types.go`:

```go
package label

type Paging struct {
	Cursors Cursors `json:"paging,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

type Label struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconURLs any    `json:"iconUrls,omitempty"`
}

type LabelListResponse struct {
	Items  []Label `json:"items"`
	Paging Paging  `json:"paging"`
}
```

- [ ] **Step 5: 编译验证**

Run: `go build ./internal/domain/...`
Expected: 无输出(通过)

- [ ] **Step 6: Commit**

```bash
git add internal/domain/ranking/types.go internal/domain/league/types.go internal/domain/label/types.go internal/domain/war/war.go
git commit -m "Add domain types for ranking, league, label modules + error codes"
```

---

### Task 2:Redis 缓存 — 3 个 Cache 文件

**Files:**
- Create: `internal/infra/redis/ranking_cache.go`
- Create: `internal/infra/redis/league_cache.go`
- Create: `internal/infra/redis/label_cache.go`

**Interfaces:**
- Produces: `RankingCache`、`LeagueCache`、`LabelCache` struct + 所有 Get/Set 方法

- [ ] **Step 1: 创建 ranking_cache.go**

创建 `internal/infra/redis/ranking_cache.go`:

```go
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
)

type RankingCache struct {
	client *goredis.Client
}

func NewRankingCache(client *goredis.Client) *RankingCache {
	return &RankingCache{client: client}
}

func (c *RankingCache) GetLocations(ctx context.Context) (ranking.LocationListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.LocationListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:locations").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return ranking.LocationListResponse{}, false, nil }
		return ranking.LocationListResponse{}, false, err
	}
	var resp ranking.LocationListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.LocationListResponse{}, false, err }
	return resp, true, nil
}

func (c *RankingCache) SetLocations(ctx context.Context, resp ranking.LocationListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "ranking:locations", data, ttl).Err()
}

func (c *RankingCache) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.ClanRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:clan:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return ranking.ClanRankingListResponse{}, false, nil }
		return ranking.ClanRankingListResponse{}, false, err
	}
	var resp ranking.ClanRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.ClanRankingListResponse{}, false, err }
	return resp, true, nil
}

func (c *RankingCache) SetClanRanking(ctx context.Context, locationID string, resp ranking.ClanRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "ranking:clan:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.PlayerRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:player:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return ranking.PlayerRankingListResponse{}, false, nil }
		return ranking.PlayerRankingListResponse{}, false, err
	}
	var resp ranking.PlayerRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.PlayerRankingListResponse{}, false, err }
	return resp, true, nil
}

func (c *RankingCache) SetPlayerRanking(ctx context.Context, locationID string, resp ranking.PlayerRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "ranking:player:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.ClanCapitalRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:capital:"+locationID).Bytes()
	if err != nil { return checkRedisErr[ranking.ClanCapitalRankingListResponse](data, err) }
	var resp ranking.ClanCapitalRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.ClanCapitalRankingListResponse{}, false, err }
	return resp, true, nil
}

func (c *RankingCache) SetClanCapitalRanking(ctx context.Context, locationID string, resp ranking.ClanCapitalRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "ranking:capital:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.ClanBuilderBaseRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:builder-clan:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return ranking.ClanBuilderBaseRankingListResponse{}, false, nil }
		return ranking.ClanBuilderBaseRankingListResponse{}, false, err
	}
	var resp ranking.ClanBuilderBaseRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.ClanBuilderBaseRankingListResponse{}, false, err }
	return resp, true, nil
}

func (c *RankingCache) SetClanBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.ClanBuilderBaseRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "ranking:builder-clan:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.PlayerBuilderBaseRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:builder-player:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return ranking.PlayerBuilderBaseRankingListResponse{}, false, nil }
		return ranking.PlayerBuilderBaseRankingListResponse{}, false, err
	}
	var resp ranking.PlayerBuilderBaseRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.PlayerBuilderBaseRankingListResponse{}, false, err }
	return resp, true, nil
}

func (c *RankingCache) SetPlayerBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.PlayerBuilderBaseRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "ranking:builder-player:"+locationID, data, ttl).Err()
}
```

**注意:** `GetClanCapitalRanking` 中有一个函数 `checkRedisErr` 不存在，请修正为:

```go
func (c *RankingCache) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return ranking.ClanCapitalRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "ranking:capital:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return ranking.ClanCapitalRankingListResponse{}, false, nil }
		return ranking.ClanCapitalRankingListResponse{}, false, err
	}
	var resp ranking.ClanCapitalRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return ranking.ClanCapitalRankingListResponse{}, false, err }
	return resp, true, nil
}
```

- [ ] **Step 2: 创建 league_cache.go**

创建 `internal/infra/redis/league_cache.go`:

```go
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/league"
)

type LeagueCache struct {
	client *goredis.Client
}

func NewLeagueCache(client *goredis.Client) *LeagueCache {
	return &LeagueCache{client: client}
}

func (c *LeagueCache) GetLeagues(ctx context.Context) (league.LeagueListResponse, bool, error) {
	if c == nil || c.client == nil { return league.LeagueListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "league:list").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.LeagueListResponse{}, false, nil }
		return league.LeagueListResponse{}, false, err
	}
	var resp league.LeagueListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return league.LeagueListResponse{}, false, err }
	return resp, true, nil
}

func (c *LeagueCache) SetLeagues(ctx context.Context, resp league.LeagueListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "league:list", data, ttl).Err()
}

func (c *LeagueCache) GetLeague(ctx context.Context, id string) (league.League, bool, error) {
	if c == nil || c.client == nil { return league.League{}, false, nil }
	data, err := c.client.Get(ctx, "league:"+id).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.League{}, false, nil }
		return league.League{}, false, err
	}
	var l league.League
	if err := json.Unmarshal(data, &l); err != nil { return league.League{}, false, err }
	return l, true, nil
}

func (c *LeagueCache) SetLeague(ctx context.Context, id string, l league.League, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(l)
	if err != nil { return err }
	return c.client.Set(ctx, "league:"+id, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, bool, error) {
	if c == nil || c.client == nil { return league.LeagueSeasonListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "league:seasons:"+leagueID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.LeagueSeasonListResponse{}, false, nil }
		return league.LeagueSeasonListResponse{}, false, err
	}
	var resp league.LeagueSeasonListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return league.LeagueSeasonListResponse{}, false, err }
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueSeasons(ctx context.Context, leagueID string, resp league.LeagueSeasonListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "league:seasons:"+leagueID, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, bool, error) {
	if c == nil || c.client == nil { return league.LeagueSeasonRankingListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "league:rankings:"+leagueID+":"+season).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.LeagueSeasonRankingListResponse{}, false, nil }
		return league.LeagueSeasonRankingListResponse{}, false, err
	}
	var resp league.LeagueSeasonRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return league.LeagueSeasonRankingListResponse{}, false, err }
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueSeasonRankings(ctx context.Context, leagueID, season string, resp league.LeagueSeasonRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "league:rankings:"+leagueID+":"+season, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, bool, error) {
	if c == nil || c.client == nil { return league.LeagueTierListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "league:tiers:"+leagueID+":"+season).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.LeagueTierListResponse{}, false, nil }
		return league.LeagueTierListResponse{}, false, err
	}
	var resp league.LeagueTierListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return league.LeagueTierListResponse{}, false, err }
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueTiers(ctx context.Context, leagueID, season string, resp league.LeagueTierListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "league:tiers:"+leagueID+":"+season, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, bool, error) {
	if c == nil || c.client == nil { return league.LeagueTier{}, false, nil }
	data, err := c.client.Get(ctx, "league:tier:"+tierID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.LeagueTier{}, false, nil }
		return league.LeagueTier{}, false, err
	}
	var t league.LeagueTier
	if err := json.Unmarshal(data, &t); err != nil { return league.LeagueTier{}, false, err }
	return t, true, nil
}

func (c *LeagueCache) SetLeagueTier(ctx context.Context, tierID string, t league.LeagueTier, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(t)
	if err != nil { return err }
	return c.client.Set(ctx, "league:tier:"+tierID, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, bool, error) {
	if c == nil || c.client == nil { return league.LeagueSeasonResultListResponse{}, false, nil }
	tag := normalizePlayerTagForCache(playerTag)
	data, err := c.client.Get(ctx, "league:history:"+tag).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.LeagueSeasonResultListResponse{}, false, nil }
		return league.LeagueSeasonResultListResponse{}, false, err
	}
	var resp league.LeagueSeasonResultListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return league.LeagueSeasonResultListResponse{}, false, err }
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueHistory(ctx context.Context, playerTag string, resp league.LeagueSeasonResultListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	tag := normalizePlayerTagForCache(playerTag)
	return c.client.Set(ctx, "league:history:"+tag, data, ttl).Err()
}

func (c *LeagueCache) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, bool, error) {
	if c == nil || c.client == nil { return league.WarLeagueListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "league:war-list").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.WarLeagueListResponse{}, false, nil }
		return league.WarLeagueListResponse{}, false, err
	}
	var resp league.WarLeagueListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return league.WarLeagueListResponse{}, false, err }
	return resp, true, nil
}

func (c *LeagueCache) SetWarLeagues(ctx context.Context, resp league.WarLeagueListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "league:war-list", data, ttl).Err()
}

func (c *LeagueCache) GetWarLeague(ctx context.Context, id string) (league.WarLeague, bool, error) {
	if c == nil || c.client == nil { return league.WarLeague{}, false, nil }
	data, err := c.client.Get(ctx, "league:war:"+id).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return league.WarLeague{}, false, nil }
		return league.WarLeague{}, false, err
	}
	var l league.WarLeague
	if err := json.Unmarshal(data, &l); err != nil { return league.WarLeague{}, false, err }
	return l, true, nil
}

func (c *LeagueCache) SetWarLeague(ctx context.Context, id string, l league.WarLeague, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(l)
	if err != nil { return err }
	return c.client.Set(ctx, "league:war:"+id, data, ttl).Err()
}

func normalizePlayerTagForCache(tag string) string {
	tag = strings.ToUpper(strings.TrimSpace(tag))
	tag = strings.TrimPrefix(tag, "#")
	return tag
}
```

**注意:** 需要导入 `"strings"`。在 import 块中追加。

- [ ] **Step 3: 创建 label_cache.go**

创建 `internal/infra/redis/label_cache.go`:

```go
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/label"
)

type LabelCache struct {
	client *goredis.Client
}

func NewLabelCache(client *goredis.Client) *LabelCache {
	return &LabelCache{client: client}
}

func (c *LabelCache) GetClanLabels(ctx context.Context) (label.LabelListResponse, bool, error) {
	if c == nil || c.client == nil { return label.LabelListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "label:clan").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return label.LabelListResponse{}, false, nil }
		return label.LabelListResponse{}, false, err
	}
	var resp label.LabelListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return label.LabelListResponse{}, false, err }
	return resp, true, nil
}

func (c *LabelCache) SetClanLabels(ctx context.Context, resp label.LabelListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "label:clan", data, ttl).Err()
}

func (c *LabelCache) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, bool, error) {
	if c == nil || c.client == nil { return label.LabelListResponse{}, false, nil }
	data, err := c.client.Get(ctx, "label:player").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) { return label.LabelListResponse{}, false, nil }
		return label.LabelListResponse{}, false, err
	}
	var resp label.LabelListResponse
	if err := json.Unmarshal(data, &resp); err != nil { return label.LabelListResponse{}, false, err }
	return resp, true, nil
}

func (c *LabelCache) SetPlayerLabels(ctx context.Context, resp label.LabelListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil { return nil }
	data, err := json.Marshal(resp)
	if err != nil { return err }
	return c.client.Set(ctx, "label:player", data, ttl).Err()
}
```

- [ ] **Step 4: 编译验证**

Run: `go build ./internal/infra/redis/...`
Expected: 无输出(通过)

- [ ] **Step 5: Commit**

```bash
git add internal/infra/redis/ranking_cache.go internal/infra/redis/league_cache.go internal/infra/redis/label_cache.go
git commit -m "Add Redis caches for ranking, league, label modules"
```

---

### Task 3:RankingService + 测试

**Files:**
- Create: `internal/service/ranking_service.go`
- Create: `internal/service/ranking_service_test.go`
- Create: `internal/service/cocapi_fake_test.go`

- [ ] **Step 1: 创建 fakeCocapiClient(供所有 Plan 2 service 测试共用)**

创建 `internal/service/cocapi_fake_test.go`:

```go
package service

import (
	"context"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeCocapiClient struct {
	locations       cocapi.LocationListResponse
	clanRanking     cocapi.ClanRankingListResponse
	playerRanking   cocapi.PlayerRankingListResponse
	capitalRanking  cocapi.ClanCapitalRankingListResponse
	builderClan     cocapi.ClanBuilderBaseRankingListResponse
	builderPlayer   cocapi.PlayerBuilderBaseRankingListResponse
	leagues         cocapi.LeagueListResponse
	league          cocapi.League
	leagueSeasons   cocapi.LeagueSeasonListResponse
	leagueRankings  cocapi.PlayerRankingListResponse
	leagueTiers     cocapi.LeagueTierListResponse
	leagueTier      cocapi.LeagueTier
	leagueHistory   cocapi.LeagueSeasonResultListResponse
	warLeagues      cocapi.WarLeagueListResponse
	warLeague       cocapi.WarLeague
	clanLabels      cocapi.LabelListResponse
	playerLabels    cocapi.LabelListResponse
	err             error
}

func (f *fakeCocapiClient) GetLocations(ctx context.Context, query cocapi.QueryGetLocations) (cocapi.LocationListResponse, error) {
	return f.locations, f.err
}
func (f *fakeCocapiClient) GetClanRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanRanking) (cocapi.ClanRankingListResponse, error) {
	return f.clanRanking, f.err
}
func (f *fakeCocapiClient) GetPlayerRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerRanking) (cocapi.PlayerRankingListResponse, error) {
	return f.playerRanking, f.err
}
func (f *fakeCocapiClient) GetClanCapitalRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanCapitalRanking) (cocapi.ClanCapitalRankingListResponse, error) {
	return f.capitalRanking, f.err
}
func (f *fakeCocapiClient) GetClanBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanBuilderBaseRanking) (cocapi.ClanBuilderBaseRankingListResponse, error) {
	return f.builderClan, f.err
}
func (f *fakeCocapiClient) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerBuilderBaseRanking) (cocapi.PlayerBuilderBaseRankingListResponse, error) {
	return f.builderPlayer, f.err
}
func (f *fakeCocapiClient) GetLeagues(ctx context.Context, query cocapi.QueryGetLeagues) (cocapi.LeagueListResponse, error) {
	return f.leagues, f.err
}
func (f *fakeCocapiClient) GetLeague(ctx context.Context, leagueID string) (cocapi.League, error) {
	return f.league, f.err
}
func (f *fakeCocapiClient) GetLeagueSeasons(ctx context.Context, leagueID string, query cocapi.QueryGetLeagueSeasons) (cocapi.LeagueSeasonListResponse, error) {
	return f.leagueSeasons, f.err
}
func (f *fakeCocapiClient) GetLeagueSeasonRankings(ctx context.Context, leagueID, seasonID string, query cocapi.QueryGetLeagueSeasonRankings) (cocapi.PlayerRankingListResponse, error) {
	return f.leagueRankings, f.err
}
func (f *fakeCocapiClient) GetLeagueTiers(ctx context.Context, query cocapi.QueryGetLeagueTiers) (cocapi.LeagueTierListResponse, error) {
	return f.leagueTiers, f.err
}
func (f *fakeCocapiClient) GetLeagueTier(ctx context.Context, tierID string) (cocapi.LeagueTier, error) {
	return f.leagueTier, f.err
}
func (f *fakeCocapiClient) GetLeagueHistory(ctx context.Context, playerTag string) (cocapi.LeagueSeasonResultListResponse, error) {
	return f.leagueHistory, f.err
}
func (f *fakeCocapiClient) GetWarLeagues(ctx context.Context, query cocapi.QueryGetWarLeagues) (cocapi.WarLeagueListResponse, error) {
	return f.warLeagues, f.err
}
func (f *fakeCocapiClient) GetWarLeague(ctx context.Context, leagueID string) (cocapi.WarLeague, error) {
	return f.warLeague, f.err
}
func (f *fakeCocapiClient) GetClanLabels(ctx context.Context, query cocapi.QueryGetClanLabels) (cocapi.LabelListResponse, error) {
	return f.clanLabels, f.err
}
func (f *fakeCocapiClient) GetPlayerLabels(ctx context.Context, query cocapi.QueryGetPlayerLabels) (cocapi.LabelListResponse, error) {
	return f.playerLabels, f.err
}
```

- [ ] **Step 2: 写 RankingService 测试**

创建 `internal/service/ranking_service_test.go`:

```go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeRankingCache struct {
	locations       ranking.LocationListResponse
	locationsHit    bool
	clanRanking     ranking.ClanRankingListResponse
	clanRankingHit  bool
	playerRanking   ranking.PlayerRankingListResponse
	playerRankingHit bool
}

func (f *fakeRankingCache) GetLocations(ctx context.Context) (ranking.LocationListResponse, bool, error) {
	return f.locations, f.locationsHit, nil
}
func (f *fakeRankingCache) SetLocations(ctx context.Context, resp ranking.LocationListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, bool, error) {
	return f.clanRanking, f.clanRankingHit, nil
}
func (f *fakeRankingCache) SetClanRanking(ctx context.Context, locationID string, resp ranking.ClanRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, bool, error) {
	return f.playerRanking, f.playerRankingHit, nil
}
func (f *fakeRankingCache) SetPlayerRanking(ctx context.Context, locationID string, resp ranking.PlayerRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error) {
	return ranking.ClanCapitalRankingListResponse{}, false, nil
}
func (f *fakeRankingCache) SetClanCapitalRanking(ctx context.Context, locationID string, resp ranking.ClanCapitalRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, bool, error) {
	return ranking.ClanBuilderBaseRankingListResponse{}, false, nil
}
func (f *fakeRankingCache) SetClanBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.ClanBuilderBaseRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, bool, error) {
	return ranking.PlayerBuilderBaseRankingListResponse{}, false, nil
}
func (f *fakeRankingCache) SetPlayerBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.PlayerBuilderBaseRankingListResponse, ttl time.Duration) error { return nil }

func TestRankingServiceGetLocationsCached(t *testing.T) {
	cached := ranking.LocationListResponse{Items: []ranking.Location{{ID: 1, Name: "Global"}}}
	svc := NewRankingService(&fakeCocapiClient{}, &fakeRankingCache{locations: cached, locationsHit: true}, 10*time.Minute)
	resp, err := svc.GetLocations(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Global" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceGetLocationsFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		locations: cocapi.LocationListResponse{Items: []cocapi.Location{{ID: 1, Name: "Global"}}},
	}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	resp, err := svc.GetLocations(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Global" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceGetClanRanking(t *testing.T) {
	fakeAPI := &fakeCocapiClient{clanRanking: cocapi.ClanRankingListResponse{Items: []cocapi.ClanRanking{{Name: "TopClan", Rank: 1}}}}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	resp, err := svc.GetClanRanking(context.Background(), "global")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "TopClan" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceGetPlayerRanking(t *testing.T) {
	fakeAPI := &fakeCocapiClient{playerRanking: cocapi.PlayerRankingListResponse{Items: []cocapi.PlayerRanking{{Name: "TopPlayer", Rank: 1}}}}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	resp, err := svc.GetPlayerRanking(context.Background(), "global")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "TopPlayer" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceNotFound(t *testing.T) {
	fakeAPI := &fakeCocapiClient{err: cocapi.ErrNotFound}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	_, err := svc.GetClanRanking(context.Background(), "999")
	if err == nil { t.Fatal("expected error") }
	if wardomain.ErrorCode(err) != wardomain.ErrorLocationNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorLocationNotFound)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/service/ -run TestRankingService -v`
Expected: FAIL — `NewRankingService undefined`

- [ ] **Step 4: 实现 RankingService**

创建 `internal/service/ranking_service.go`:

```go
package service

import (
	"context"
	"errors"
	"time"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type rankingCache interface {
	GetLocations(ctx context.Context) (ranking.LocationListResponse, bool, error)
	SetLocations(ctx context.Context, resp ranking.LocationListResponse, ttl time.Duration) error
	GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, bool, error)
	SetClanRanking(ctx context.Context, locationID string, resp ranking.ClanRankingListResponse, ttl time.Duration) error
	GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, bool, error)
	SetPlayerRanking(ctx context.Context, locationID string, resp ranking.PlayerRankingListResponse, ttl time.Duration) error
	GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error)
	SetClanCapitalRanking(ctx context.Context, locationID string, resp ranking.ClanCapitalRankingListResponse, ttl time.Duration) error
	GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, bool, error)
	SetClanBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.ClanBuilderBaseRankingListResponse, ttl time.Duration) error
	GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, bool, error)
	SetPlayerBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.PlayerBuilderBaseRankingListResponse, ttl time.Duration) error
}

type cocapiRankingClient interface {
	GetLocations(ctx context.Context, query cocapi.QueryGetLocations) (cocapi.LocationListResponse, error)
	GetClanRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanRanking) (cocapi.ClanRankingListResponse, error)
	GetPlayerRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerRanking) (cocapi.PlayerRankingListResponse, error)
	GetClanCapitalRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanCapitalRanking) (cocapi.ClanCapitalRankingListResponse, error)
	GetClanBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanBuilderBaseRanking) (cocapi.ClanBuilderBaseRankingListResponse, error)
	GetPlayerBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerBuilderBaseRanking) (cocapi.PlayerBuilderBaseRankingListResponse, error)
}

type RankingService struct {
	cocapi   cocapiRankingClient
	cache    rankingCache
	cacheTTL time.Duration
}

func NewRankingService(cocapi cocapiRankingClient, cache rankingCache, ttl time.Duration) *RankingService {
	if ttl <= 0 { ttl = 10 * time.Minute }
	return &RankingService{cocapi: cocapi, cache: cache, cacheTTL: ttl}
}

func (s *RankingService) GetLocations(ctx context.Context) (ranking.LocationListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLocations(ctx)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetLocations(ctx, cocapi.QueryGetLocations{})
	if err != nil { return ranking.LocationListResponse{}, mapCocapiRankingError(err, "") }
	resp := ranking.LocationListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, loc := range cocapiResp.Items {
		resp.Items = append(resp.Items, ranking.Location{
			ID: loc.ID, Name: loc.Name, CountryCode: loc.CountryCode,
			IsCountry: loc.IsCountry, LocalizedName: loc.LocalizedName,
		})
	}
	if s.cache != nil { _ = s.cache.SetLocations(ctx, resp, s.cacheTTL) }
	return resp, nil
}

func (s *RankingService) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanRanking(ctx, locationID)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetClanRanking(ctx, locationID, cocapi.QueryGetClanRanking{})
	if err != nil { return ranking.ClanRankingListResponse{}, mapCocapiRankingError(err, wardomain.ErrorLocationNotFound) }
	resp := ranking.ClanRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.ClanRankingEntry{Tag: entry.Tag, Name: entry.Name, ClanLevel: entry.ClanLevel, ClanPoints: entry.ClanPoints, Members: entry.Members, Rank: entry.Rank, PreviousRank: entry.PreviousRank, BadgeURLs: entry.BadgeURLs}
		if entry.Location.ID != 0 { item.Location = &ranking.Location{ID: entry.Location.ID, Name: entry.Location.Name, CountryCode: entry.Location.CountryCode, IsCountry: entry.Location.IsCountry, LocalizedName: entry.Location.LocalizedName} }
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil { _ = s.cache.SetClanRanking(ctx, locationID, resp, s.cacheTTL) }
	return resp, nil
}

func (s *RankingService) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetPlayerRanking(ctx, locationID)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetPlayerRanking(ctx, locationID, cocapi.QueryGetPlayerRanking{})
	if err != nil { return ranking.PlayerRankingListResponse{}, mapCocapiRankingError(err, wardomain.ErrorLocationNotFound) }
	resp := ranking.PlayerRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.PlayerRankingEntry{Tag: entry.Tag, Name: entry.Name, ExpLevel: entry.ExpLevel, Trophies: entry.Trophies, Rank: entry.Rank, PreviousRank: entry.PreviousRank, AttackWins: entry.AttackWins, DefenseWins: entry.DefenseWins}
		if entry.Clan.Tag != "" { item.Clan = &ranking.ClanRef{Tag: entry.Clan.Tag, Name: entry.Clan.Name, BadgeURLs: entry.Clan.BadgeURLs} }
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil { _ = s.cache.SetPlayerRanking(ctx, locationID, resp, s.cacheTTL) }
	return resp, nil
}

func (s *RankingService) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanCapitalRanking(ctx, locationID)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetClanCapitalRanking(ctx, locationID, cocapi.QueryGetClanCapitalRanking{})
	if err != nil { return ranking.ClanCapitalRankingListResponse{}, mapCocapiRankingError(err, wardomain.ErrorLocationNotFound) }
	resp := ranking.ClanCapitalRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.ClanCapitalRankingEntry{Tag: entry.Tag, Name: entry.Name, ClanLevel: entry.ClanLevel, ClanCapitalPoints: entry.ClanCapitalPoints, Rank: entry.Rank, PreviousRank: entry.PreviousRank, Members: entry.Members, BadgeURLs: entry.BadgeURLs}
		if entry.Location.ID != 0 { item.Location = &ranking.Location{ID: entry.Location.ID, Name: entry.Location.Name, CountryCode: entry.Location.CountryCode, IsCountry: entry.Location.IsCountry, LocalizedName: entry.Location.LocalizedName} }
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil { _ = s.cache.SetClanCapitalRanking(ctx, locationID, resp, s.cacheTTL) }
	return resp, nil
}

func (s *RankingService) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanBuilderBaseRanking(ctx, locationID)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetClanBuilderBaseRanking(ctx, locationID, cocapi.QueryGetClanBuilderBaseRanking{})
	if err != nil { return ranking.ClanBuilderBaseRankingListResponse{}, mapCocapiRankingError(err, wardomain.ErrorLocationNotFound) }
	resp := ranking.ClanBuilderBaseRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.ClanBuilderBaseRankingEntry{Tag: entry.Tag, Name: entry.Name, ClanLevel: entry.ClanLevel, ClanBuilderBasePoints: entry.ClanBuilderBasePoints, Rank: entry.Rank, PreviousRank: entry.PreviousRank, Members: entry.Members, BadgeURLs: entry.BadgeURLs}
		if entry.Location.ID != 0 { item.Location = &ranking.Location{ID: entry.Location.ID, Name: entry.Location.Name, CountryCode: entry.Location.CountryCode, IsCountry: entry.Location.IsCountry, LocalizedName: entry.Location.LocalizedName} }
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil { _ = s.cache.SetClanBuilderBaseRanking(ctx, locationID, resp, s.cacheTTL) }
	return resp, nil
}

func (s *RankingService) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetPlayerBuilderBaseRanking(ctx, locationID)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetPlayerBuilderBaseRanking(ctx, locationID, cocapi.QueryGetPlayerBuilderBaseRanking{})
	if err != nil { return ranking.PlayerBuilderBaseRankingListResponse{}, mapCocapiRankingError(err, wardomain.ErrorLocationNotFound) }
	resp := ranking.PlayerBuilderBaseRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.PlayerBuilderBaseRankingEntry{Tag: entry.Tag, Name: entry.Name, ExpLevel: entry.ExpLevel, BuilderBaseTrophies: entry.BuilderBaseTrophies, Rank: entry.Rank, PreviousRank: entry.PreviousRank}
		if entry.Clan.Tag != "" { item.Clan = &ranking.ClanRef{Tag: entry.Clan.Tag, Name: entry.Clan.Name, BadgeURLs: entry.Clan.BadgeURLs} }
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil { _ = s.cache.SetPlayerBuilderBaseRanking(ctx, locationID, resp, s.cacheTTL) }
	return resp, nil
}

func toDomainRankingPaging(p cocapi.Paging) ranking.Paging {
	return ranking.Paging{Cursors: ranking.Cursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiRankingError(err error, notFoundCode string) error {
	switch {
	case errors.Is(err, cocapi.ErrNotFound):
		if notFoundCode != "" { return wardomain.NewError(notFoundCode, err.Error()) }
		return wardomain.NewError(wardomain.ErrorWarNotFound, err.Error())
	case errors.Is(err, cocapi.ErrAPINotConfigured):
		return wardomain.NewError(wardomain.ErrorAPINotConfigured, err.Error())
	case errors.Is(err, cocapi.ErrAPIAccessDenied):
		return wardomain.NewError(wardomain.ErrorAPIAccessDenied, err.Error())
	case errors.Is(err, cocapi.ErrAPIRequestFailed):
		return wardomain.NewError(wardomain.ErrorAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrRateLimited):
		return wardomain.NewError(wardomain.ErrorAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrAPIResponseInvalid):
		return wardomain.NewError(wardomain.ErrorAPIResponseInvalid, err.Error())
	default:
		return err
	}
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `go test ./internal/service/ -run TestRankingService -v`
Expected: 4 个测试全部 PASS

- [ ] **Step 6: Commit**

```bash
git add internal/service/ranking_service.go internal/service/ranking_service_test.go internal/service/cocapi_fake_test.go
git commit -m "Add RankingService with 6 ranking endpoints and caching"
```

---

### Task 4:LeagueService + 测试

**Files:**
- Create: `internal/service/league_service.go`
- Create: `internal/service/league_service_test.go`

- [ ] **Step 1: 写 LeagueService 测试(先写测试)**

创建 `internal/service/league_service_test.go`:

```go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/ww1489/WarSpark/internal/domain/league"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeLeagueCache struct {
	leagues      league.LeagueListResponse
	leaguesHit   bool
	l            league.League
	leagueHit    bool
	seasons      league.LeagueSeasonListResponse
	seasonsHit   bool
	rankings     league.LeagueSeasonRankingListResponse
	rankingsHit  bool
	tiers        league.LeagueTierListResponse
	tiersHit     bool
	tier         league.LeagueTier
	tierHit      bool
	history      league.LeagueSeasonResultListResponse
	historyHit   bool
	warLeagues   league.WarLeagueListResponse
	warLeaguesHit bool
	warLeague    league.WarLeague
	warLeagueHit bool
}

func (f *fakeLeagueCache) GetLeagues(ctx context.Context) (league.LeagueListResponse, bool, error) { return f.leagues, f.leaguesHit, nil }
func (f *fakeLeagueCache) SetLeagues(ctx context.Context, resp league.LeagueListResponse, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetLeague(ctx context.Context, id string) (league.League, bool, error) { return f.l, f.leagueHit, nil }
func (f *fakeLeagueCache) SetLeague(ctx context.Context, id string, l league.League, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, bool, error) { return f.seasons, f.seasonsHit, nil }
func (f *fakeLeagueCache) SetLeagueSeasons(ctx context.Context, leagueID string, resp league.LeagueSeasonListResponse, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, bool, error) { return f.rankings, f.rankingsHit, nil }
func (f *fakeLeagueCache) SetLeagueSeasonRankings(ctx context.Context, leagueID, season string, resp league.LeagueSeasonRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, bool, error) { return f.tiers, f.tiersHit, nil }
func (f *fakeLeagueCache) SetLeagueTiers(ctx context.Context, leagueID, season string, resp league.LeagueTierListResponse, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, bool, error) { return f.tier, f.tierHit, nil }
func (f *fakeLeagueCache) SetLeagueTier(ctx context.Context, tierID string, t league.LeagueTier, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, bool, error) { return f.history, f.historyHit, nil }
func (f *fakeLeagueCache) SetLeagueHistory(ctx context.Context, playerTag string, resp league.LeagueSeasonResultListResponse, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, bool, error) { return f.warLeagues, f.warLeaguesHit, nil }
func (f *fakeLeagueCache) SetWarLeagues(ctx context.Context, resp league.WarLeagueListResponse, ttl time.Duration) error { return nil }
func (f *fakeLeagueCache) GetWarLeague(ctx context.Context, id string) (league.WarLeague, bool, error) { return f.warLeague, f.warLeagueHit, nil }
func (f *fakeLeagueCache) SetWarLeague(ctx context.Context, id string, l league.WarLeague, ttl time.Duration) error { return nil }

func TestLeagueServiceGetLeaguesCached(t *testing.T) {
	cached := league.LeagueListResponse{Items: []league.League{{ID: 1, Name: "Champion"}}}
	svc := NewLeagueService(&fakeCocapiClient{}, &fakeLeagueCache{leagues: cached, leaguesHit: true}, 30*time.Minute)
	resp, err := svc.GetLeagues(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Champion" { t.Fatalf("got %+v", resp.Items) }
}

func TestLeagueServiceGetLeaguesFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagues: cocapi.LeagueListResponse{Items: []cocapi.League{{ID: 1, Name: "Champion"}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagues(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Champion" { t.Fatalf("got %+v", resp.Items) }
}

func TestLeagueServiceGetLeague(t *testing.T) {
	fakeAPI := &fakeCocapiClient{league: cocapi.League{ID: 1, Name: "Champion"}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	l, err := svc.GetLeague(context.Background(), "1")
	if err != nil { t.Fatalf("error: %v", err) }
	if l.Name != "Champion" { t.Fatalf("got %q", l.Name) }
}

func TestLeagueServiceGetWarLeagues(t *testing.T) {
	fakeAPI := &fakeCocapiClient{warLeagues: cocapi.WarLeagueListResponse{Items: []cocapi.WarLeague{{ID: 1, Name: "Master"}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetWarLeagues(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Master" { t.Fatalf("got %+v", resp.Items) }
}

func TestLeagueServiceGetLeagueHistory(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagueHistory: cocapi.LeagueSeasonResultListResponse{Items: []cocapi.LeagueSeasonResult{{Placement: 1}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagueHistory(context.Background(), "#P1")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Placement != 1 { t.Fatalf("got %+v", resp.Items) }
}

func TestLeagueServiceNotFound(t *testing.T) {
	fakeAPI := &fakeCocapiClient{err: cocapi.ErrNotFound}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	_, err := svc.GetLeague(context.Background(), "999")
	if err == nil { t.Fatal("expected error") }
	if wardomain.ErrorCode(err) != wardomain.ErrorLeagueNotFound { t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorLeagueNotFound) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/service/ -run TestLeagueService -v`
Expected: FAIL — `NewLeagueService undefined`

- [ ] **Step 3: 实现 LeagueService**

创建 `internal/service/league_service.go`:

```go
package service

import (
	"context"
	"errors"
	"time"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	"github.com/ww1489/WarSpark/internal/domain/league"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type cocapiLeagueClient interface {
	GetLeagues(ctx context.Context, query cocapi.QueryGetLeagues) (cocapi.LeagueListResponse, error)
	GetLeague(ctx context.Context, leagueID string) (cocapi.League, error)
	GetLeagueSeasons(ctx context.Context, leagueID string, query cocapi.QueryGetLeagueSeasons) (cocapi.LeagueSeasonListResponse, error)
	GetLeagueSeasonRankings(ctx context.Context, leagueID, seasonID string, query cocapi.QueryGetLeagueSeasonRankings) (cocapi.PlayerRankingListResponse, error)
	GetLeagueTiers(ctx context.Context, query cocapi.QueryGetLeagueTiers) (cocapi.LeagueTierListResponse, error)
	GetLeagueTier(ctx context.Context, leagueTierID string) (cocapi.LeagueTier, error)
	GetLeagueHistory(ctx context.Context, playerTag string) (cocapi.LeagueSeasonResultListResponse, error)
	GetWarLeagues(ctx context.Context, query cocapi.QueryGetWarLeagues) (cocapi.WarLeagueListResponse, error)
	GetWarLeague(ctx context.Context, leagueID string) (cocapi.WarLeague, error)
}

type leagueCache interface {
	GetLeagues(ctx context.Context) (league.LeagueListResponse, bool, error)
	SetLeagues(ctx context.Context, resp league.LeagueListResponse, ttl time.Duration) error
	GetLeague(ctx context.Context, id string) (league.League, bool, error)
	SetLeague(ctx context.Context, id string, l league.League, ttl time.Duration) error
	GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, bool, error)
	SetLeagueSeasons(ctx context.Context, leagueID string, resp league.LeagueSeasonListResponse, ttl time.Duration) error
	GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, bool, error)
	SetLeagueSeasonRankings(ctx context.Context, leagueID, season string, resp league.LeagueSeasonRankingListResponse, ttl time.Duration) error
	GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, bool, error)
	SetLeagueTiers(ctx context.Context, leagueID, season string, resp league.LeagueTierListResponse, ttl time.Duration) error
	GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, bool, error)
	SetLeagueTier(ctx context.Context, tierID string, t league.LeagueTier, ttl time.Duration) error
	GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, bool, error)
	SetLeagueHistory(ctx context.Context, playerTag string, resp league.LeagueSeasonResultListResponse, ttl time.Duration) error
	GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, bool, error)
	SetWarLeagues(ctx context.Context, resp league.WarLeagueListResponse, ttl time.Duration) error
	GetWarLeague(ctx context.Context, id string) (league.WarLeague, bool, error)
	SetWarLeague(ctx context.Context, id string, l league.WarLeague, ttl time.Duration) error
}

type LeagueService struct {
	cocapi   cocapiLeagueClient
	cache    leagueCache
	cacheTTL time.Duration
}

func NewLeagueService(cocapi cocapiLeagueClient, cache leagueCache, ttl time.Duration) *LeagueService {
	if ttl <= 0 { ttl = 30 * time.Minute }
	return &LeagueService{cocapi: cocapi, cache: cache, cacheTTL: ttl}
}

func (s *LeagueService) GetLeagues(ctx context.Context) (league.LeagueListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagues(ctx)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetLeagues(ctx, cocapi.QueryGetLeagues{})
	if err != nil { return league.LeagueListResponse{}, mapCocapiLeagueError(err, "") }
	resp := league.LeagueListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, l := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.League{ID: l.ID, Name: string(l.Name), IconURLs: l.IconURLs})
	}
	if s.cache != nil { _ = s.cache.SetLeagues(ctx, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LeagueService) GetLeague(ctx context.Context, leagueID string) (league.League, error) {
	if s.cache != nil {
		l, ok, err := s.cache.GetLeague(ctx, leagueID)
		if err == nil && ok { return l, nil }
	}
	l, err := s.cocapi.GetLeague(ctx, leagueID)
	if err != nil { return league.League{}, mapCocapiLeagueError(err, wardomain.ErrorLeagueNotFound) }
	dom := league.League{ID: l.ID, Name: string(l.Name), IconURLs: l.IconURLs}
	if s.cache != nil { _ = s.cache.SetLeague(ctx, leagueID, dom, s.cacheTTL) }
	return dom, nil
}

func (s *LeagueService) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueSeasons(ctx, leagueID)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetLeagueSeasons(ctx, leagueID, cocapi.QueryGetLeagueSeasons{})
	if err != nil { return league.LeagueSeasonListResponse{}, mapCocapiLeagueError(err, wardomain.ErrorLeagueNotFound) }
	resp := league.LeagueSeasonListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, s := range cocapiResp.Items { resp.Items = append(resp.Items, league.LeagueSeason{ID: s.ID}) }
	if s.cache != nil { _ = s.cache.SetLeagueSeasons(ctx, leagueID, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LeagueService) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueSeasonRankings(ctx, leagueID, season)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetLeagueSeasonRankings(ctx, leagueID, season, cocapi.QueryGetLeagueSeasonRankings{})
	if err != nil { return league.LeagueSeasonRankingListResponse{}, mapCocapiLeagueError(err, wardomain.ErrorLeagueNotFound) }
	resp := league.LeagueSeasonRankingListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := league.LeagueSeasonRankingEntry{Tag: entry.Tag, Name: entry.Name, ExpLevel: entry.ExpLevel, Trophies: entry.Trophies, Rank: entry.Rank, PreviousRank: entry.PreviousRank, AttackWins: entry.AttackWins, DefenseWins: entry.DefenseWins}
		if entry.Clan.Tag != "" { item.Clan = &league.ClanRef{Tag: entry.Clan.Tag, Name: entry.Clan.Name, BadgeURLs: entry.Clan.BadgeURLs} }
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil { _ = s.cache.SetLeagueSeasonRankings(ctx, leagueID, season, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LeagueService) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueTiers(ctx, leagueID, season)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetLeagueTiers(ctx, cocapi.QueryGetLeagueTiers{})
	if err != nil { return league.LeagueTierListResponse{}, mapCocapiLeagueError(err, wardomain.ErrorLeagueNotFound) }
	resp := league.LeagueTierListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, t := range cocapiResp.Items { resp.Items = append(resp.Items, league.LeagueTier{ID: t.ID, Name: string(t.Name), IconURLs: t.IconURLs}) }
	if s.cache != nil { _ = s.cache.SetLeagueTiers(ctx, leagueID, season, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LeagueService) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, error) {
	if s.cache != nil {
		t, ok, err := s.cache.GetLeagueTier(ctx, tierID)
		if err == nil && ok { return t, nil }
	}
	t, err := s.cocapi.GetLeagueTier(ctx, tierID)
	if err != nil { return league.LeagueTier{}, mapCocapiLeagueError(err, wardomain.ErrorLeagueNotFound) }
	dom := league.LeagueTier{ID: t.ID, Name: string(t.Name), IconURLs: t.IconURLs}
	if s.cache != nil { _ = s.cache.SetLeagueTier(ctx, tierID, dom, s.cacheTTL) }
	return dom, nil
}

func (s *LeagueService) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, error) {
	normalizedTag, err := NormalizeClanTag(playerTag)
	if err != nil { return league.LeagueSeasonResultListResponse{}, err }
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueHistory(ctx, normalizedTag)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetLeagueHistory(ctx, normalizedTag)
	if err != nil { return league.LeagueSeasonResultListResponse{}, mapCocapiLeagueError(err, "") }
	resp := league.LeagueSeasonResultListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, r := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.LeagueSeasonResult{
			LeagueSeasonID: int64(r.LeagueSeasonID), LeagueTierID: r.LeagueTierID, Trophies: r.Trophies,
			Placement: r.Placement, AttackWins: r.AttackWins, AttackLosses: r.AttackLosses, AttackStars: r.AttackStars,
			DefenseWins: r.DefenseWins, DefenseLosses: r.DefenseLosses, DefenseStars: r.DefenseStars, MaxBattles: r.MaxBattles,
		})
	}
	if s.cache != nil { _ = s.cache.SetLeagueHistory(ctx, normalizedTag, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LeagueService) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetWarLeagues(ctx)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetWarLeagues(ctx, cocapi.QueryGetWarLeagues{})
	if err != nil { return league.WarLeagueListResponse{}, mapCocapiLeagueError(err, "") }
	resp := league.WarLeagueListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, wl := range cocapiResp.Items { resp.Items = append(resp.Items, league.WarLeague{ID: wl.ID, Name: string(wl.Name)}) }
	if s.cache != nil { _ = s.cache.SetWarLeagues(ctx, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LeagueService) GetWarLeague(ctx context.Context, leagueID string) (league.WarLeague, error) {
	if s.cache != nil {
		l, ok, err := s.cache.GetWarLeague(ctx, leagueID)
		if err == nil && ok { return l, nil }
	}
	wl, err := s.cocapi.GetWarLeague(ctx, leagueID)
	if err != nil { return league.WarLeague{}, mapCocapiLeagueError(err, wardomain.ErrorLeagueNotFound) }
	dom := league.WarLeague{ID: wl.ID, Name: string(wl.Name)}
	if s.cache != nil { _ = s.cache.SetWarLeague(ctx, leagueID, dom, s.cacheTTL) }
	return dom, nil
}

func toDomainLeaguePaging(p cocapi.Paging) league.Paging {
	return league.Paging{Cursors: league.Cursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiLeagueError(err error, notFoundCode string) error {
	if notFoundCode != "" && errors.Is(err, cocapi.ErrNotFound) {
		return wardomain.NewError(notFoundCode, err.Error())
	}
	return mapCocapiRankingError(err, notFoundCode)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/service/ -run TestLeagueService -v`
Expected: 6 个测试全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/league_service.go internal/service/league_service_test.go
git commit -m "Add LeagueService with 9 league endpoints and caching"
```

---

### Task 5:LabelService + 测试

**Files:**
- Create: `internal/service/label_service.go`
- Create: `internal/service/label_service_test.go`

- [ ] **Step 1: 写 LabelService 测试(先写测试)**

创建 `internal/service/label_service_test.go`:

```go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/ww1489/WarSpark/internal/domain/label"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeLabelCache struct {
	clanLabels   label.LabelListResponse
	clanHit      bool
	playerLabels label.LabelListResponse
	playerHit    bool
}

func (f *fakeLabelCache) GetClanLabels(ctx context.Context) (label.LabelListResponse, bool, error) { return f.clanLabels, f.clanHit, nil }
func (f *fakeLabelCache) SetClanLabels(ctx context.Context, resp label.LabelListResponse, ttl time.Duration) error { return nil }
func (f *fakeLabelCache) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, bool, error) { return f.playerLabels, f.playerHit, nil }
func (f *fakeLabelCache) SetPlayerLabels(ctx context.Context, resp label.LabelListResponse, ttl time.Duration) error { return nil }

func TestLabelServiceGetClanLabelsCached(t *testing.T) {
	cached := label.LabelListResponse{Items: []label.Label{{ID: 1, Name: "Clan War"}}}
	svc := NewLabelService(&fakeCocapiClient{}, &fakeLabelCache{clanLabels: cached, clanHit: true}, time.Hour)
	resp, err := svc.GetClanLabels(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Clan War" { t.Fatalf("got %+v", resp.Items) }
}

func TestLabelServiceGetClanLabelsFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{clanLabels: cocapi.LabelListResponse{Items: []cocapi.Label{{ID: 1, Name: "Clan War"}}}}
	svc := NewLabelService(fakeAPI, &fakeLabelCache{}, time.Hour)
	resp, err := svc.GetClanLabels(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Clan War" { t.Fatalf("got %+v", resp.Items) }
}

func TestLabelServiceGetPlayerLabels(t *testing.T) {
	fakeAPI := &fakeCocapiClient{playerLabels: cocapi.LabelListResponse{Items: []cocapi.Label{{ID: 2, Name: "Active"}}}}
	svc := NewLabelService(fakeAPI, &fakeLabelCache{}, time.Hour)
	resp, err := svc.GetPlayerLabels(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Active" { t.Fatalf("got %+v", resp.Items) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/service/ -run TestLabelService -v`
Expected: FAIL — `NewLabelService undefined`

- [ ] **Step 3: 实现 LabelService**

创建 `internal/service/label_service.go`:

```go
package service

import (
	"context"
	"time"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	"github.com/ww1489/WarSpark/internal/domain/label"
)

type labelCache interface {
	GetClanLabels(ctx context.Context) (label.LabelListResponse, bool, error)
	SetClanLabels(ctx context.Context, resp label.LabelListResponse, ttl time.Duration) error
	GetPlayerLabels(ctx context.Context) (label.LabelListResponse, bool, error)
	SetPlayerLabels(ctx context.Context, resp label.LabelListResponse, ttl time.Duration) error
}

type cocapiLabelClient interface {
	GetClanLabels(ctx context.Context, query cocapi.QueryGetClanLabels) (cocapi.LabelListResponse, error)
	GetPlayerLabels(ctx context.Context, query cocapi.QueryGetPlayerLabels) (cocapi.LabelListResponse, error)
}

type LabelService struct {
	cocapi   cocapiLabelClient
	cache    labelCache
	cacheTTL time.Duration
}

func NewLabelService(cocapi cocapiLabelClient, cache labelCache, ttl time.Duration) *LabelService {
	if ttl <= 0 { ttl = time.Hour }
	return &LabelService{cocapi: cocapi, cache: cache, cacheTTL: ttl}
}

func (s *LabelService) GetClanLabels(ctx context.Context) (label.LabelListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanLabels(ctx)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetClanLabels(ctx, cocapi.QueryGetClanLabels{})
	if err != nil { return label.LabelListResponse{}, mapCocapiLabelError(err) }
	resp := label.LabelListResponse{Paging: toDomainLabelPaging(cocapiResp.Paging)}
	for _, lbl := range cocapiResp.Items {
		resp.Items = append(resp.Items, label.Label{ID: lbl.ID, Name: string(lbl.Name), IconURLs: lbl.IconURLs})
	}
	if s.cache != nil { _ = s.cache.SetClanLabels(ctx, resp, s.cacheTTL) }
	return resp, nil
}

func (s *LabelService) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetPlayerLabels(ctx)
		if err == nil && ok { return resp, nil }
	}
	cocapiResp, err := s.cocapi.GetPlayerLabels(ctx, cocapi.QueryGetPlayerLabels{})
	if err != nil { return label.LabelListResponse{}, mapCocapiLabelError(err) }
	resp := label.LabelListResponse{Paging: toDomainLabelPaging(cocapiResp.Paging)}
	for _, lbl := range cocapiResp.Items {
		resp.Items = append(resp.Items, label.Label{ID: lbl.ID, Name: string(lbl.Name), IconURLs: lbl.IconURLs})
	}
	if s.cache != nil { _ = s.cache.SetPlayerLabels(ctx, resp, s.cacheTTL) }
	return resp, nil
}

func toDomainLabelPaging(p cocapi.Paging) label.Paging {
	return label.Paging{Cursors: label.Cursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiLabelError(err error) error {
	return mapCocapiRankingError(err, "")
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/service/ -run TestLabelService -v`
Expected: 3 个测试全部 PASS

- [ ] **Step 5: 全量 service 测试验证**

Run: `go test ./internal/service/ -v`
Expected: 全部 PASS(原有 + 新增 13 个 = 总计 ~25 个测试)

- [ ] **Step 6: Commit**

```bash
git add internal/service/label_service.go internal/service/label_service_test.go
git commit -m "Add LabelService with clan/player label endpoints and caching"
```

---

### Task 6:RankingController + 测试

**Files:**
- Create: `internal/controller/ranking_controller.go`
- Create: `internal/controller/ranking_controller_test.go`

- [ ] **Step 1: 写 RankingController 测试(先写测试)**

创建 `internal/controller/ranking_controller_test.go`:

```go
package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeRankingService struct {
	locations    ranking.LocationListResponse
	clanRanking  ranking.ClanRankingListResponse
	playerRanking ranking.PlayerRankingListResponse
	err          error
}

func (f *fakeRankingService) GetLocations(ctx context.Context) (ranking.LocationListResponse, error) { return f.locations, f.err }
func (f *fakeRankingService) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, error) { return f.clanRanking, f.err }
func (f *fakeRankingService) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, error) { return f.playerRanking, f.err }
func (f *fakeRankingService) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, error) { return ranking.ClanCapitalRankingListResponse{}, f.err }
func (f *fakeRankingService) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, error) { return ranking.ClanBuilderBaseRankingListResponse{}, f.err }
func (f *fakeRankingService) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, error) { return ranking.PlayerBuilderBaseRankingListResponse{}, f.err }

func TestRankingControllerGetLocations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{locations: ranking.LocationListResponse{Items: []ranking.Location{{ID: 1, Name: "Global"}}}}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations", ctrl.GetLocations)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String()) }
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) != 1 { t.Fatalf("len = %d", len(items)) }
}

func TestRankingControllerGetClanRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/clans", ctrl.GetClanRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/global/rankings/clans", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestRankingControllerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{err: wardomain.NewError(wardomain.ErrorLocationNotFound, "not found")}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/clans", ctrl.GetClanRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/999/rankings/clans", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound { t.Fatalf("status = %d, want 404", rec.Code) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/controller/ -run TestRankingController -v`
Expected: FAIL — `NewRankingController undefined`

- [ ] **Step 3: 实现 RankingController**

创建 `internal/controller/ranking_controller.go`:

```go
package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type RankingReader interface {
	GetLocations(ctx context.Context) (ranking.LocationListResponse, error)
	GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, error)
	GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, error)
	GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, error)
	GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, error)
	GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, error)
}

type RankingController struct {
	service RankingReader
}

func NewRankingController(service RankingReader) *RankingController {
	return &RankingController{service: service}
}

func (c *RankingController) GetLocations(ctx *gin.Context) {
	resp, err := c.service.GetLocations(ctx.Request.Context())
	if err != nil { failRanking(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *RankingController) GetClanRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetClanRanking(ctx.Request.Context(), locationID)
	if err != nil { failRanking(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *RankingController) GetPlayerRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetPlayerRanking(ctx.Request.Context(), locationID)
	if err != nil { failRanking(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *RankingController) GetClanCapitalRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetClanCapitalRanking(ctx.Request.Context(), locationID)
	if err != nil { failRanking(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *RankingController) GetClanBuilderBaseRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetClanBuilderBaseRanking(ctx.Request.Context(), locationID)
	if err != nil { failRanking(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *RankingController) GetPlayerBuilderBaseRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetPlayerBuilderBaseRanking(ctx.Request.Context(), locationID)
	if err != nil { failRanking(ctx, err); return }
	utils.OK(ctx, resp)
}

func failRanking(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorLocationNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/controller/ -run TestRankingController -v`
Expected: 3 个测试全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/controller/ranking_controller.go internal/controller/ranking_controller_test.go
git commit -m "Add RankingController with 6 location/ranking endpoints"
```

---

### Task 7:LeagueController + 测试

**Files:**
- Create: `internal/controller/league_controller.go`
- Create: `internal/controller/league_controller_test.go`

- [ ] **Step 1: 写 LeagueController 测试(先写测试)**

创建 `internal/controller/league_controller_test.go`:

```go
package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/league"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeLeagueService struct {
	leagues    league.LeagueListResponse
	l          league.League
	seasons    league.LeagueSeasonListResponse
	rankings   league.LeagueSeasonRankingListResponse
	tiers      league.LeagueTierListResponse
	tier       league.LeagueTier
	history    league.LeagueSeasonResultListResponse
	warLeagues league.WarLeagueListResponse
	warLeague  league.WarLeague
	err        error
}

func (f *fakeLeagueService) GetLeagues(ctx context.Context) (league.LeagueListResponse, error) { return f.leagues, f.err }
func (f *fakeLeagueService) GetLeague(ctx context.Context, id string) (league.League, error) { return f.l, f.err }
func (f *fakeLeagueService) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, error) { return f.seasons, f.err }
func (f *fakeLeagueService) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, error) { return f.rankings, f.err }
func (f *fakeLeagueService) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, error) { return f.tiers, f.err }
func (f *fakeLeagueService) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, error) { return f.tier, f.err }
func (f *fakeLeagueService) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, error) { return f.history, f.err }
func (f *fakeLeagueService) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, error) { return f.warLeagues, f.err }
func (f *fakeLeagueService) GetWarLeague(ctx context.Context, id string) (league.WarLeague, error) { return f.warLeague, f.err }

func TestLeagueControllerGetLeagues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{leagues: league.LeagueListResponse{Items: []league.League{{Name: "Champion"}}}}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues", ctrl.GetLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeague(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{l: league.League{Name: "Champion"}}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id", ctrl.GetLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetWarLeagues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{warLeagues: league.WarLeagueListResponse{Items: []league.WarLeague{{Name: "Master"}}}}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/war-leagues", ctrl.GetWarLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/war-leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeagueHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons/:season/tiers/:tier/history", ctrl.GetLeagueHistory)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons/2026-01/tiers/1/history?player_tag=%23P1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{err: wardomain.NewError(wardomain.ErrorLeagueNotFound, "not found")}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id", ctrl.GetLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/999", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound { t.Fatalf("status = %d, want 404", rec.Code) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/controller/ -run TestLeagueController -v`
Expected: FAIL — `NewLeagueController undefined`

- [ ] **Step 3: 实现 LeagueController**

创建 `internal/controller/league_controller.go`:

```go
package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/league"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LeagueReader interface {
	GetLeagues(ctx context.Context) (league.LeagueListResponse, error)
	GetLeague(ctx context.Context, id string) (league.League, error)
	GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, error)
	GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, error)
	GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, error)
	GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, error)
	GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, error)
	GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, error)
	GetWarLeague(ctx context.Context, id string) (league.WarLeague, error)
}

type LeagueController struct {
	service LeagueReader
}

func NewLeagueController(service LeagueReader) *LeagueController {
	return &LeagueController{service: service}
}

func (c *LeagueController) GetLeagues(ctx *gin.Context) {
	resp, err := c.service.GetLeagues(ctx.Request.Context())
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetLeague(ctx *gin.Context) {
	id := ctx.Param("id")
	l, err := c.service.GetLeague(ctx.Request.Context(), id)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, l)
}

func (c *LeagueController) GetLeagueSeasons(ctx *gin.Context) {
	id := ctx.Param("id")
	resp, err := c.service.GetLeagueSeasons(ctx.Request.Context(), id)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetLeagueSeasonRankings(ctx *gin.Context) {
	id := ctx.Param("id"); season := ctx.Param("season")
	resp, err := c.service.GetLeagueSeasonRankings(ctx.Request.Context(), id, season)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetLeagueTiers(ctx *gin.Context) {
	id := ctx.Param("id"); season := ctx.Param("season")
	resp, err := c.service.GetLeagueTiers(ctx.Request.Context(), id, season)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetLeagueTier(ctx *gin.Context) {
	tierID := ctx.Param("tier")
	t, err := c.service.GetLeagueTier(ctx.Request.Context(), tierID)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, t)
}

func (c *LeagueController) GetLeagueHistory(ctx *gin.Context) {
	playerTag := ctx.Query("player_tag")
	if playerTag == "" { utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "player_tag is required")); return }
	resp, err := c.service.GetLeagueHistory(ctx.Request.Context(), playerTag)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetWarLeagues(ctx *gin.Context) {
	resp, err := c.service.GetWarLeagues(ctx.Request.Context())
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetWarLeague(ctx *gin.Context) {
	id := ctx.Param("id")
	l, err := c.service.GetWarLeague(ctx.Request.Context(), id)
	if err != nil { failLeague(ctx, err); return }
	utils.OK(ctx, l)
}

func failLeague(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorLeagueNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, http.StatusUnprocessableEntity, int(utils.ErrInvalidField), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/controller/ -run TestLeagueController -v`
Expected: 5 个测试全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/controller/league_controller.go internal/controller/league_controller_test.go
git commit -m "Add LeagueController with 9 league/war-league endpoints"
```

---

### Task 8:LabelController + 测试

**Files:**
- Create: `internal/controller/label_controller.go`
- Create: `internal/controller/label_controller_test.go`

- [ ] **Step 1: 写 LabelController 测试(先写测试)**

创建 `internal/controller/label_controller_test.go`:

```go
package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/label"
)

type fakeLabelService struct {
	clanLabels   label.LabelListResponse
	playerLabels label.LabelListResponse
	err          error
}

func (f *fakeLabelService) GetClanLabels(ctx context.Context) (label.LabelListResponse, error) { return f.clanLabels, f.err }
func (f *fakeLabelService) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, error) { return f.playerLabels, f.err }

func TestLabelControllerGetClanLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLabelService{clanLabels: label.LabelListResponse{Items: []label.Label{{ID: 1, Name: "Clan War"}}}}
	ctrl := NewLabelController(svc)
	router := gin.New()
	router.GET("/clans/labels", ctrl.GetClanLabels)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/labels", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLabelControllerGetPlayerLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLabelService{playerLabels: label.LabelListResponse{Items: []label.Label{{ID: 2, Name: "Active"}}}}
	ctrl := NewLabelController(svc)
	router := gin.New()
	router.GET("/players/labels", ctrl.GetPlayerLabels)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/labels", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/controller/ -run TestLabelController -v`
Expected: FAIL — `NewLabelController undefined`

- [ ] **Step 3: 实现 LabelController**

创建 `internal/controller/label_controller.go`:

```go
package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/label"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LabelReader interface {
	GetClanLabels(ctx context.Context) (label.LabelListResponse, error)
	GetPlayerLabels(ctx context.Context) (label.LabelListResponse, error)
}

type LabelController struct {
	service LabelReader
}

func NewLabelController(service LabelReader) *LabelController {
	return &LabelController{service: service}
}

func (c *LabelController) GetClanLabels(ctx *gin.Context) {
	resp, err := c.service.GetClanLabels(ctx.Request.Context())
	if err != nil { failLabel(ctx, err); return }
	utils.OK(ctx, resp)
}

func (c *LabelController) GetPlayerLabels(ctx *gin.Context) {
	resp, err := c.service.GetPlayerLabels(ctx.Request.Context())
	if err != nil { failLabel(ctx, err); return }
	utils.OK(ctx, resp)
}

func failLabel(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/controller/ -run TestLabelController -v`
Expected: 2 个测试全部 PASS

- [ ] **Step 5: 全量 controller 测试**

Run: `go test ./internal/controller/ -v`
Expected: 全部 PASS(原有 6 + 新增 10 = 16 个测试)

- [ ] **Step 6: Commit**

```bash
git add internal/controller/label_controller.go internal/controller/label_controller_test.go
git commit -m "Add LabelController with clan/player label endpoints"
```

---

### Task 9:Bootstrap + 路由注册

**Files:**
- Modify: `internal/api/v1/routes.go`(17 个新路由)
- Modify: `internal/app/bootstrap.go`(无改动 — routes.go 已通过 runtimeConfig 持有 Redis 和 Cocapi)

- [ ] **Step 1: 修改 routes.go**

在 `internal/api/v1/routes.go` 中，在 `playerController := controller.NewPlayerController(playerService)` 之后、`router.GET("/health", ...)` 之前追加 wiring 代码:

```go
	rankingService := service.NewRankingService(warAPIClient, infraredis.NewRankingCache(runtimeConfig.Redis), 10*time.Minute)
	rankingController := controller.NewRankingController(rankingService)
	leagueService := service.NewLeagueService(warAPIClient, infraredis.NewLeagueCache(runtimeConfig.Redis), 30*time.Minute)
	leagueController := controller.NewLeagueController(leagueService)
	labelService := service.NewLabelService(warAPIClient, infraredis.NewLabelCache(runtimeConfig.Redis), time.Hour)
	labelController := controller.NewLabelController(labelService)
```

在 `api` group 内，`api.GET("/players/:tag/battle-log", playerController.GetBattleLog)` 之后追加 17 个路由。注意:将 `/clans/labels` 放在 `/clans/:tag` 之前:

```go
		api.GET("/clans/labels", labelController.GetClanLabels)
		api.GET("/players/labels", labelController.GetPlayerLabels)
		api.GET("/locations", rankingController.GetLocations)
		api.GET("/locations/:id/rankings/clans", rankingController.GetClanRanking)
		api.GET("/locations/:id/rankings/players", rankingController.GetPlayerRanking)
		api.GET("/locations/:id/rankings/clans-capital", rankingController.GetClanCapitalRanking)
		api.GET("/locations/:id/rankings/clans-builder-base", rankingController.GetClanBuilderBaseRanking)
		api.GET("/locations/:id/rankings/players-builder-base", rankingController.GetPlayerBuilderBaseRanking)
		api.GET("/leagues", leagueController.GetLeagues)
		api.GET("/leagues/:id", leagueController.GetLeague)
		api.GET("/leagues/:id/seasons", leagueController.GetLeagueSeasons)
		api.GET("/leagues/:id/seasons/:season/rankings", leagueController.GetLeagueSeasonRankings)
		api.GET("/leagues/:id/seasons/:season/tiers", leagueController.GetLeagueTiers)
		api.GET("/leagues/:id/seasons/:season/tiers/:tier", leagueController.GetLeagueTier)
		api.GET("/leagues/:id/seasons/:season/tiers/:tier/history", leagueController.GetLeagueHistory)
		api.GET("/war-leagues", leagueController.GetWarLeagues)
		api.GET("/war-leagues/:id", leagueController.GetWarLeague)
```

确认 `routes.go` 的 import 已有 `"time"`。同时确认 `"github.com/ww1489/WarSpark/pkg/cocapi"` import 未被删除(warAPIClient 已在用)。

- [ ] **Step 2: 编译验证**

Run: `go build ./...`
Expected: 无输出(通过)

- [ ] **Step 3: 运行全量测试**

Run: `make check`
Expected: fmt 无输出、vet 无输出、test 全绿、build 通过

- [ ] **Step 4: Commit**

```bash
git add internal/api/v1/routes.go
git commit -m "Register 17 ranking/league/label routes in v1 API"
```

---

### Task 10:最终验证

- [ ] **Step 1: 确认工作区干净**

```bash
git status
```
Expected: 工作区干净

- [ ] **Step 2: 全量验证**

```bash
make check
```
Expected: 全绿

- [ ] **Step 3: 检查 endpoint 覆盖率**

确认 17 个新路由(共 20 个公开 GET 路由):

| # | 路径 | 模块 |
|---|------|------|
| 1 | /api/v1/locations | ranking |
| 2 | /api/v1/locations/:id/rankings/clans | ranking |
| 3 | /api/v1/locations/:id/rankings/players | ranking |
| 4 | /api/v1/locations/:id/rankings/clans-capital | ranking |
| 5 | /api/v1/locations/:id/rankings/clans-builder-base | ranking |
| 6 | /api/v1/locations/:id/rankings/players-builder-base | ranking |
| 7 | /api/v1/leagues | league |
| 8 | /api/v1/leagues/:id | league |
| 9 | /api/v1/leagues/:id/seasons | league |
| 10 | /api/v1/leagues/:id/seasons/:season/rankings | league |
| 11 | /api/v1/leagues/:id/seasons/:season/tiers | league |
| 12 | /api/v1/leagues/:id/seasons/:season/tiers/:tier | league |
| 13 | /api/v1/leagues/:id/seasons/:season/tiers/:tier/history | league |
| 14 | /api/v1/war-leagues | league |
| 15 | /api/v1/war-leagues/:id | league |
| 16 | /api/v1/clans/labels | label |
| 17 | /api/v1/players/labels | label |

cocapi 端点新增 18 个(21 个原计划 − 3 个偏离)。合计使用:6 → 24 / 35(69%)。

---

## 完成检查清单

- [ ] 17 个接口路由已注册(ranking 6 + league 9 + label 2)
- [ ] 3 个 domain 包(ranking/league/label)
- [ ] 3 个 Redis cache 文件
- [ ] 3 个 service 含缓存 + cocapi 错误映射
- [ ] 3 个 controller 含错误码 HTTP 映射
- [ ] 所有测试全绿
- [ ] `make check` 全绿
- [ ] cocapi 端点使用:6 → 24 / 35(69%)
