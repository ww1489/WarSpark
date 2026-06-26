# 部族概览 + 玩家概览 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 `GET /api/v1/clans/:tag`、`GET /api/v1/players/:tag`、`GET /api/v1/players/:tag/battle-log` 三个接口,接入 cocapi 的 GetClan/GetClanMembers/GetPlayer/GetBattleLog 端点。

**Architecture:** 沿用 adapter 模式:`controller → service → infracoc.Client(adapter) → cocapi.Client`。新增 `internal/domain/clan/` 领域包,`internal/infra/redis/clan_cache.go` Redis 缓存。部族概览聚合 GetClan + GetClanMembers 两个端点。

**Tech Stack:** Go 1.26, Gin, Redis (go-redis), cocapi (本仓库 pkg/cocapi), 项目的 utils.OK/Fail 响应模式。

## Global Constraints

- 使用中文回复(AGENTS.md 规则)。
- 代码不加注释(AGENTS.md 规则),但 Swagger 注释放保留。
- 响应格式:`{"code": 0, "message": "ok", "data": {}}`,用 `utils.OK()` / `utils.Fail()`。
- 错误用 `wardomain.Error` 带稳定 code,controller 通过 `StatusForCode` 做 HTTP 映射。
- Tag 校验复用 `service.NormalizeClanTag`(已有,在 war_service.go)。
- `make check` 必须全绿(fmt + vet + test + build)。
- cocapi 方法签名(本 plan 依赖):
  - `GetClan(ctx, clanTag string) (Clan, error)`
  - `GetClanMembers(ctx, clanTag string, query QueryGetClanMembers) (ClanMemberListResponse, error)`
  - `GetPlayer(ctx, playerTag string) (Player, error)`
  - `GetBattleLog(ctx, playerTag string) (BattleLogEntryListResponse, error)`
- `ClanMemberListResponse` 结构:`{Items []ClanMember, Paging Paging}`
- `BattleLogEntryListResponse` 结构:`{Items []BattleLogEntry, Paging Paging}`

## Spec 偏离说明

Spec 设计文档(`docs/superpowers/specs/2026-06-26-cocapi-features-design.md`)说 adapter 新增 21 个转换方法。本 plan 简化为 4 个 — 只有需要 domain 语义转换的模块(war/clan/player)走 adapter。排名/联赛/标签等纯展示数据(Plan 2)service 直接持有 `cocapi.Client`,不走 adapter。理由:YAGNI,纯透传类型无需隔离层。

---

### Task 1:领域类型 + 错误码扩展

**Files:**
- Create: `internal/domain/clan/clan.go`
- Modify: `internal/domain/war/war.go:8-16`(错误码常量块)

**Interfaces:**
- Produces: `clan.ClanDetail`、`clan.ClanOverview`、`clan.ClanMemberSummary`、`clan.PlayerOverview`、`clan.BattleLogSummary`、`clan.AchievementProgress`、`clan.HeroLevel`、`clan.PlayerClanInfo` 类型;`wardomain.ErrorClanNotFound`、`wardomain.ErrorPlayerNotFound` 常量。

- [ ] **Step 1: 扩展 war.go 错误码**

在 `internal/domain/war/war.go` 的错误码常量块末尾(`ErrorAPIResponseInvalid` 之后)追加:

```go
	ErrorClanNotFound        = "clan_not_found"
	ErrorPlayerNotFound      = "player_not_found"
```

- [ ] **Step 2: 创建 clan domain 包**

创建 `internal/domain/clan/clan.go`:

```go
package clan

type ClanDetail struct {
	Clan    ClanOverview       `json:"clan"`
	Members []ClanMemberSummary `json:"members"`
}

type ClanOverview struct {
	Tag           string `json:"tag"`
	Name          string `json:"name"`
	ClanLevel     int    `json:"clanLevel"`
	Description   string `json:"description,omitempty"`
	Members       int    `json:"members"`
	ClanPoints    int    `json:"clanPoints,omitempty"`
	WarWins       int    `json:"warWins,omitempty"`
	WarLosses     int    `json:"warLosses,omitempty"`
	WarTies       int    `json:"warTies,omitempty"`
	WarWinStreak  int    `json:"warWinStreak,omitempty"`
	WarFrequency  string `json:"warFrequency,omitempty"`
	Type          string `json:"type,omitempty"`
	IsWarLogPublic bool  `json:"isWarLogPublic,omitempty"`
	BadgeURLs     any    `json:"badgeUrls,omitempty"`
	Labels        []Label `json:"labels,omitempty"`
	Location      *Location `json:"location,omitempty"`
	WarLeague     *LeagueRef `json:"warLeague,omitempty"`
}

type ClanMemberSummary struct {
	Tag              string `json:"tag"`
	Name             string `json:"name"`
	TownHallLevel    int    `json:"townHallLevel"`
	ExpLevel         int    `json:"expLevel"`
	Role             string `json:"role"`
	Trophies         int    `json:"trophies"`
	ClanRank         int    `json:"clanRank"`
	Donations        int    `json:"donations"`
	DonationsReceived int   `json:"donationsReceived"`
	League           *LeagueRef `json:"league,omitempty"`
}

type PlayerOverview struct {
	Tag                string              `json:"tag"`
	Name               string              `json:"name"`
	TownHallLevel      int                 `json:"townHallLevel"`
	TownHallWeaponLevel int                 `json:"townHallWeaponLevel,omitempty"`
	ExpLevel           int                 `json:"expLevel"`
	Role               string              `json:"role,omitempty"`
	WarStars           int                 `json:"warStars,omitempty"`
	AttackWins         int                 `json:"attackWins,omitempty"`
	DefenseWins        int                 `json:"defenseWins,omitempty"`
	Trophies           int                 `json:"trophies"`
	BestTrophies       int                 `json:"bestTrophies,omitempty"`
	WarPreference      string              `json:"warPreference,omitempty"`
	Clan               *PlayerClanInfo     `json:"clan,omitempty"`
	Heroes             []HeroLevel         `json:"heroes,omitempty"`
	Achievements       []AchievementProgress `json:"achievements,omitempty"`
	Labels             []Label             `json:"labels,omitempty"`
	League             *LeagueRef          `json:"league,omitempty"`
}

type PlayerClanInfo struct {
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	ClanLevel int    `json:"clanLevel"`
	BadgeURLs any    `json:"badgeUrls,omitempty"`
}

type HeroLevel struct {
	Name      string `json:"name"`
	Level     int    `json:"level"`
	MaxLevel  int    `json:"maxLevel,omitempty"`
	Village   string `json:"village,omitempty"`
}

type AchievementProgress struct {
	Name      string `json:"name"`
	Stars     int    `json:"stars"`
	Target    int    `json:"target"`
	Value     int    `json:"value"`
	Village   string `json:"village,omitempty"`
	Info      string `json:"info,omitempty"`
}

type BattleLogSummary struct {
	Items  []BattleLogEntry `json:"items"`
	Paging Paging           `json:"paging"`
}

type BattleLogEntry struct {
	ArmyShareCode         string `json:"armyShareCode,omitempty"`
	Attack                bool   `json:"attack,omitempty"`
	BattleTime            int    `json:"battleTime,omitempty"`
	BattleTimestamp       string `json:"battleTimestamp,omitempty"`
	BattleType            string `json:"battleType,omitempty"`
	DestructionPercentage int    `json:"destructionPercentage,omitempty"`
	OpponentName          string `json:"opponentName,omitempty"`
	OpponentPlayerTag     string `json:"opponentPlayerTag,omitempty"`
	OpponentTownHallLevel int    `json:"opponentTownHallLevel,omitempty"`
	Stars                 int    `json:"stars,omitempty"`
}

type Label struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Location struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CountryCode string `json:"countryCode,omitempty"`
}

type LeagueRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Paging struct {
	Cursors struct {
		After  string `json:"after,omitempty"`
		Before string `json:"before,omitempty"`
	} `json:"cursors,omitempty"`
}
```

- [ ] **Step 3: 编译验证**

Run: `go build ./internal/domain/...`
Expected: 无输出(通过)

- [ ] **Step 4: Commit**

```bash
git add internal/domain/clan/clan.go internal/domain/war/war.go
git commit -m "Add clan/player domain types and not-found error codes"
```

---

### Task 2:Adapter 扩展 — Clan/Player 转换方法 + 错误映射

**Files:**
- Modify: `internal/infra/coc/client.go`(新增 4 个方法 + 1 个错误映射辅助 + 转换函数)
- Modify: `internal/infra/coc/client_test.go`(新增测试)

**Interfaces:**
- Consumes: `cocapi.Client.GetClan/GetClanMembers/GetPlayer/GetBattleLog`,`clan.ClanDetail` 等类型
- Produces: `(*Client).Clan(ctx, clanTag) (clan.ClanDetail, error)`、`(*Client).Player(ctx, playerTag) (clan.PlayerOverview, error)`、`(*Client).BattleLog(ctx, playerTag) (clan.BattleLogSummary, error)`

- [ ] **Step 1: 写 adapter 测试(先写测试)**

在 `internal/infra/coc/client_test.go` 末尾追加:

```go
func TestClanConvertsCocapiToDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/clans/%23AAA":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tag": "#AAA", "name": "My Clan", "clanLevel": 15,
				"members": 40, "warWins": 100, "warLosses": 20, "warTies": 5,
				"warWinStreak": 3, "type": "open", "isWarLogPublic": true,
				"description": "A clan",
			})
		case "/clans/%23AAA/members":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"tag": "#P1", "name": "Player1", "townHallLevel": 14, "role": "leader", "trophies": 5000, "clanRank": 1, "expLevel": 200, "donations": 1000, "donationsReceived": 500},
				},
				"paging": map[string]any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	detail, err := client.Clan(context.Background(), "#AAA")
	if err != nil {
		t.Fatalf("Clan error: %v", err)
	}
	if detail.Clan.Name != "My Clan" || detail.Clan.WarWins != 100 {
		t.Fatalf("clan: %+v", detail.Clan)
	}
	if len(detail.Members) != 1 || detail.Members[0].Name != "Player1" {
		t.Fatalf("members: %+v", detail.Members)
	}
	if detail.Members[0].TownHallLevel != 14 || detail.Members[0].Role != "leader" {
		t.Fatalf("member detail: %+v", detail.Members[0])
	}
}

func TestPlayerConvertsCocapiToDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag": "#P1", "name": "Player1", "townHallLevel": 14, "townHallWeaponLevel": 5,
			"expLevel": 200, "role": "leader", "warStars": 800, "attackWins": 300,
			"defenseWins": 50, "trophies": 5000, "bestTrophies": 5500,
			"clan": map[string]any{"tag": "#AAA", "name": "My Clan", "clanLevel": 15},
			"heroes": []map[string]any{{"name": "Barbarian King", "level": 80, "maxLevel": 90, "village": "home"}},
		})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	player, err := client.Player(context.Background(), "#P1")
	if err != nil {
		t.Fatalf("Player error: %v", err)
	}
	if player.Name != "Player1" || player.TownHallLevel != 14 {
		t.Fatalf("player: %+v", player)
	}
	if player.Clan == nil || player.Clan.Name != "My Clan" {
		t.Fatalf("clan: %+v", player.Clan)
	}
	if len(player.Heroes) != 1 || player.Heroes[0].Level != 80 {
		t.Fatalf("heroes: %+v", player.Heroes)
	}
}

func TestBattleLogConvertsItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"stars": 3, "destructionPercentage": 100, "opponentName": "Enemy", "battleType": "clanWar", "attack": true},
			},
		})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	log, err := client.BattleLog(context.Background(), "#P1")
	if err != nil {
		t.Fatalf("BattleLog error: %v", err)
	}
	if len(log.Items) != 1 || log.Items[0].Stars != 3 || log.Items[0].OpponentName != "Enemy" {
		t.Fatalf("items: %+v", log.Items)
	}
}

func TestClanNotFoundMapsToClanNotFoundCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"reason": "notFound", "message": "clan not found"})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	_, err := client.Clan(context.Background(), "#AAA")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorClanNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorClanNotFound)
	}
}

func TestPlayerNotFoundMapsToPlayerNotFoundCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"reason": "notFound", "message": "player not found"})
	}))
	defer server.Close()

	client := New(cocapi.Config{BaseURL: server.URL, APIToken: "secret", Timeout: time.Second})
	_, err := client.Player(context.Background(), "#P1")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorPlayerNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorPlayerNotFound)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/infra/coc/ -run "TestClan|TestPlayer|TestBattleLog" -v`
Expected: FAIL — `client.Clan undefined`、`client.Player undefined`、`client.BattleLog undefined`

- [ ] **Step 3: 实现 adapter 方法和转换函数**

在 `internal/infra/coc/client.go` 追加(在 `CWLGroup` 方法之后、`toDomainCurrentWar` 之前):

```go
func (c *Client) Clan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	clan, err := c.api.GetClan(ctx, clanTag)
	if err != nil {
		return clandomain.ClanDetail{}, mapErrorWithNotFound(err, wardomain.ErrorClanNotFound)
	}
	membersResp, err := c.api.GetClanMembers(ctx, clanTag, cocapi.QueryGetClanMembers{})
	if err != nil {
		return clandomain.ClanDetail{}, mapErrorWithNotFound(err, wardomain.ErrorClanNotFound)
	}
	return clandomain.ClanDetail{
		Clan:    toDomainClanOverview(clan),
		Members: toDomainClanMembers(membersResp.Items),
	}, nil
}

func (c *Client) Player(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	player, err := c.api.GetPlayer(ctx, playerTag)
	if err != nil {
		return clandomain.PlayerOverview{}, mapErrorWithNotFound(err, wardomain.ErrorPlayerNotFound)
	}
	return toDomainPlayerOverview(player), nil
}

func (c *Client) BattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	log, err := c.api.GetBattleLog(ctx, playerTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, mapErrorWithNotFound(err, wardomain.ErrorPlayerNotFound)
	}
	return toDomainBattleLog(log), nil
}
```

在 `client.go` import 块追加 `"github.com/ww1489/WarSpark/internal/domain/clan" as clandomain`。修改 import 为:

```go
import (
	"context"
	"errors"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)
```

在 `client.go` 末尾(`mapError` 函数之后)追加转换函数和错误映射辅助:

```go
func mapErrorWithNotFound(err error, notFoundCode string) error {
	mapped := mapError(err)
	var warErr wardomain.Error
	if errors.As(mapped, &warErr) && warErr.Code == wardomain.ErrorWarNotFound {
		return wardomain.WrapError(notFoundCode, err.Error(), err)
	}
	return mapped
}

func toDomainClanOverview(c cocapi.Clan) clandomain.ClanOverview {
	overview := clandomain.ClanOverview{
		Tag:             c.Tag,
		Name:            c.Name,
		ClanLevel:       c.ClanLevel,
		Description:     c.Description,
		Members:         c.Members,
		ClanPoints:      c.ClanPoints,
		WarWins:         c.WarWins,
		WarLosses:       c.WarLosses,
		WarTies:         c.WarTies,
		WarWinStreak:    c.WarWinStreak,
		WarFrequency:    c.WarFrequency,
		Type:            c.Type,
		IsWarLogPublic:  c.IsWarLogPublic,
		BadgeURLs:       c.BadgeURLs,
		Labels:          toDomainLabels(c.Labels),
	}
	if c.Location.ID != 0 {
		loc := toDomainLocation(c.Location)
		overview.Location = &loc
	}
	if c.WarLeague.ID != 0 {
		league := toDomainLeagueRef(c.WarLeague)
		overview.WarLeague = &league
	}
	return overview
}

func toDomainClanMembers(members []cocapi.ClanMember) []clandomain.ClanMemberSummary {
	result := make([]clandomain.ClanMemberSummary, 0, len(members))
	for _, m := range members {
		summary := clandomain.ClanMemberSummary{
			Tag:               m.Tag,
			Name:              m.Name,
			TownHallLevel:     m.TownHallLevel,
			ExpLevel:          m.ExpLevel,
			Role:              m.Role,
			Trophies:          m.Trophies,
			ClanRank:          m.ClanRank,
			Donations:         m.Donations,
			DonationsReceived: m.DonationsReceived,
		}
		if m.League.ID != 0 {
			league := toDomainLeagueRef(cocapi.League(m.League))
			summary.League = &league
		}
		result = append(result, summary)
	}
	return result
}

func toDomainPlayerOverview(p cocapi.Player) clandomain.PlayerOverview {
	overview := clandomain.PlayerOverview{
		Tag:                 p.Tag,
		Name:                p.Name,
		TownHallLevel:       p.TownHallLevel,
		TownHallWeaponLevel: p.TownHallWeaponLevel,
		ExpLevel:            p.ExpLevel,
		Role:                p.Role,
		WarStars:            p.WarStars,
		AttackWins:          p.AttackWins,
		DefenseWins:         p.DefenseWins,
		Trophies:            p.Trophies,
		BestTrophies:        p.BestTrophies,
		WarPreference:       p.WarPreference,
		Labels:              toDomainLabels(p.Labels),
		Heroes:              toDomainHeroLevels(p.Heroes),
		Achievements:        toDomainAchievements(p.Achievements),
	}
	if p.Clan.Tag != "" {
		overview.Clan = &clandomain.PlayerClanInfo{
			Tag:       p.Clan.Tag,
			Name:      p.Clan.Name,
			ClanLevel: p.Clan.ClanLevel,
			BadgeURLs: p.Clan.BadgeURLs,
		}
	}
	if p.League.ID != 0 {
		league := toDomainLeagueRef(p.League)
		overview.League = &league
	}
	return overview
}

func toDomainBattleLog(log cocapi.BattleLogEntryListResponse) clandomain.BattleLogSummary {
	items := make([]clandomain.BattleLogEntry, 0, len(log.Items))
	for _, e := range log.Items {
		items = append(items, clandomain.BattleLogEntry{
			ArmyShareCode:         e.ArmyShareCode,
			Attack:                e.Attack,
			BattleTime:            e.BattleTime,
			BattleTimestamp:       e.BattleTimestamp,
			BattleType:            e.BattleType,
			DestructionPercentage: e.DestructionPercentage,
			OpponentName:          e.OpponentName,
			OpponentPlayerTag:     e.OpponentPlayerTag,
			OpponentTownHallLevel: e.OpponentTownHallLevel,
			Stars:                 e.Stars,
		})
	}
	return clandomain.BattleLogSummary{Items: items, Paging: clandomain.Paging{}}
}

func toDomainLabels(labels cocapi.LabelList) []clandomain.Label {
	result := make([]clandomain.Label, 0, len(labels))
	for _, l := range labels {
		result = append(result, clandomain.Label{
			ID:   l.ID,
			Name: string(l.Name),
		})
	}
	return result
}

func toDomainLocation(loc cocapi.Location) clandomain.Location {
	return clandomain.Location{
		ID:          loc.ID,
		Name:        string(loc.Name),
		CountryCode: loc.CountryCode,
	}
}

func toDomainLeagueRef(league cocapi.League) clandomain.LeagueRef {
	return clandomain.LeagueRef{
		ID:   league.ID,
		Name: string(league.Name),
	}
}

func toDomainHeroLevels(heroes cocapi.PlayerItemLevelList) []clandomain.HeroLevel {
	result := make([]clandomain.HeroLevel, 0, len(heroes))
	for _, h := range heroes {
		result = append(result, clandomain.HeroLevel{
			Name:     string(h.Name),
			Level:    h.Level,
			MaxLevel: h.MaxLevel,
			Village:  h.Village,
		})
	}
	return result
}

func toDomainAchievements(achievements cocapi.PlayerAchievementProgressList) []clandomain.AchievementProgress {
	result := make([]clandomain.AchievementProgress, 0, len(achievements))
	for _, a := range achievements {
		result = append(result, clandomain.AchievementProgress{
			Name:    string(a.Name),
			Stars:   a.Stars,
			Target:  a.Target,
			Value:   a.Value,
			Village: a.Village,
			Info:    string(a.Info),
		})
	}
	return result
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/infra/coc/ -v`
Expected: 全部 PASS(原有 7 个 + 新增 5 个 = 12 个测试)

- [ ] **Step 5: Commit**

```bash
git add internal/infra/coc/client.go internal/infra/coc/client_test.go
git commit -m "Add clan/player/battle-log adapter methods with context-aware error mapping"
```

---

### Task 3:Redis 缓存 — ClanCache

**Files:**
- Create: `internal/infra/redis/clan_cache.go`

**Interfaces:**
- Consumes: `clan.ClanDetail`、`clan.PlayerOverview`、`clan.BattleLogSummary` 类型
- Produces: `ClanCache` struct,`NewClanCache(client *goredis.Client) *ClanCache`,`GetClan/SetClan/GetPlayer/SetPlayer/GetBattleLog/SetBattleLog` 方法

- [ ] **Step 1: 创建 clan_cache.go**

创建 `internal/infra/redis/clan_cache.go`,模式与 `war_cache.go` 一致:

```go
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type ClanCache struct {
	client *goredis.Client
}

func NewClanCache(client *goredis.Client) *ClanCache {
	return &ClanCache{client: client}
}

func (c *ClanCache) GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error) {
	if c == nil || c.client == nil {
		return clandomain.ClanDetail{}, false, nil
	}
	data, err := c.client.Get(ctx, clanDetailKey(clanTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return clandomain.ClanDetail{}, false, nil
		}
		return clandomain.ClanDetail{}, false, err
	}
	var detail clandomain.ClanDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return clandomain.ClanDetail{}, false, err
	}
	return detail, true, nil
}

func (c *ClanCache) SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, clanDetailKey(clanTag), data, ttl).Err()
}

func (c *ClanCache) GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error) {
	if c == nil || c.client == nil {
		return clandomain.PlayerOverview{}, false, nil
	}
	data, err := c.client.Get(ctx, playerOverviewKey(playerTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return clandomain.PlayerOverview{}, false, nil
		}
		return clandomain.PlayerOverview{}, false, err
	}
	var player clandomain.PlayerOverview
	if err := json.Unmarshal(data, &player); err != nil {
		return clandomain.PlayerOverview{}, false, err
	}
	return player, true, nil
}

func (c *ClanCache) SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(player)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, playerOverviewKey(playerTag), data, ttl).Err()
}

func (c *ClanCache) GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error) {
	if c == nil || c.client == nil {
		return clandomain.BattleLogSummary{}, false, nil
	}
	data, err := c.client.Get(ctx, battleLogKey(playerTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return clandomain.BattleLogSummary{}, false, nil
		}
		return clandomain.BattleLogSummary{}, false, err
	}
	var log clandomain.BattleLogSummary
	if err := json.Unmarshal(data, &log); err != nil {
		return clandomain.BattleLogSummary{}, false, err
	}
	return log, true, nil
}

func (c *ClanCache) SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, battleLogKey(playerTag), data, ttl).Err()
}

func clanDetailKey(clanTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(clanTag))
	tag = strings.TrimPrefix(tag, "#")
	return "clan:detail:" + tag
}

func playerOverviewKey(playerTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(playerTag))
	tag = strings.TrimPrefix(tag, "#")
	return "player:overview:" + tag
}

func battleLogKey(playerTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(playerTag))
	tag = strings.TrimPrefix(tag, "#")
	return "player:battlelog:" + tag
}
```

- [ ] **Step 2: 编译验证**

Run: `go build ./internal/infra/redis/...`
Expected: 无输出(通过)

- [ ] **Step 3: Commit**

```bash
git add internal/infra/redis/clan_cache.go
git commit -m "Add Redis cache for clan/player overview and battle log"
```

---

### Task 4:ClanService + 测试

**Files:**
- Create: `internal/service/clan_service.go`
- Create: `internal/service/clan_service_test.go`

**Interfaces:**
- Consumes: `ClanAPIClient` interface(`Clan(ctx, clanTag) (clan.ClanDetail, error)`),`ClanCache` interface,`NormalizeClanTag`(已有)
- Produces: `ClanService` struct,`NewClanService(client ClanAPIClient, cache ClanCache, ttl time.Duration) *ClanService`,`FetchClan(ctx, clanTag) (clan.ClanDetail, error)` 方法

- [ ] **Step 1: 写 ClanService 测试(先写测试)**

创建 `internal/service/clan_service_test.go`:

```go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeClanAPIClient struct {
	detail    clandomain.ClanDetail
	err       error
	gotTag    string
}

func (f *fakeClanAPIClient) Clan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	f.gotTag = clanTag
	if f.err != nil {
		return clandomain.ClanDetail{}, f.err
	}
	return f.detail, nil
}

func TestClanServiceFetchReturnsCached(t *testing.T) {
	cached := clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Cached"}}
	cache := &fakeClanCache{clanDetail: cached, clanHit: true}
	svc := NewClanService(&fakeClanAPIClient{}, cache, 5*time.Minute)

	detail, err := svc.FetchClan(context.Background(), "#AAA")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if detail.Clan.Name != "Cached" {
		t.Fatalf("got %q, want Cached", detail.Clan.Name)
	}
}

func TestClanServiceFetchCallsAPIOnMiss(t *testing.T) {
	api := &fakeClanAPIClient{detail: clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Fresh"}}}
	cache := &fakeClanCache{clanHit: false}
	svc := NewClanService(api, cache, 5*time.Minute)

	detail, err := svc.FetchClan(context.Background(), "#AAA")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if detail.Clan.Name != "Fresh" {
		t.Fatalf("got %q, want Fresh", detail.Clan.Name)
	}
	if api.gotTag != "#AAA" {
		t.Fatalf("api got tag %q, want #AAA", api.gotTag)
	}
	if cache.setClanTag != "#AAA" {
		t.Fatalf("cache set tag %q, want #AAA", cache.setClanTag)
	}
}

func TestClanServiceFetchPropagatesNotFound(t *testing.T) {
	api := &fakeClanAPIClient{err: wardomain.NewError(wardomain.ErrorClanNotFound, "not found")}
	cache := &fakeClanCache{clanHit: false}
	svc := NewClanService(api, cache, 5*time.Minute)

	_, err := svc.FetchClan(context.Background(), "#AAA")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorClanNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorClanNotFound)
	}
}

func TestClanServiceFetchRejectsInvalidTag(t *testing.T) {
	svc := NewClanService(&fakeClanAPIClient{}, &fakeClanCache{}, 5*time.Minute)
	_, err := svc.FetchClan(context.Background(), "!!!")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorInvalidTag {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorInvalidTag)
	}
}

type fakeClanCache struct {
	clanDetail    clandomain.ClanDetail
	clanHit       bool
	setClanTag    string
	playerOverview clandomain.PlayerOverview
	playerHit      bool
	setPlayerTag   string
	battleLog      clandomain.BattleLogSummary
	battleLogHit   bool
	setBattleTag   string
}

func (f *fakeClanCache) GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error) {
	return f.clanDetail, f.clanHit, nil
}
func (f *fakeClanCache) SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail, ttl time.Duration) error {
	f.setClanTag = clanTag
	return nil
}
func (f *fakeClanCache) GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error) {
	return f.playerOverview, f.playerHit, nil
}
func (f *fakeClanCache) SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview, ttl time.Duration) error {
	f.setPlayerTag = playerTag
	return nil
}
func (f *fakeClanCache) GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error) {
	return f.battleLog, f.battleLogHit, nil
}
func (f *fakeClanCache) SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary, ttl time.Duration) error {
	f.setBattleTag = playerTag
	return nil
}

```
<br>

**注意:** `fakeClanCache` 同时满足 `ClanCache`(Task 4)和 `PlayerCache`(Task 5)接口,因为它们的方法集是子集关系。Task 5 的 `fakePlayerAPIClient` 直接引用此 fake。

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/service/ -run TestClanService -v`
Expected: FAIL — `NewClanService undefined`

- [ ] **Step 3: 实现 ClanService**

创建 `internal/service/clan_service.go`:

```go
package service

import (
	"context"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type ClanAPIClient interface {
	Clan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error)
}

type ClanCache interface {
	GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error)
	SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail, ttl time.Duration) error
}

type ClanService struct {
	client   ClanAPIClient
	cache    ClanCache
	cacheTTL time.Duration
}

func NewClanService(client ClanAPIClient, cache ClanCache, ttl time.Duration) *ClanService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &ClanService{client: client, cache: cache, cacheTTL: ttl}
}

func (s *ClanService) FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return clandomain.ClanDetail{}, err
	}

	if s.cache != nil {
		detail, ok, err := s.cache.GetClan(ctx, normalizedTag)
		if err == nil && ok {
			return detail, nil
		}
	}

	detail, err := s.client.Clan(ctx, normalizedTag)
	if err != nil {
		return clandomain.ClanDetail{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetClan(ctx, normalizedTag, detail, s.cacheTTL)
	}
	return detail, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/service/ -run TestClanService -v`
Expected: 4 个测试全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/clan_service.go internal/service/clan_service_test.go
git commit -m "Add ClanService with Redis cache and tag validation"
```

---

### Task 5:PlayerService + 测试

**Files:**
- Create: `internal/service/player_service.go`
- Create: `internal/service/player_service_test.go`

**Interfaces:**
- Consumes: `PlayerAPIClient` interface(`Player` + `BattleLog` 方法),`PlayerCache` interface(复用 Task 4 的 `fakeClanCache`)
- Produces: `PlayerService`,`NewPlayerService(client PlayerAPIClient, cache PlayerCache, ttl time.Duration)`,`FetchPlayer`、`FetchBattleLog` 方法

- [ ] **Step 1: 写 PlayerService 测试**

创建 `internal/service/player_service_test.go`:

```go
package service

import (
	"context"
	"testing"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakePlayerAPIClient struct {
	player    clandomain.PlayerOverview
	battleLog clandomain.BattleLogSummary
	err       error
	gotTag    string
}

func (f *fakePlayerAPIClient) Player(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	f.gotTag = playerTag
	if f.err != nil {
		return clandomain.PlayerOverview{}, f.err
	}
	return f.player, nil
}

func (f *fakePlayerAPIClient) BattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	f.gotTag = playerTag
	if f.err != nil {
		return clandomain.BattleLogSummary{}, f.err
	}
	return f.battleLog, nil
}

func TestPlayerServiceFetchReturnsCached(t *testing.T) {
	cached := clandomain.PlayerOverview{Name: "Cached"}
	cache := &fakeClanCache{playerOverview: cached, playerHit: true}
	svc := NewPlayerService(&fakePlayerAPIClient{}, cache, 5*time.Minute)

	player, err := svc.FetchPlayer(context.Background(), "#P1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if player.Name != "Cached" {
		t.Fatalf("got %q, want Cached", player.Name)
	}
}

func TestPlayerServiceFetchCallsAPIOnMiss(t *testing.T) {
	api := &fakePlayerAPIClient{player: clandomain.PlayerOverview{Name: "Fresh"}}
	cache := &fakeClanCache{playerHit: false}
	svc := NewPlayerService(api, cache, 5*time.Minute)

	player, err := svc.FetchPlayer(context.Background(), "#P1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if player.Name != "Fresh" {
		t.Fatalf("got %q, want Fresh", player.Name)
	}
	if cache.setPlayerTag != "#P1" {
		t.Fatalf("cache set tag %q, want #P1", cache.setPlayerTag)
	}
}

func TestPlayerServiceFetchBattleLog(t *testing.T) {
	api := &fakePlayerAPIClient{battleLog: clandomain.BattleLogSummary{Items: []clandomain.BattleLogEntry{{Stars: 3}}}}
	cache := &fakeClanCache{battleLogHit: false}
	svc := NewPlayerService(api, cache, 5*time.Minute)

	log, err := svc.FetchBattleLog(context.Background(), "#P1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(log.Items) != 1 || log.Items[0].Stars != 3 {
		t.Fatalf("items: %+v", log.Items)
	}
}

func TestPlayerServiceFetchPropagatesNotFound(t *testing.T) {
	api := &fakePlayerAPIClient{err: wardomain.NewError(wardomain.ErrorPlayerNotFound, "not found")}
	cache := &fakeClanCache{playerHit: false}
	svc := NewPlayerService(api, cache, 5*time.Minute)

	_, err := svc.FetchPlayer(context.Background(), "#P1")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorPlayerNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorPlayerNotFound)
	}
}

func TestPlayerServiceFetchRejectsInvalidTag(t *testing.T) {
	svc := NewPlayerService(&fakePlayerAPIClient{}, &fakeClanCache{}, 5*time.Minute)
	_, err := svc.FetchPlayer(context.Background(), "!!!")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorInvalidTag {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorInvalidTag)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/service/ -run TestPlayerService -v`
Expected: FAIL — `NewPlayerService undefined`

- [ ] **Step 3: 实现 PlayerService**

创建 `internal/service/player_service.go`:

```go
package service

import (
	"context"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type PlayerAPIClient interface {
	Player(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error)
	BattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error)
}

type PlayerCache interface {
	GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error)
	SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview, ttl time.Duration) error
	GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error)
	SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary, ttl time.Duration) error
}

type PlayerService struct {
	client   PlayerAPIClient
	cache    PlayerCache
	cacheTTL time.Duration
}

func NewPlayerService(client PlayerAPIClient, cache PlayerCache, ttl time.Duration) *PlayerService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &PlayerService{client: client, cache: cache, cacheTTL: ttl}
}

func (s *PlayerService) FetchPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	normalizedTag, err := NormalizeClanTag(playerTag)
	if err != nil {
		return clandomain.PlayerOverview{}, err
	}

	if s.cache != nil {
		player, ok, err := s.cache.GetPlayer(ctx, normalizedTag)
		if err == nil && ok {
			return player, nil
		}
	}

	player, err := s.client.Player(ctx, normalizedTag)
	if err != nil {
		return clandomain.PlayerOverview{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetPlayer(ctx, normalizedTag, player, s.cacheTTL)
	}
	return player, nil
}

func (s *PlayerService) FetchBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	normalizedTag, err := NormalizeClanTag(playerTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, err
	}

	if s.cache != nil {
		log, ok, err := s.cache.GetBattleLog(ctx, normalizedTag)
		if err == nil && ok {
			return log, nil
		}
	}

	log, err := s.client.BattleLog(ctx, normalizedTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetBattleLog(ctx, normalizedTag, log, s.cacheTTL)
	}
	return log, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/service/ -run TestPlayerService -v`
Expected: 5 个测试全部 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/player_service.go internal/service/player_service_test.go
git commit -m "Add PlayerService with player overview and battle log cache"
```

---

### Task 6:Controller + 路由注册

**Files:**
- Create: `internal/controller/clan_controller.go`
- Create: `internal/controller/player_controller.go`
- Modify: `internal/api/v1/routes.go`(注册 3 个路由)
- Modify: `internal/api/v1/routes_test.go`(适配签名,如果需要)

**Interfaces:**
- Consumes: `ClanService.FetchClan`、`PlayerService.FetchPlayer/FetchBattleLog`,现有 `failWar` 错误处理模式
- Produces: 3 个 HTTP 端点:`GET /api/v1/clans/:tag`、`GET /api/v1/players/:tag`、`GET /api/v1/players/:tag/battle-log`

- [ ] **Step 1: 创建 ClanController**

创建 `internal/controller/clan_controller.go`:

```go
package controller

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type ClanReader interface {
	FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error)
}

type ClanController struct {
	service ClanReader
}

func NewClanController(service ClanReader) *ClanController {
	return &ClanController{service: service}
}

func (c *ClanController) GetClan(ctx *gin.Context) {
	clanTag := ctx.Param("tag")
	if clanTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "tag is required"))
		return
	}

	detail, err := c.service.FetchClan(ctx.Request.Context(), clanTag)
	if err != nil {
		failClan(ctx, err)
		return
	}
	utils.OK(ctx, detail)
}

func failClan(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, 422, int(utils.ErrInvalidField), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, 503, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorClanNotFound:
			utils.JSON(ctx, 404, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, 403, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, 502, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
```

- [ ] **Step 2: 创建 PlayerController**

创建 `internal/controller/player_controller.go`:

```go
package controller

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type PlayerReader interface {
	FetchPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error)
	FetchBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error)
}

type PlayerController struct {
	service PlayerReader
}

func NewPlayerController(service PlayerReader) *PlayerController {
	return &PlayerController{service: service}
}

func (c *PlayerController) GetPlayer(ctx *gin.Context) {
	playerTag := ctx.Param("tag")
	if playerTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "tag is required"))
		return
	}

	player, err := c.service.FetchPlayer(ctx.Request.Context(), playerTag)
	if err != nil {
		failPlayer(ctx, err)
		return
	}
	utils.OK(ctx, player)
}

func (c *PlayerController) GetBattleLog(ctx *gin.Context) {
	playerTag := ctx.Param("tag")
	if playerTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "tag is required"))
		return
	}

	log, err := c.service.FetchBattleLog(ctx.Request.Context(), playerTag)
	if err != nil {
		failPlayer(ctx, err)
		return
	}
	utils.OK(ctx, log)
}

func failPlayer(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, 422, int(utils.ErrInvalidField), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, 503, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorPlayerNotFound:
			utils.JSON(ctx, 404, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, 403, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, 502, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
```

- [ ] **Step 3: 注册路由**

在 `internal/api/v1/routes.go` 的 `SetupRoutes` 函数中,在 `warController` 构造之后(`warController := controller.NewWarController(warService)` 行之后),追加。复用已有的 `warAPIClient`(`*infracoc.Client` 现在也实现了 `ClanAPIClient`/`PlayerAPIClient` 接口):

```go
	clanCache := infraredis.NewClanCache(runtimeConfig.Redis)
	clanService := service.NewClanService(warAPIClient, clanCache, 5*time.Minute)
	clanController := controller.NewClanController(clanService)
	playerService := service.NewPlayerService(warAPIClient, clanCache, 5*time.Minute)
	playerController := controller.NewPlayerController(playerService)
```

在 `api` group 的路由注册块中(在 `api.GET("/war/snapshots/..."` 之后)追加:

```go
		api.GET("/clans/:tag", clanController.GetClan)
		api.GET("/players/:tag", playerController.GetPlayer)
		api.GET("/players/:tag/battle-log", playerController.GetBattleLog)
```

在 `routes.go` import 块追加 `"time"`(现有 import 没有 time 包,需手动添加)。

- [ ] **Step 4: 编译验证**

Run: `go build ./...`
Expected: 无输出(通过)

- [ ] **Step 5: 运行全量测试**

Run: `make check`
Expected: fmt 无输出、vet 无输出、test 全绿、build 通过

- [ ] **Step 6: Commit**

```bash
git add internal/controller/clan_controller.go internal/controller/player_controller.go internal/api/v1/routes.go
git commit -m "Add clan/player controllers and register 3 public routes"
```

---

### Task 7:Controller 测试

**Files:**
- Create: `internal/controller/clan_controller_test.go`
- Create: `internal/controller/player_controller_test.go`

**Interfaces:**
- Consumes: `ClanController`、`PlayerController`、`clandomain.ClanDetail`/`PlayerOverview`/`BattleLogSummary` 类型

- [ ] **Step 1: 写 ClanController 测试**

创建 `internal/controller/clan_controller_test.go`:

```go
package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeClanService struct {
	detail clandomain.ClanDetail
	err    error
}

func (f *fakeClanService) FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	if f.err != nil {
		return clandomain.ClanDetail{}, f.err
	}
	return f.detail, nil
}

func TestClanControllerGetClanReturnsDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeClanService{detail: clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Test Clan"}}}
	ctrl := NewClanController(svc)

	router := gin.New()
	router.GET("/clans/:tag", ctrl.GetClan)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/%23AAA", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	clan := data["clan"].(map[string]any)
	if clan["name"] != "Test Clan" {
		t.Fatalf("name = %v, want Test Clan", clan["name"])
	}
}

func TestClanControllerGetClanReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeClanService{err: wardomain.NewError(wardomain.ErrorClanNotFound, "not found")}
	ctrl := NewClanController(svc)

	router := gin.New()
	router.GET("/clans/:tag", ctrl.GetClan)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/%23AAA", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestClanControllerGetClanRejectsMissingTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewClanController(&fakeClanService{})

	router := gin.New()
	router.GET("/clans/:tag", ctrl.GetClan)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d (gin router no match expected)", rec.Code)
	}
}
```

- [ ] **Step 2: 写 PlayerController 测试**

创建 `internal/controller/player_controller_test.go`:

```go
package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakePlayerService struct {
	player    clandomain.PlayerOverview
	battleLog clandomain.BattleLogSummary
	err       error
}

func (f *fakePlayerService) FetchPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	if f.err != nil {
		return clandomain.PlayerOverview{}, f.err
	}
	return f.player, nil
}

func (f *fakePlayerService) FetchBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	if f.err != nil {
		return clandomain.BattleLogSummary{}, f.err
	}
	return f.battleLog, nil
}

func TestPlayerControllerGetPlayerReturnsOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePlayerService{player: clandomain.PlayerOverview{Name: "TestPlayer"}}
	ctrl := NewPlayerController(svc)

	router := gin.New()
	router.GET("/players/:tag", ctrl.GetPlayer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/%23P1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["name"] != "TestPlayer" {
		t.Fatalf("name = %v, want TestPlayer", data["name"])
	}
}

func TestPlayerControllerGetBattleLogReturnsItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePlayerService{battleLog: clandomain.BattleLogSummary{Items: []clandomain.BattleLogEntry{{Stars: 3}}}}
	ctrl := NewPlayerController(svc)

	router := gin.New()
	router.GET("/players/:tag/battle-log", ctrl.GetBattleLog)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/%23P1/battle-log", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestPlayerControllerGetPlayerReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePlayerService{err: wardomain.NewError(wardomain.ErrorPlayerNotFound, "not found")}
	ctrl := NewPlayerController(svc)

	router := gin.New()
	router.GET("/players/:tag", ctrl.GetPlayer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/%23P1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
```

- [ ] **Step 3: 运行测试确认通过**

Run: `go test ./internal/controller/ -run "TestClanController|TestPlayerController" -v`
Expected: 6 个测试全部 PASS

- [ ] **Step 4: 运行全量验证**

Run: `make check`
Expected: 全绿

- [ ] **Step 5: Commit**

```bash
git add internal/controller/clan_controller_test.go internal/controller/player_controller_test.go
git commit -m "Add controller tests for clan and player endpoints"
```

---

## 完成检查清单

- [ ] 3 个接口路由已注册:`GET /clans/:tag`、`GET /players/:tag`、`GET /players/:tag/battle-log`
- [ ] adapter 4 个转换方法 + context-aware 错误映射
- [ ] ClanService + PlayerService 含缓存
- [ ] controller 含错误码 HTTP 映射
- [ ] `make check` 全绿
- [ ] cocapi 端点使用:2 → 6 / 35(新增 GetClan/GetClanMembers/GetPlayer/GetBattleLog)
