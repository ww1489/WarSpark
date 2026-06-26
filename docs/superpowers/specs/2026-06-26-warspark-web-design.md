# WarSpark Web — 前端设计文档

## 概述

WarSpark Web 是 WarSpark 后端的前台界面，面向 Clash of Clans 公会管理层和普通玩家，
提供部落/玩家数据查询、排名浏览、战争监控等功能。

## 技术选型

| 维度 | 选择 | 理由 |
|------|------|------|
| 框架 | React 18 + TypeScript | 生态丰富，适合中大型项目 |
| 元框架 | TanStack Start | 文件路由 + SSR 支持 |
| 样式 | Tailwind CSS 4 | 原子化 CSS，开发效率高 |
| 数据请求 | TanStack Query | 缓存、重试、自动刷新 |
| 路由 | TanStack Router | 类型安全路由 |
| 构建 | Vite | 快 |
| 包管理 | pnpm | 快，磁盘友好 |
| 存储 | 独立 Git 仓库 | 与后端解耦，独立部署 |
| Lint | Biome | 替代 ESLint + Prettier，快速 |

## 项目结构

```
warspark-web/
├── app/
│   ├── __root.tsx              # 根布局：Navbar + 内容区 + Footer
│   ├── index.tsx               # 首页
│   ├── clans.$tag.tsx          # 部落详情（含 loader）
│   ├── players.$tag.tsx        # 玩家详情（含 loader）
│   ├── leaderboards.tsx        # 排行榜
│   ├── search.tsx              # 部落高级搜索
│   ├── gold-pass.tsx           # Gold Pass 赛季
│   └── admin/
│       ├── __layout.tsx        # Admin 布局壳
│       ├── war.tsx             # 当前战争监控
│       └── cwl.tsx             # CWL 联赛一览
├── lib/
│   ├── api.ts                  # fetch 封装 + 所有 API 函数
│   ├── types.ts                # 全部 TypeScript 类型定义
│   ├── queries.ts              # TanStack Query hooks
│   └── utils.ts                # 格式化工具（tag、数字、时间）
├── components/
│   └── ui/                     # 原子组件
│       ├── Button.tsx
│       ├── Card.tsx
│       ├── Input.tsx
│       ├── Select.tsx
│       ├── Table.tsx
│       ├── Badge.tsx
│       ├── Tabs.tsx
│       ├── Skeleton.tsx
│       ├── ErrorCard.tsx
│       ├── EmptyState.tsx
│       └── SearchBar.tsx
├── styles/
│   └── globals.css             # Tailwind + CSS 变量
├── public/
│   └── favicon.ico
├── package.json
├── tsconfig.json
├── tailwind.config.ts
├── vite.config.ts
└── biome.json
```

---

## 导航结构

### 桌面端 (≥1024px)

```
┌─────────────────────────────────────────────────────┐
│  🔥 WarSpark   首页 排行榜 部落搜索 GoldPass   [管理] │
│                                          🔍 搜索框  │
└─────────────────────────────────────────────────────┘
```

- Logo 左侧，点击回首页
- 导航项：首页 / 排行榜 / 部落搜索 / Gold Pass
- 右侧搜索框：输入部落标签或玩家标签 → 快捷跳转
- 「管理」链接仅在有 token 时显示（当前阶段隐藏不显示）

### 移动端 (<1024px)

Hamburger 菜单替换水平导航，搜索框并入菜单。

---

## 路由与页面详细说明

### 1. 首页 `/`

**目的**：快速搜索入口 + 全球热门数据一览

**布局**：

```
┌──────────────────────────────────────────┐
│  HeroSection                              │
│  ┌────────────────────────────────────┐  │
│  │   🔍 输入部落标签或玩家标签...      │  │
│  │   [搜索]  #2PP  #ABC123            │  │
│  └────────────────────────────────────┘  │
├──────────────────────────────────────────┤
│  📊 热门排行榜                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ TOP 部落  │ │ TOP 玩家  │ │ TOP 都城  │ │
│  │ 1. Clan A │ │ 1. Player │ │ 1. Clan X │ │
│  │ 2. Clan B │ │ 2. Player │ │ 2. Clan Y │ │
│  │ 3. Clan C │ │ 3. Player │ │ 3. Clan Z │ │
│  │  查看更多→ │ │  查看更多→ │ │  查看更多→ │ │
│  └──────────┘ └──────────┘ └──────────┘ │
└──────────────────────────────────────────┘
```

**数据获取**：
- `useQuery({ queryKey: ['locations'], queryFn: getLocations })` — 获取位置列表
- 选中默认位置（global `32000006`）
- `useQuery({ queryKey: ['rankings', locId, 'clans'], queryFn: () => getClanRanking(locId) })` — TOP 5 部落
- `useQuery({ queryKey: ['rankings', locId, 'players'], queryFn: () => getPlayerRanking(locId) })` — TOP 5 玩家

**交互**：
- 搜索框输入部落标签（`#` 开头）→ 跳转 `/clans/:tag`
- 搜索框输入玩家标签 → 跳转 `/players/:tag`
- 点击「查看更多」→ 跳转 `/leaderboards`
- 每个卡片展示前 5 条，带徽章编号

**状态处理**：
- `isLoading` → 3 个 `Skeleton` 卡片
- `isError` → `ErrorCard` + 重试按钮
- 数据为空 → `EmptyState("暂无数据")`

---

### 2. 部落详情 `/clans/:tag`

**目的**：部落完整信息 + 战争历史 + 突袭赛季

**布局**（Tab 切换）：

```
┌──────────────────────────────────────────┐
│  ClanHeader                               │
│  [徽章] 部落名称  #TAG  Lv.15            │
│  📍 位置 · ⚔️ 战争频率 · 📊 部落积分      │
│  成员: 45/50  胜:120 负:20 平:5          │
├──────────────────────────────────────────┤
│  [概览]  [战争日志]  [突袭赛季]           │
├──────────────────────────────────────────┤
│  Tab 内容区                               │
└──────────────────────────────────────────┘
```

**Tab 0 — 概览**：

成员表格：
| # | 名称 | 角色 | 等级 | 大本营 | 奖杯 | 捐兵 |
|---|------|------|------|--------|------|------|
| 1 | 玩家A | 首领 | 200 | TH16 | 5500 | 1200 |
| 2 | 玩家B | 副首领 | 180 | TH15 | 5000 | 800  |

**Tab 1 — 战争日志**：

战争历史表格：
| 日期 | 对手 | 规模 | 结果 | 星数 | 摧毁% |
|------|------|------|------|------|-------|
| 06-25 | 部落X | 15v15 | 🟢 胜利 | 42-38 | 95%-82% |
| 06-23 | 部落Y | 20v20 | 🔴 失败 | 35-40 | 80%-92% |

**Tab 2 — 突袭赛季**：

赛季卡片列表，每个展示：
- 赛季起止时间
- 突袭完成数 / 总攻击次数
- 首都总战利品 / 进攻/防御奖励

**数据获取**：
- `useQuery({ queryKey: ['clan', tag], queryFn: () => getClan(tag) })`
- `useQuery({ queryKey: ['warLog', tag], queryFn: () => getWarLog(tag) })`
- `useQuery({ queryKey: ['raidSeasons', tag], queryFn: () => getCapitalRaidSeasons(tag) })`

**组件 Props**：

```ts
// ClanHeader
interface ClanHeaderProps {
  clan: ClanOverview
}

// MembersTable
interface MembersTableProps {
  members: ClanMemberSummary[]
}

// WarLogTable
interface WarLogTableProps {
  items: WarLogEntry[]
  isLoading: boolean
}

// CapitalRaidCard
interface CapitalRaidCardProps {
  season: CapitalRaidSeason
}
```

---

### 3. 玩家详情 `/players/:tag`

**布局**：

```
┌──────────────────────────────────────────┐
│  PlayerHeader                             │
│  [等级图标] 玩家名  #TAG                 │
│  🏆 奖杯 5000 · ⭐ 战争之星 800          │
│  🏰 TH16 · 🔧 BH10                       │
│  👥 部落: 部落A (#CLAN)                  │
├──────────────────────────────────────────┤
│  英雄                                      │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐    │
│  │ 👑   │ │ 👸   │ │ 🛡️   │ │ 🏇   │    │
│  │ BK80 │ │ AQ85 │ │ GW60 │ │ RC30 │    │
│  └──────┘ └──────┘ └──────┘ └──────┘    │
├──────────────────────────────────────────┤
│  战斗日志                                  │
│  ┌──────────────────────────────────────┐ │
│  │ 06-25 ⚔️ 进攻 · ★★★ 100% · 对手X    │ │
│  │ 06-25 🛡️ 防守 · ★★☆ 67% · 来自Y    │ │
│  │ 06-24 ⚔️ 进攻 · ★★★ 100% · 对手Z    │ │
│  └──────────────────────────────────────┘ │
├──────────────────────────────────────────┤
│  联赛分组                                  │
│  ┌──────────────────────────────────────┐ │
│  │ 🏆 Crystal III · 排名: 3/15          │ │
│  │ 进攻: 3胜1负 · 防守: 2胜2负          │ │
│  └──────────────────────────────────────┘ │
└──────────────────────────────────────────┘
```

**数据获取**：
- `useQuery({ queryKey: ['player', tag], queryFn: () => getPlayer(tag) })`
- `useQuery({ queryKey: ['battleLog', tag], queryFn: () => getBattleLog(tag) })`
- `useQuery({ queryKey: ['leagueGroup', tag], queryFn: () => getLeagueGroup(tag) })`

**组件 Props**：

```ts
interface PlayerHeaderProps {
  player: PlayerOverview
}

interface HeroCardProps {
  name: string
  level: number
  maxLevel: number
  village: 'home' | 'builderBase'
}

interface BattleLogEntryProps {
  entry: BattleLogEntry  // 进攻/防守类型 + 星数 + 摧毁%
}

interface LeagueGroupCardProps {
  members: LeagueGroupMember[]
  attackLogs: LeagueBattleLogEntry[]
  defenseLogs: LeagueBattleLogEntry[]
}
```

---

### 4. 排行榜 `/leaderboards`

**布局**：

```
┌──────────────────────────────────────────┐
│  位置: [ ▼ Global (32000006) ]           │
│                                          │
│  [部落] [玩家] [都城] [夜世界部落] [夜世界玩家] │
├──────────────────────────────────────────┤
│  # │ 部落名     │ 等级 │ 积分  │ 成员   │
│  1 │ Clan A    │  25  │ 58000 │ 50/50  │
│  2 │ Clan B    │  23  │ 57000 │ 48/50  │
│  3 │ Clan C    │  22  │ 56000 │ 50/50  │
│  ...                                      │
├──────────────────────────────────────────┤
│           [加载更多]                       │
└──────────────────────────────────────────┘
```

**交互流**：
1. 进入页面 → `useQuery(['locations'])` 加载位置列表
2. 默认选中 Global → `useQuery(['rankings', locId, type])` 加载对应排名
3. 用户切换排名类型 → URL search params 更新 `?type=clans`
4. 用户更换位置 → 重置排名数据，重新加载

**URL 设计**：`/leaderboards?location=32000006&type=clans`

**数据获取**：
```ts
const { location } = useSearch({ from: '/leaderboards' })
const locId = location?.location ?? '32000006'
const type = location?.type ?? 'clans'

useQuery({
  queryKey: ['locations'],
  queryFn: getLocations,
})

useQuery({
  queryKey: ['rankings', locId, type],
  queryFn: () => {
    switch (type) {
      case 'clans': return getClanRanking(locId)
      case 'players': return getPlayerRanking(locId)
      case 'capital': return getClanCapitalRanking(locId)
      case 'builder-clans': return getClanBuilderBaseRanking(locId)
      case 'builder-players': return getPlayerBuilderBaseRanking(locId)
    }
  },
})
```

---

### 5. 部落搜索 `/search`

**布局**：

```
┌──────────────────────────────────────────┐
│  搜索条件                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ 名称     │ │ 战争频率  │ │ 位置     │ │
│  │ [____]  │ │ [▼ 全部] │ │ [▼ 全部] │ │
│  └──────────┘ └──────────┘ └──────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ 最少成员  │ │ 最少等级  │ │ 最少积分  │ │
│  │ [  0  ]  │ │ [  0  ]  │ │ [  0  ]  │ │
│  └──────────┘ └──────────┘ └──────────┘ │
│  [搜索] [重置]                             │
├──────────────────────────────────────────┤
│  搜索结果 (N 条)                           │
│  ┌──────────────────────────────────────┐ │
│  │ [徽章] 部落名 #TAG Lv.10             │ │
│  │ 成员 30/50 · 🏆 25000 · ⚔️ 一直打仗  │ │
│  │ [查看详情]                           │ │
│  └──────────────────────────────────────┘ │
│  ┌──────────────────────────────────────┐ │
│  │ ...更多结果...                       │ │
│  └──────────────────────────────────────┘ │
└──────────────────────────────────────────┘
```

**URL 设计**：所有搜索参数同步到 URL search params，支持分享。

```
/search?name=war&warFrequency=always&locationId=32000006&minMembers=30
```

**数据获取**：
```ts
const searchParams = useSearch({ from: '/search' })

useQuery({
  queryKey: ['searchClans', searchParams],
  queryFn: () => searchClans(searchParams),
  enabled: !!searchParams.name, // 至少需要名称
})
```

---

### 6. Gold Pass `/gold-pass`

**布局**：

```
┌──────────────────────────────────────────┐
│  🏆 Gold Pass 当前赛季                    │
│  ┌──────────────────────────────────────┐ │
│  │  2026-06-01 ～ 2026-07-01           │ │
│  │  ⏰ 剩余 XX 天                       │ │
│  └──────────────────────────────────────┘ │
└──────────────────────────────────────────┘
```

- 简单信息展示页
- `useQuery({ queryKey: ['goldPass'], queryFn: getCurrentGoldPass })`
- 赛季起止日期 + 剩余天数计算

---

### 7. Admin — 战争监控 `/admin/war`

**目的**：输入多个部落标签，实时查看当前战争状态

**布局**：

```
┌──────────────────────────────────────────┐
│  战争监控                                  │
│  ┌──────────────────────────────────────┐ │
│  │ 输入部落标签: #AAA111 #BBB222 [#CCC] │ │
│  │ [+ 添加]                             │ │
│  └──────────────────────────────────────┘ │
├──────────────────────────────────────────┤
│  ┌────────────────┐ ┌────────────────┐   │
│  │ #AAA111        │ │ #BBB222        │   │
│  │ 状态: 战争中    │ │ 状态: 未参战    │   │
│  │ 15v15          │ │ —              │   │
│  │ ⭐ 42 vs 38    │ │                │   │
│  │ 💥 95% vs 82%  │ │                │   │
│  │ ⏰ 剩余 2h      │ │                │   │
│  └────────────────┘ └────────────────┘   │
└──────────────────────────────────────────┘
```

**交互**：
- 用 localStorage 持久化部落标签列表
- 每个标签独立查询：`useQuery({ queryKey: ['currentWar', tag], queryFn: () => getCurrentWar(tag) })`
- 30秒自动刷新：`refetchInterval: 30000`
- 战争状态：`notInWar` / `preparation` / `inWar` / `warEnded`

---

### 8. Admin — CWL 联赛 `/admin/cwl`

**布局**：

```
┌──────────────────────────────────────────┐
│  CWL 联赛                                 │
│  部落标签: [#AAA111]                      │
├──────────────────────────────────────────┤
│  联赛信息                                  │
│  赛季: 2026-06 · 状态: 进行中              │
├──────────────────────────────────────────┤
│  参赛部落                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Clan A   │ │ Clan B   │ │ Clan C   │  │
│  │ Lv.15    │ │ Lv.14    │ │ Lv.13    │  │
│  │ ⭐ 120   │ │ ⭐ 110   │ │ ⭐ 105   │  │
│  └──────────┘ └──────────┘ └──────────┘  │
├──────────────────────────────────────────┤
│  轮次                                      │
│  ┌──────────────────────────────────────┐ │
│  │ 第1轮 · #WAR1 → [查看详情]           │ │
│  │ 第2轮 · #WAR2 → [查看详情]           │ │
│  │ ...                                  │ │
│  └──────────────────────────────────────┘ │
└──────────────────────────────────────────┘
```

**数据获取**：
- `useQuery({ queryKey: ['cwlGroup', tag], queryFn: () => getCWLGroup(tag) })`

---

## API 层设计

### `lib/types.ts`

```ts
// 统一 API 响应格式
interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

// 分页游标
interface PagingCursors {
  after?: string
  before?: string
}
interface Paging {
  cursors: PagingCursors
}

// === 部落 ===
interface ClanOverview {
  tag: string
  name: string
  clanLevel: number
  description: string
  members: number
  clanPoints: number
  warWins: number
  warLosses?: number
  warTies?: number
  warWinStreak: number
  warFrequency: string
  type: string
  isWarLogPublic: boolean
  location?: Location
  warLeague?: LeagueRef
  badgeUrls: { small: string; medium: string; large: string }
  labels: Label[]
}

interface ClanDetail {
  clan: ClanOverview
  members: ClanMemberSummary[]
}

interface ClanMemberSummary {
  tag: string
  name: string
  role: string
  expLevel: number
  townHallLevel: number
  trophies: number
  clanRank: number
  donations: number
  donationsReceived: number
  league?: LeagueRef
}

// === 玩家 ===
interface PlayerOverview {
  tag: string
  name: string
  townHallLevel: number
  expLevel: number
  role: string
  warStars: number
  attackWins: number
  defenseWins: number
  trophies: number
  bestTrophies: number
  clan?: PlayerClanInfo
  league?: LeagueRef
  labels: Label[]
  heroes: HeroLevel[]
  achievements: AchievementProgress[]
}

interface PlayerClanInfo {
  tag: string
  name: string
  clanLevel: number
  badgeUrls: { small: string; medium: string }
}

interface HeroLevel {
  name: string
  level: number
  maxLevel: number
  village: string
}

// === 战争 ===
interface WarLogEntry {
  result: string
  teamSize: number
  clan: WarLogClan
  opponent: WarLogClan
  endTime: string
}

interface WarLogClan {
  tag: string
  name: string
  clanLevel: number
  stars: number
  destructionPercentage: number
  attacks: number
}

// === 排行榜 ===
interface RankingEntry {
  tag: string
  name: string
  clanLevel?: number
  rank: number
  trophies?: number
  clanPoints?: number
  members?: number
}

// === 搜索 ===
interface ClanSearchParams {
  name?: string
  warFrequency?: string
  locationId?: number
  minMembers?: number
  maxMembers?: number
  minClanPoints?: number
  minClanLevel?: number
  limit?: number
  after?: string
  before?: string
}

// === Gold Pass ===
interface GoldPassSeason {
  startTime: string
  endTime: string
}

// === 通用 ===
interface Label {
  id: number
  name: string
  iconUrls: { small: string; medium: string }
}

interface Location {
  id: number
  name: string
  countryCode?: string
}

interface LeagueRef {
  id: number
  name: string
}
```

### `lib/api.ts`

```ts
const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1'

async function fetchApi<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) throw new ApiError(res.status, await res.text())
  const json: ApiResponse<T> = await res.json()
  if (json.code !== 0) throw new ApiError(json.code, json.message)
  return json.data
}

// 部落
export const getClan = (tag: string) =>
  fetchApi<ClanDetail>(`/clans/${encodeURIComponent(tag)}`)

// 玩家
export const getPlayer = (tag: string) =>
  fetchApi<PlayerOverview>(`/players/${encodeURIComponent(tag)}`)

// 战争日志
export const getWarLog = (tag: string, limit = 10, after?: string) =>
  fetchApi<PagedList<WarLogEntry>>(
    `/clans/${encodeURIComponent(tag)}/war-log?limit=${limit}${after ? `&after=${after}` : ''}`
  )

// 排名
export const getLocations = () =>
  fetchApi<{ items: Location[] }>('/locations')

export const getClanRanking = (locationId: string) =>
  fetchApi<PagedList<ClanRankingEntry>>(`/locations/${locationId}/rankings/clans`)

export const getPlayerRanking = (locationId: string) =>
  fetchApi<PagedList<PlayerRankingEntry>>(`/locations/${locationId}/rankings/players`)

// 搜索
export const searchClans = (params: ClanSearchParams) =>
  fetchApi<PagedList<ClanOverview>>(`/clans?${new URLSearchParams(
    Object.entries(params).filter(([, v]) => v != null).map(([k, v]) => [k, String(v)])
  )}`)

// Gold Pass
export const getCurrentGoldPass = () =>
  fetchApi<GoldPassSeason>('/gold-pass/current')

// 当前战争
export const getCurrentWar = (tag: string) =>
  fetchApi<CurrentWar>(`/war/current?clan_tag=${encodeURIComponent(tag)}`)

// CWL
export const getCWLGroup = (tag: string) =>
  fetchApi<CWLGroup>(`/war/cwl?clan_tag=${encodeURIComponent(tag)}`)
```

### `lib/queries.ts`

```ts
// TanStack Query hooks — 集中导出，每个页面文件直接引用

export function useClan(tag: string) {
  return useQuery({
    queryKey: ['clan', tag],
    queryFn: () => getClan(tag),
    staleTime: 5 * 60 * 1000,   // 5min 缓存
    enabled: !!tag,
  })
}

export function usePlayer(tag: string) {
  return useQuery({
    queryKey: ['player', tag],
    queryFn: () => getPlayer(tag),
    staleTime: 5 * 60 * 1000,
    enabled: !!tag,
  })
}

export function useWarLog(tag: string) {
  return useQuery({
    queryKey: ['warLog', tag],
    queryFn: () => getWarLog(tag),
    staleTime: 5 * 60 * 1000,
    enabled: !!tag,
  })
}

export function useRankings(locationId: string, type: RankingType) {
  return useQuery({
    queryKey: ['rankings', locationId, type],
    queryFn: () => {
      switch (type) {
        case 'clans': return getClanRanking(locationId)
        case 'players': return getPlayerRanking(locationId)
        case 'capital': return getClanCapitalRanking(locationId)
        case 'builder-clans': return getClanBuilderBaseRanking(locationId)
        case 'builder-players': return getPlayerBuilderBaseRanking(locationId)
      }
    },
    staleTime: 10 * 60 * 1000,
    enabled: !!locationId,
  })
}

export function useSearchClans(params: ClanSearchParams) {
  return useQuery({
    queryKey: ['searchClans', params],
    queryFn: () => searchClans(params),
    enabled: !!params.name || !!params.locationId,
  })
}

export function useCurrentWar(tag: string) {
  return useQuery({
    queryKey: ['currentWar', tag],
    queryFn: () => getCurrentWar(tag),
    refetchInterval: 30_000,  // 战中30秒刷新
    enabled: !!tag,
  })
}
```

---

## 通用组件规格

### Button

```tsx
<Button variant="primary" size="md" onClick={...}>
  搜索
</Button>
<Button variant="secondary">取消</Button>
<Button variant="ghost" size="sm">查看更多</Button>
<Button loading disabled>加载中...</Button>

// variants: primary | secondary | ghost | danger
// sizes: sm (h-8) | md (h-10) | lg (h-12)
```

### Card

```tsx
<Card>
  <Card.Header>标题</Card.Header>
  <Card.Body>内容</Card.Body>
  <Card.Footer>底部操作</Card.Footer>
</Card>

// 带 hover 效果：hover:shadow-md transition-shadow
```

### Table

```tsx
<Table
  columns={[
    { key: 'rank', header: '排名', width: '80px' },
    { key: 'name', header: '名称' },
    { key: 'trophies', header: '奖杯', align: 'right' },
  ]}
  data={items}
  loading={isLoading}
  emptyText="暂无数据"
  onRowClick={(row) => navigate(`/clans/${row.tag}`)}
/>
```

### Skeleton

```tsx
// 文本骨架
<Skeleton.Text lines={3} />

// 卡片骨架
<Skeleton.Card />

// 表格骨架
<Skeleton rows={5} cols={4} />
```

### ErrorCard

```tsx
<ErrorCard
  title="加载失败"
  message="无法获取部落数据，请稍后重试"
  onRetry={() => refetch()}
/>
```

### SearchBar（全局导航栏内）

```tsx
<SearchBar
  placeholder="输入部落标签或玩家标签"
  onSearch={(value) => {
    const tag = normalizeTag(value)
    if (tag.startsWith('#')) {
      // 判断是部落还是玩家：部落标签通常由字母+数字组成
      navigate(tag.length <= 12 ? `/players/${tag}` : `/search?tag=${tag}`)
    }
  }}
/>
```

---

## 工具函数 `lib/utils.ts`

```ts
// 标签归一化：#2pp → #2PP
export function normalizeTag(tag: string): string {
  return tag.trim().toUpperCase().replace(/^(?!\#)/, '#')
}

// 数字格式化：58000 → 58,000
export function formatNumber(n: number): string {
  return n.toLocaleString()
}

// 时间格式化：2026-06-25T12:00:00Z → 06-25 12:00
export function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${d.getMonth()+1}-${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2,'0')}`
}

// 战争结果映射
export function warResultLabel(result: string): { text: string; color: string } {
  switch (result) {
    case 'win': return { text: '胜利', color: 'text-green-600' }
    case 'lose': return { text: '失败', color: 'text-red-600' }
    case 'tie': return { text: '平局', color: 'text-yellow-600' }
    default: return { text: result, color: 'text-gray-500' }
  }
}

// 角色中文映射
export function roleLabel(role: string): string {
  const map: Record<string, string> = {
    leader: '首领', coLeader: '副首领', admin: '长老', member: '成员'
  }
  return map[role] ?? role
}

// 大本营图标获取
export function townHallIcon(level: number): string {
  return `/th-icons/th${level}.png`
}
```

---

## 响应式设计

| 断点 | 宽度 | 布局变化 |
|------|------|---------|
| Desktop | ≥1024px | 完整布局，导航水平 |
| Tablet | 768–1023px | 导航折叠，卡片网格 2 列 |
| Mobile | <768px | 汉堡菜单，卡片/表格全宽，数字简化展示 |

- 表格在移动端改为垂直卡片形式（Card Stack）
- 搜索表单在移动端折叠为可展开区域

---

## 错误处理策略

| 层级 | 处理方式 |
|------|---------|
| `lib/api.ts` | 统一解构 `{ code, message, data }`，`code !== 0` 抛 `ApiError` |
| TanStack Query | `isError` / `error` 自动传递到组件 |
| 页面组件 | 根据 `isError` 渲染 `ErrorCard`，提供 `refetch()` 重试 |
| 全局 | `QueryClient` 配置 `retry: 1`（失败重试 1 次） |

---

## 环境变量

```bash
# .env
VITE_API_URL=http://localhost:8080/api/v1

# .env.production
VITE_API_URL=https://api.warspark.com/api/v1
```

---

## 依赖清单

```json
{
  "dependencies": {
    "@tanstack/react-query": "^5",
    "@tanstack/react-router": "^1",
    "@tanstack/react-start": "^1",
    "react": "^18",
    "react-dom": "^18"
  },
  "devDependencies": {
    "@types/react": "^18",
    "@types/react-dom": "^18",
    "@biomejs/biome": "^1",
    "tailwindcss": "^4",
    "typescript": "^5",
    "vite": "^6"
  }
}
```

## 非目标（YAGNI）

- 不做 PWA / Service Worker
- 不做 WebSocket 实时推送
- 不做国际化（仅中文）
- 不做暗色模式（后续再加）
- 不做 SEO（TanStack Start SSR 仅用于首屏加速）
