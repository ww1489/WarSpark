# CoC API 功能扩展设计:CWL 情报增强 + 部族/玩家概览 + 排名/联赛/标签

> 日期:2026-06-26
> 模块:github.com/ww1489/WarSpark
> 状态:待实施
> 端点使用:7 → 23 / 35(66%)

## 1. 背景与目标

### 1.1 现状

`pkg/cocapi` 已封装 Clash of Clans 官方 API 全部 35 个端点(2026-06-25 重构完成),但业务层只使用了 2 个(`GetCurrentWar` `GetClanWarLeagueGroup`)。剩余 33 个端点的封装价值未变现。

### 1.2 目标

基于官方 API 实现一批功能,优先覆盖与 WarSpark "战争情报 + 找阵" 主线相关的端点。本设计是第一个子项目,聚焦:

- **CWL 情报增强**:组排名 / 7 轮对战 / 玩家排名 / 奖牌预估,并落库支撑历史浏览。
- **部族 / 玩家概览**:补齐 P2-9 契约,为战争情报页提供基础上下文。
- **排名榜**:全球 + 地区维度的部族 / 玩家 / 都城 / 建筑基地排名。
- **联赛元数据**:联赛列表 / 详情 / 赛季 / 层级 / 历史排名的完整浏览。
- **标签**:部族 / 玩家标签列表,用于详情页展示和筛选器选项。

### 1.3 竞品参考

竞品 [War Report](https://warreport.app/about/docs/) 的差异化启发:

| 竞品功能 | 竞品做法 | WarSpark 做法 | 差异化 |
| --- | --- | --- | --- |
| CWL 组 | 组排名 + 7 轮 + 玩家排名 + 奖牌 | 同上 + 每轮敌方目标可进入找阵 | 竞品止于查看,WarSpark 接找阵闭环 |
| 部族页 | 详情 + 成员 + 战 log + 统计 + 都城 | 详情 + 成员 + 战 log(战争维度) | 砍掉都城 / 统计(MVP 聚焦战争) |
| 玩家页 | 详情 + 成就 + 对战日志 + 攻击档案 + 统计 + 配兵 | 详情 + 成就 + 战争攻击 / 防守档案 | 砍掉配兵统计(MVP 不做配兵),保留战争档案 |
| 历史战争 | 靠 ClashKing Discord bot(非官方 API) | 靠自存 war_snapshots | 自主可控,不依赖第三方 |

### 1.4 竞品 API 评估

`api.warreport.app` 公开 API 只提供 Legend League 赛季快照(全球排名、热门配兵、英雄装备使用统计),无 key。**对 WarSpark 使用价值低** — WarSpark 不做 Legend League 工具或配兵推荐。但其 OpenAPI 3.1 规范发布模式可参考。

### 1.5 关键限制

官方 API **不提供历史战争 / 历史 CWL 数据** — 竞品的"过去战争"靠 ClashKing Discord bot。WarSpark 要做历史浏览,只能靠自己的 `war_snapshots` 表积累(每次拉取 current war 时落库)。这决定了"历史 / 统计类"功能需要数据积累期,不能立刻见效。

## 2. 架构概览

### 2.1 数据流

沿用现有 adapter 模式:

```text
controller → service → adapter(infracoc.Client) → cocapi.Client
                    ↓
               repository(落库 + 聚合查询)
                    ↓
               MySQL(cwl_groups / cwl_rounds / war_attacks + 复用 war_snapshots / war_members)
                    ↓
               Redis 缓存(部族 / 玩家概览 TTL)
```

### 2.2 扩展点(不改动现有层)

- **`infracoc.Client`**:新增 21 个转换方法(对应 21 个新端点;`GetLocation` 合并到列表、`GetLeagueGroup` 合并到赛季,故 23 端点 → 21 方法),实现新增的 service 接口。
- **`service`**:新增 `ClanService` `PlayerService` `RankingService` `LeagueService` `LabelService`,`WarService` 扩展 CWL 增强。
- **`repository`**:新增 `CWLRepository`(组 / 轮落库) + `WarAttackRepository`(攻击详情),扩展现有 `WarRepository`。

### 2.3 落库策略

**被动落库**:用户查看 CWL / 战争时实时 upsert 到数据库。CWL 进行中每次查看更新快照,状态变 `ended` 后冻结保留。无需后台任务,复用现有"查看即落库"模式(当前 war 已这样做)。

### 2.4 不做(明确边界)

- ❌ 都城突袭(`GetCapitalRaidSeasons`)— 与战争情报无关,后续子项目。
- ❌ Gold Pass(`GetCurrentGoldPassSeason`)— 与战争情报无关。
- ❌ `SearchClans` — 搜索功能,与 tag 查询维度不同,后续。
- ❌ `GetClanWarLog` — 战争日志摘要,不含攻击详情,后续"历史战争浏览"子项目。
- ❌ `VerifyToken` — 账号系统,MVP 不做。
- ❌ 玩家配兵统计 — MVP 不做配兵识别。
- ❌ 玩家对战日志落库 — 数据量大,只缓存不存。
- ❌ 收藏 / 书签 — MVP 不做用户系统。

## 3. 数据模型

### 3.1 新增表(3 张)

**`cwl_groups`** — CWL 组主表

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | VARCHAR(36) PK | UUID |
| `clan_tag` | VARCHAR(20) | 我方部落 tag |
| `season` | VARCHAR(20) | 赛季标识 |
| `state` | VARCHAR(20) | 组状态(notInLeague / preparation / inWar / ended) |
| `group_size` | INT | 组内部落数(通常 8) |
| `clans_json` | JSON | 8 个部族信息(tag / name / level / members / warWins) |
| `rounds_json` | JSON | 7 轮 warTags 数组 |
| `fetched_at` | DATETIME | 获取时间 |
| `created_at` | DATETIME | 创建时间 |
| `updated_at` | DATETIME | 更新时间 |

索引:`UNIQUE(clan_tag, season)`(同一部落同一赛季只存一份)、`INDEX(season)`。

用 JSON 存 clans / rounds,因为结构固定且只读,不值得拆表。

**`cwl_rounds`** — 轮次对战映射(用于历史查询)

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | VARCHAR(36) PK | UUID |
| `cwl_group_id` | VARCHAR(36) FK | 关联 cwl_groups.id,ON DELETE CASCADE |
| `round_number` | INT | 轮次序号 1-7 |
| `war_tag` | VARCHAR(20) | 该轮对战 warTag |
| `war_snapshot_id` | VARCHAR(36) FK nullable | 落库后关联 war_snapshots.id |
| `state` | VARCHAR(20) | 该轮状态 |

索引:`UNIQUE(cwl_group_id, round_number)`、`INDEX(war_tag)`。

**`war_attacks`** — 攻击详情(支撑玩家战争档案)

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | VARCHAR(36) PK | UUID |
| `war_snapshot_id` | VARCHAR(36) FK | 关联 war_snapshots.id,ON DELETE CASCADE |
| `war_member_id` | VARCHAR(36) FK | 关联 war_members.id,ON DELETE CASCADE |
| `attacker_tag` | VARCHAR(20) | 攻击方玩家 tag |
| `defender_tag` | VARCHAR(20) | 防守方玩家 tag |
| `attacker_name` | VARCHAR(100) | 攻击方名 |
| `defender_name` | VARCHAR(100) | 防守方名 |
| `stars` | INT | 星数 0-3 |
| `destruction_percent` | DECIMAL(5,2) | 摧毁率 |
| `order` | INT | 攻击序号 |
| `duration` | INT | 攻击时长(秒) |
| `side` | VARCHAR(10) | clan / opponent |
| `fetched_at` | DATETIME | 获取时间 |

索引:`INDEX(attacker_tag)`、`INDEX(defender_tag)`、`INDEX(war_snapshot_id)`。

现有 `war_members` 只存聚合(`attacks_used` / `best_stars_against`),无法支撑玩家档案。新增此表后,`GET /players/:tag/war-record` 按 `attacker_tag` / `defender_tag` 聚合。

### 3.2 复用现有表

- `war_snapshots`:CWL 每轮对战落库时复用(已有 war_state / team_size / 双方星数 / 摧毁率字段)。
- `war_members`:CWL 每轮对战成员复用。
- `war_targets`:CWL 敌方目标可绑定找阵任务(已建表,Phase 2 预留)。

### 3.3 不入库的模块

- **排名榜**:官方数据快照,Redis 缓存 10min,不入库。
- **联赛元数据**:极少变动,Redis 缓存 30min,不入库。
- **标签**:极少变动,Redis 缓存 1h,不入库。
- **部族 / 玩家概览**:官方 API 数据,Redis 缓存 5min,不入库(非 WarSpark 资产)。

## 4. 接口设计

### 4.1 部族概览(ClanService)— 2 端点

| # | 方法 | 路径 | cocapi 端点 | 缓存 | 落库 |
| --- | --- | --- | --- | --- | --- |
| 1 | GET | /api/v1/clans/:tag | GetClan + GetClanMembers | 5min | 否 |

**响应**:部族基本信息(等级 / 描述 / wars 赢平负 / 成员数 / 徽章 / 标签)+ 成员列表(tag / 名 / TH / 职位 / 捐兵 / 对战星数)。

`GetClan` 和 `GetClanMembers` 聚合到一次响应(成员列表可能较长,但与现有 `/layouts/:id` 聚合风格一致)。

### 4.2 玩家概览 + 战争档案(PlayerService)— 2 端点 + 1 聚合

| # | 方法 | 路径 | cocapi 端点 | 缓存 | 落库 |
| --- | --- | --- | --- | --- | --- |
| 2 | GET | /api/v1/players/:tag | GetPlayer | 5min | 否 |
| 3 | GET | /api/v1/players/:tag/battle-log | GetBattleLog | 5min | 否 |
| 4 | GET | /api/v1/players/:tag/war-record | (DB 聚合) | 无 | 查 |

**`GET /players/:tag`**:基础信息(TH / 大本营等级 / 英雄等级 / 部族 / 联赛)+ 成就进度 + 曾属部族历史。全部在 `GetPlayer` 响应里,1 次调用。

**`GET /players/:tag/battle-log`**:`GetBattleLog` 近期对战记录(含军队组成),数据量大,单独接口。只缓存不入库。

**`GET /players/:tag/war-record`**:从 `war_attacks` 表按 `attacker_tag` / `defender_tag` 聚合该玩家所有战争攻击 / 防守记录。数据靠战争落库积累,实时查库。返回:攻击列表(星数 / 摧毁率 / 对手 / 战争信息)+ 防守列表(被打星数 / 摧毁率 / 攻击方 / 战争信息)+ 聚合统计(总攻击数 / 平均星数 / 平均摧毁率 / 总防守数 / 被打平均星数)。

### 4.3 CWL 情报增强(WarService 扩展)— 1 端点增强 + 1 新端点 + 落库

| # | 方法 | 路径 | cocapi 端点 | 缓存 | 落库 |
| --- | --- | --- | --- | --- | --- |
| 5 | GET | /api/v1/war/cwl | GetClanWarLeagueGroup + GetClanWarLeagueWar×7 | 5min | 是(upsert) |
| 6 | GET | /api/v1/war/cwl/history | (DB 查询) | 无 | 查 |

**`GET /war/cwl` 增强**:

1. 调 `GetClanWarLeagueGroup` 获取组结构(season / state / clans / 7 轮 warTags)。
2. 并行调 7 次 `GetClanWarLeagueWar`(每轮 warTag),聚合每轮对战结果。
3. 计算组排名(按 warWins 排序)、玩家排名(按总星数 / 摧毁率排序)。
4. 奖牌预估(按竞品公开公式,纯计算)。
5. 落库:upsert `cwl_groups` + `cwl_rounds`;每轮 war 存 `war_snapshots` + `war_members` + `war_attacks`;`state=ended` 后冻结。

**`GET /war/cwl/history`**:查 `cwl_groups` 表,返回历史赛季列表(season / state / group_size / 我方排名 / fetched_at)。

### 4.4 排名榜(RankingService)— 7 端点,5 个接口

| # | 方法 | 路径 | cocapi 端点 | 缓存 |
| --- | --- | --- | --- | --- |
| 7 | GET | /api/v1/locations | GetLocations | 10min |
| 8 | GET | /api/v1/locations/:id/rankings/clans | GetClanRanking | 10min |
| 9 | GET | /api/v1/locations/:id/rankings/players | GetPlayerRanking | 10min |
| 10 | GET | /api/v1/locations/:id/rankings/clans-capital | GetClanCapitalRanking | 10min |
| 11 | GET | /api/v1/locations/:id/rankings/clans-builder-base | GetClanBuilderBaseRanking | 10min |
| 12 | GET | /api/v1/locations/:id/rankings/players-builder-base | GetPlayerBuilderBaseRanking | 10min |

- `GetLocation` 单个地区详情合并到 `/locations` 列表(列表已含详情字段,不必单独接口)。
- 全球排名用 `locationId=global`(官方 API 约定),地区用具体 ID。
- Redis 缓存 10min(排名变动慢,竞品也是周期性更新)。
- 不入库(排名是官方数据快照,非 WarSpark 资产)。

**数据流**:`GET /locations` 首页展示地区列表 → 选地区 → 4 种排名榜 tab 切换。

### 4.5 联赛元数据(LeagueService)— 12 端点,8 个接口

| # | 方法 | 路径 | cocapi 端点 | 缓存 |
| --- | --- | --- | --- | --- |
| 13 | GET | /api/v1/leagues | GetLeagues | 30min |
| 14 | GET | /api/v1/leagues/:id | GetLeague | 30min |
| 15 | GET | /api/v1/leagues/:id/seasons | GetLeagueSeasons | 30min |
| 16 | GET | /api/v1/leagues/:id/seasons/:season/rankings | GetLeagueSeasonRankings | 30min |
| 17 | GET | /api/v1/leagues/:id/seasons/:season/tiers | GetLeagueTiers | 30min |
| 18 | GET | /api/v1/leagues/:id/seasons/:season/tiers/:tier | GetLeagueTier | 30min |
| 19 | GET | /api/v1/leagues/:id/seasons/:season/tiers/:tier/history | GetLeagueHistory | 30min |
| 20 | GET | /api/v1/war-leagues | GetWarLeagues | 30min |
| 21 | GET | /api/v1/war-leagues/:id | GetWarLeague | 30min |

- `GetBuilderBaseLeagues` / `GetBuilderBaseLeague` 合并到 `/leagues` 列表(按 leagueType 区分 home / builder / capital / war,统一查询)。
- `GetLeagueGroup` 赛季分组信息合并到 `/leagues/:id/seasons/:season` 响应。
- Redis 缓存 30min(联赛元数据极少变动)。
- 不入库。

### 4.6 标签(LabelService)— 2 端点,2 个接口

| # | 方法 | 路径 | cocapi 端点 | 缓存 |
| --- | --- | --- | --- | --- |
| 22 | GET | /api/v1/clans/labels | GetClanLabels | 1h |
| 23 | GET | /api/v1/players/labels | GetPlayerLabels | 1h |

- Redis 缓存 1h(标签极少变动)。
- 不入库。
- 用途:① 部族 / 玩家详情页展示标签徽章 ② 前端筛选器选项来源(如部族列表按标签筛选时,标签选项从这里来)。

### 4.7 接口清单汇总

本次新增 / 增强 **23 个 cocapi 端点**,对应 **23 个 HTTP 接口**(21 个新接口 + 1 个已有接口增强 + 1 个 DB 聚合接口)。端点使用 7 → 23 / 35(66%)。

**按 service 分**:

| Service | 接口数 | 端点数 | 落库 |
| --- | --- | --- | --- |
| ClanService | 1 | 2 | 否 |
| PlayerService | 3 | 2 + 1 聚合 | 否(档案查库) |
| WarService(扩展) | 2(1 增强 + 1 新) | 1 + 7 | 是 |
| RankingService | 6 | 7 | 否 |
| LeagueService | 9 | 12 | 否 |
| LabelService | 2 | 2 | 否 |
| **合计** | **23** | **23** | — |

注:接口数 23 = 21 个新 HTTP 接口 + 1 个已有接口增强(`GET /war/cwl`)+ 1 个 DB 聚合接口(`GET /players/:tag/war-record`)。其中 2 个接口(`war-record` 和 `cwl/history`)不调 cocapi 端点,纯查库,故 23 接口对应 23 cocapi 端点(含 `GetClanWarLeagueWar`×7 计 7 次)。

## 5. 缓存策略

| 模块 | TTL | 理由 |
| --- | --- | --- |
| 部族概览 | 5min | 成员列表变动较频,但 5min 足够 |
| 玩家概览 | 5min | 与部族一致 |
| 玩家对战日志 | 5min | 数据量大,缓存减轻 API 负担 |
| 玩家战争档案 | 无 | 实时查库,数据靠落库积累 |
| CWL 组 | 5min | 进行中时每轮结果会变,5min 平衡时效与 API 配额 |
| CWL 历史 | 无 | 查库 |
| 排名榜 | 10min | 排名变动慢,竞品也是周期性更新 |
| 联赛元数据 | 30min | 极少变动 |
| 标签 | 1h | 极少变动 |

缓存 key 命名沿用现有 `war:current:` / `war:cwl:` 模式,扩展为 `clan:` / `player:` / `ranking:` / `league:` / `label:` 前缀。

## 6. 错误处理

沿用现有 `wardomain.Error` + `StatusForCode` 模式。新增错误码:

| 错误码 | HTTP | 含义 |
| --- | --- | --- |
| `invalid_tag` | 400 | tag 格式错误(已有) |
| `api_not_configured` | 500 | CoC API token 未配置(已有) |
| `api_request_failed` | 502 | 官方 API 请求失败(已有) |
| `api_access_denied` | 403 | 官方 API 403(已有) |
| `war_not_found` | 404 | 无当前战争(已有) |
| `clan_not_found` | 404 | 部族不存在(新增) |
| `player_not_found` | 404 | 玩家不存在(新增) |
| `league_not_found` | 404 | 联赛不存在(新增) |
| `location_not_found` | 404 | 地区不存在(新增) |

adapter 的 `mapError` 函数扩展,将 cocapi 的 `ErrNotFound` 按上下文映射为对应的领域错误码。

## 7. 测试策略

沿用现有测试模式:

- **adapter 测试**(`internal/infra/coc/client_test.go`):类型转换、错误映射、tag 编码。新增 21 个转换方法的测试。
- **service 测试**(`internal/service/*_test.go`):用 fake repository / fake adapter,验证业务逻辑(CWL 聚合 / 奖牌计算 / 战争档案聚合)。各 service 独立测试。
- **controller 测试**(`internal/controller/*_test.go`):绑定、响应格式、错误码 HTTP 映射。
- **repository 测试**:CWL / war_attacks 落库 SQL 正确性(属 P0-3 待办,本设计暂不要求,但接口预留可测性)。

## 8. 实施顺序

建议按依赖关系分批:

1. **数据模型 + adapter 扩展**(地基):3 张表迁移 + `infracoc.Client` 21 个转换方法 + 测试。
2. **ClanService + PlayerService**(独立模块):不依赖落库,见效快。
3. **RankingService + LeagueService + LabelService**(独立模块):纯缓存,不依赖落库,模式统一。
4. **WarService CWL 增强**(依赖地基):落库 + 聚合 + 奖牌计算,最复杂,放最后。

每批完成后 `make check` 验证。

## 9. 端点覆盖统计

### 9.1 本次使用(23 / 35)

| 域 | 端点 | 数量 |
| --- | --- | --- |
| 部族核心 | GetClan / GetClanMembers / GetClanLabels | 3 |
| 战争 | GetCurrentWar / GetClanWarLeagueGroup / GetClanWarLeagueWar | 3 |
| 玩家 | GetPlayer / GetBattleLog / GetPlayerLabels | 3 |
| 排名 / 地点 | GetLocations / GetLocation / GetClanCapitalRanking / GetClanRanking / GetClanBuilderBaseRanking / GetPlayerRanking / GetPlayerBuilderBaseRanking | 7 |
| 联赛元数据 | GetLeagues / GetLeague / GetLeagueSeasons / GetLeagueSeasonRankings / GetLeagueTiers / GetLeagueTier / GetLeagueGroup / GetLeagueHistory / GetWarLeagues / GetWarLeague / GetBuilderBaseLeagues / GetBuilderBaseLeague | 12(其中 GetLocation 合并到列表、GetLeagueGroup 合并到赛季) |

### 9.2 剩余未用(12 / 35,留给后续子项目)

| 域 | 端点 | 后续用途 |
| --- | --- | --- |
| 都城 | GetCapitalRaidSeasons / GetCapitalLeagues / GetCapitalLeague | 都城突袭模块 |
| Gold Pass | GetCurrentGoldPassSeason | 日历 / 活动模块 |
| 搜索 | SearchClans | 部族搜索(按名) |
| 战争日志 | GetClanWarLog | 历史战争浏览(需配合自存快照) |
| 账号 | VerifyToken | 账号系统(MVP 不做) |

## 10. 验收标准

- 23 个端点全部接入,实时 API 调用验证响应正确。
- 3 张新表迁移落库,CWL 落库逻辑正确(upsert + ended 冻结)。
- 玩家战争档案能从 `war_attacks` 聚合出攻击 / 防守记录 + 统计。
- 缓存 TTL 按设计生效,未命中才请求官方 API。
- 错误码 HTTP 映射正确(404 / 403 / 502)。
- `make check` 全绿(build + fmt + vet + test)。
- adapter / service / controller 层测试覆盖。
