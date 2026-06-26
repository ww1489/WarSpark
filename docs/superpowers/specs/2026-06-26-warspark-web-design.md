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
| 构建 | Vite | 快 |
| 存储 | 独立 Git 仓库 | 与后端解耦，独立部署 |

## 项目结构

```
warspark-web/
├── app/
│   ├── __root.tsx             # 根布局：导航栏 + 内容区
│   ├── index.tsx              # 首页：搜索入口 + 热门排行榜
│   ├── clans.$tag.tsx         # 部落详情
│   ├── players.$tag.tsx       # 玩家详情
│   ├── leaderboards.tsx       # 排行榜
│   ├── search.tsx             # 部落搜索
│   ├── gold-pass.tsx          # Gold Pass 赛季
│   └── admin/
│       ├── __layout.tsx       # Admin 布局
│       ├── war.tsx            # 战争监控
│       └── cwl.tsx            # CWL 联赛
├── lib/
│   ├── api.ts                 # fetch 封装
│   └── types.ts               # API 响应类型
├── components/
│   └── ui/                    # 原子组件
├── styles/
│   └── globals.css            # Tailwind 入口
├── package.json
├── tsconfig.json
├── tailwind.config.ts
└── vite.config.ts
```

## 路由与页面

| 路由 | 页面 | 调用的 API |
|------|------|-----------|
| `/` | 首页 | GET /locations, GET /locations/:id/rankings/clans |
| `/clans/:tag` | 部落详情 | GET /clans/:tag, /war-log, /capital-raid-seasons |
| `/players/:tag` | 玩家详情 | GET /players/:tag, /battle-log, /league-group |
| `/leaderboards` | 排行榜 | GET /locations, /locations/:id/rankings/* |
| `/search` | 部落搜索 | GET /clans?name=... |
| `/gold-pass` | Gold Pass | GET /gold-pass/current |
| `/admin/war` | 战争监控 | GET /war/current |
| `/admin/cwl` | CWL 联赛 | GET /war/cwl, /war/cwl/wars/:tag |

## 组件树

```
__root.tsx
├── Navbar (SearchBar, 登录入口)
├── index.tsx
│   ├── HeroSection (快速搜索)
│   ├── TopClansPreview (TOP 部落卡片)
│   └── TopPlayersPreview (TOP 玩家卡片)
├── clans.$tag.tsx
│   ├── ClanHeader (部落名称、等级、标签、徽章)
│   ├── Tabs
│   │   ├── Overview (成员列表)
│   │   ├── WarLog (战争日志表格)
│   │   └── CapitalRaids (突袭赛季)
├── players.$tag.tsx
│   ├── PlayerHeader (名称、等级、奖杯)
│   ├── HeroGrid (英雄等级)
│   ├── BattleLogTable (战斗日志)
│   └── LeagueGroupCard (联赛分组)
├── leaderboards.tsx
│   ├── LocationPicker (下拉位置选择)
│   ├── RankingTypeTabs (clan/player/capital/builder)
│   └── RankingTable (排名表格)
├── search.tsx
│   ├── SearchForm (多条件筛选)
│   └── SearchResults (部落列表)
├── gold-pass.tsx
│   └── GoldPassCard (赛季信息)
├── admin/__layout.tsx
└── admin/
    ├── war.tsx (当前战争监控面板)
    └── cwl.tsx (CWL 联赛分组)
```

## 数据流

```
用户操作 → TanStack Router 路由 → 页面组件 → lib/api.ts (fetch)
  → WarSpark Backend (:8080) → JSON 响应 → 组件渲染
```

- 页面级数据用 TanStack Query (React Query) 管理缓存和刷新
- 搜索/查询参数用 URL search params 保持可分享
- 错误状态：`isLoading` / `isError` 自动处理，对应 Skeleton / ErrorCard

## API 类型定义

`lib/types.ts` 集中管理所有 API 响应类型：
- 从后端 Swagger JSON 手动迁移核心类型
- 统一响应格式 `{ code, message, data }`
- 每个端点定义请求参数和响应类型

## 认证

阶段暂时跳过认证，所有页面公开访问。Admin 路由的 JWT 校验由后端关闭（后续 C 阶段统一接入）。

## 状态管理

- 服务端数据：TanStack Query（缓存、重试、自动刷新）
- UI 状态：URL search params + React useState（无需全局状态库）

## 设计系统

参考 RunnerGo (runnergo.com) 的 SaaS 风格。

### 配色

| Token | 值 | 用途 |
|-------|-----|------|
| `--primary` | `#3A4DF0` | 主色调，品牌蓝 |
| `--primary-light` | `#8E9AFF` | 主色浅色变体 |
| `--accent-start` | `#3A4DF0` | 渐变起 |
| `--accent-end` | `#8E9AFF` | 渐变止（按钮/头部） |
| `--bg-page` | `#F7F9FD` | 页面背景（浅蓝灰） |
| `--bg-card` | `#FFFFFF` | 卡片/内容区背景 |
| `--text-primary` | `#1A1A2E` | 主文字 |
| `--text-secondary` | `#6B7280` | 次要文字 |
| `--border` | `#E5E7EB` | 边框/分割线 |
| `--shadow` | `#DCE8FF` | 卡片投影色 |

### 圆角

- 大按钮/Card：`12px`
- 小按钮/Badge：`8px`
- 输入框：`8px`

### 字体

- 标题：Inter / 系统默认 sans-serif，`font-semibold`
- 正文：Inter，`text-sm` / `text-base`
- 数据/数字：Tabular Nums（等宽数字，排名表格用）

### 间距

- 页面内容区最大宽度：`1280px`，居中
- 卡片间距：`gap-6`（24px）
- 区块间距：`py-12`（48px 上下）

### Tailwind 配置

```ts
// tailwind.config.ts
export default {
  theme: {
    extend: {
      colors: {
        primary: { DEFAULT: '#3A4DF0', light: '#8E9AFF', dark: '#2E3DC4' },
        accent: { DEFAULT: '#F05B3A', light: '#FDBD41' },
        surface: { page: '#F7F9FD', card: '#FFFFFF' },
      },
      borderRadius: { btn: '12px', card: '12px', tag: '8px' },
      boxShadow: { card: '0 2px 12px rgba(220, 232, 255, 0.5)' },
      maxWidth: { page: '1280px' },
    }
  }
}
```

## 非目标（YAGNI）

- 不做 PWA / Service Worker
- 不做 WebSocket 实时推送
- 不做国际化（仅中文）
- 不做暗色模式（后续再加）
