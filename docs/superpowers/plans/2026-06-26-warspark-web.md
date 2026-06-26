# WarSpark Web — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build WarSpark Web — a React/TypeScript frontend for the WarSpark Clash of Clans data API, with 8 pages covering clan/player details, leaderboards, search, and admin monitoring.

**Architecture:** TanStack Start (file-based routing, SSR) + TanStack Query (data fetching/caching) + Tailwind CSS 4 (styling). Components are page-level (route files) + app-level (UI library). API layer is a thin fetch wrapper in lib/api.ts with typed queries in lib/queries.ts.

**Tech Stack:** React 18, TypeScript 5, TanStack Start v1, TanStack Query v5, Tailwind CSS 4, Vite 6, Biome, pnpm

## Global Constraints

- React 18 + TypeScript 5 strict mode
- TanStack Start v1 file-based routing
- TanStack Query v5 for all server data
- Tailwind CSS 4 (no separate config file — use @theme in CSS)
- Biome for linting/formatting
- pnpm as package manager
- All API calls go through lib/api.ts (single fetch wrapper)
- All tanstack query hooks in lib/queries.ts
- 8 pages: home, clan detail, player detail, leaderboards, search, gold-pass, admin/war, admin/cwl
- Background: #F7F9FD, Cards: #FFFFFF, Primary: #3A4DF0
- No auth (all pages public), no PWA, no i18n, no dark mode
- Design system reference: RunnerGo SaaS style

---

## File Structure

```
warspark-web/
├── app/
│   ├── __root.tsx              # Root layout: Navbar + content area
│   ├── index.tsx               # Home page
│   ├── clans.$tag.tsx          # Clan detail
│   ├── players.$tag.tsx        # Player detail
│   ├── leaderboards.tsx        # Leaderboards
│   ├── search.tsx              # Clan search
│   ├── gold-pass.tsx           # Gold Pass
│   └── admin/
│       ├── __layout.tsx        # Admin layout shell
│       ├── war.tsx             # War monitoring
│       └── cwl.tsx             # CWL overview
├── lib/
│   ├── api.ts                  # fetch wrapper + all API functions
│   ├── types.ts                # all TypeScript type definitions
│   ├── queries.ts              # TanStack Query hooks
│   └── utils.ts                # formatting utilities
├── components/
│   └── ui/
│       ├── Button.tsx          # primary/secondary/ghost/danger, sm/md/lg
│       ├── Card.tsx            # Header/Body/Footer compound
│       ├── Input.tsx           # text input with label
│       ├── Select.tsx          # dropdown select
│       ├── Table.tsx           # data table with loading/empty/columns
│       ├── Badge.tsx           # colored label badge
│       ├── Tabs.tsx            # tab switcher
│       ├── Skeleton.tsx        # loading placeholders
│       ├── ErrorCard.tsx       # error display with retry
│       ├── EmptyState.tsx      # empty data display
│       └── SearchBar.tsx       # global search input
├── styles/
│   └── globals.css             # Tailwind import + @theme tokens
├── public/
│   └── favicon.svg
├── app.config.ts               # TanStack Start config
├── package.json
├── tsconfig.json
└── biome.json
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `warspark-web/` directory with all config files
- Create: `warspark-web/styles/globals.css`
- Create: `warspark-web/public/favicon.svg`
- Create: `warspark-web/app.config.ts`

**Interfaces:**
- Produces: working `pnpm dev` that shows blank page with Tailwind

- [ ] **Step 1: Initialize project**

```bash
cd ~/code
mkdir warspark-web && cd warspark-web
pnpm init
```

- [ ] **Step 2: Install dependencies**

```bash
pnpm add @tanstack/react-start @tanstack/react-router @tanstack/react-query react react-dom
pnpm add -D @types/react @types/react-dom @biomejs/biome tailwindcss @tailwindcss/vite typescript vite
```

- [ ] **Step 3: Create package.json scripts**

```json
{
  "name": "warspark-web",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite dev",
    "build": "vite build",
    "preview": "vite preview",
    "lint": "biome check .",
    "format": "biome format --write ."
  },
  "dependencies": {
    "@tanstack/react-query": "^5",
    "@tanstack/react-router": "^1",
    "@tanstack/react-start": "^1",
    "react": "^18",
    "react-dom": "^18"
  },
  "devDependencies": {
    "@biomejs/biome": "^1",
    "@tailwindcss/vite": "^4",
    "@types/react": "^18",
    "@types/react-dom": "^18",
    "tailwindcss": "^4",
    "typescript": "^5",
    "vite": "^6"
  }
}
```

- [ ] **Step 4: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "paths": { "~/*": ["./*"] },
    "baseUrl": "."
  },
  "include": ["**/*.ts", "**/*.tsx"]
}
```

- [ ] **Step 5: Create vite.config.ts**

```ts
import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [tailwindcss()],
})
```

- [ ] **Step 6: Create biome.json**

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.4/schema.json",
  "formatter": {
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "linter": {
    "enabled": true,
    "rules": { "recommended": true }
  }
}
```

- [ ] **Step 7: Create styles/globals.css**

```css
@import "tailwindcss";

@theme {
  --color-primary: #3A4DF0;
  --color-primary-light: #8E9AFF;
  --color-primary-dark: #2E3DC4;
  --color-accent: #F05B3A;
  --color-accent-light: #FDBD41;
  --color-surface-page: #F7F9FD;
  --color-surface-card: #FFFFFF;
  --color-text-primary: #1A1A2E;
  --color-text-secondary: #6B7280;
  --color-border: #E5E7EB;
}

body {
  background: var(--color-surface-page);
  color: var(--color-text-primary);
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
}
```

- [ ] **Step 8: Create public/favicon.svg**

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <rect width="32" height="32" rx="6" fill="#3A4DF0"/>
  <text x="16" y="23" text-anchor="middle" fill="white" font-size="20" font-weight="bold" font-family="sans-serif">W</text>
</svg>
```

- [ ] **Step 9: Create app.config.ts**

```ts
import { defineConfig } from '@tanstack/react-start/config'

export default defineConfig({
  tsr: { appDirectory: 'app' },
})
```

- [ ] **Step 10: Create minimal app/__root.tsx and app/index.tsx to verify**

`app/__root.tsx`:
```tsx
import { Outlet, createRootRoute } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: () => (
    <div className="min-h-screen bg-surface-page">
      <Outlet />
    </div>
  ),
})
```

`app/index.tsx`:
```tsx
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  component: () => (
    <h1 className="text-3xl font-bold text-primary p-12">WarSpark</h1>
  ),
})
```

- [ ] **Step 11: Create app/routeTree.gen.ts**

```ts
/* eslint-disable */
import { Route as rootRoute } from './__root'
import { Route as IndexRoute } from './index'

declare module '@tanstack/react-router' {
  interface FileRoutesByPath {
    '/': { preLoaderRoute: typeof IndexRoute }
  }
}

export const routeTree = rootRoute.addChildren([IndexRoute])
```

- [ ] **Step 12: Verify**

```bash
pnpm dev
```

Open http://localhost:5173 — should show "WarSpark" in blue text on light background.

- [ ] **Step 13: Create app/ssr.tsx and app/client.tsx**

`app/ssr.tsx`:
```tsx
import { createRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'

export function createRouter() {
  return createRouter({ routeTree })
}
```

`app/client.tsx`:
```tsx
import { StartClient } from '@tanstack/react-start/client'
import './styles/globals.css'

export default StartClient
```

- [ ] **Step 14: Create index.html for Vite entry**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>WarSpark</title>
  <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/app/client.tsx"></script>
</body>
</html>
```

- [ ] **Step 15: Commit**

```bash
git init && git add -A && git commit -m "feat: scaffold WarSpark Web project"
```

---

### Task 2: Design System Foundation — CSS & Global Styles

**Files:**
- Modify: `warspark-web/styles/globals.css`
- Create: `warspark-web/app/__root.tsx` (update with Navbar)

**Interfaces:**
- Produces: `@theme` tokens available in all components, Navbar shell rendered

- [ ] **Step 1: Expand globals.css with full design tokens**

`styles/globals.css`:
```css
@import "tailwindcss";

@theme {
  --color-primary: #3A4DF0;
  --color-primary-light: #8E9AFF;
  --color-primary-dark: #2E3DC4;
  --color-accent: #F05B3A;
  --color-accent-light: #FDBD41;
  --color-surface-page: #F7F9FD;
  --color-surface-card: #FFFFFF;
  --color-text-primary: #1A1A2E;
  --color-text-secondary: #6B7280;
  --color-border: #E5E7EB;
  --color-success: #22C55E;
  --color-danger: #EF4444;
  --color-warning: #F59E0B;
  --radius-btn: 12px;
  --radius-card: 12px;
  --radius-tag: 8px;
  --shadow-card: 0 2px 12px rgba(220, 232, 255, 0.5);
  --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
}

body {
  background: var(--color-surface-page);
  color: var(--color-text-primary);
  font-family: var(--font-sans);
  -webkit-font-smoothing: antialiased;
}

/* Tabular numbers for data tables */
.tabular-nums { font-variant-numeric: tabular-nums; }
```

- [ ] **Step 2: Update __root.tsx with Navbar layout**

`app/__root.tsx`:
```tsx
import { Outlet, createRootRoute } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  return (
    <div className="min-h-screen bg-surface-page flex flex-col">
      <nav className="h-16 bg-surface-card border-b border-border flex items-center px-6 shadow-sm sticky top-0 z-50">
        <a href="/" className="text-xl font-bold text-primary mr-8">
          🔥 WarSpark
        </a>
        <div className="flex gap-6 text-sm font-medium">
          <a href="/" className="text-text-primary hover:text-primary transition-colors">首页</a>
          <a href="/leaderboards" className="text-text-secondary hover:text-primary transition-colors">排行榜</a>
          <a href="/search" className="text-text-secondary hover:text-primary transition-colors">部落搜索</a>
          <a href="/gold-pass" className="text-text-secondary hover:text-primary transition-colors">Gold Pass</a>
        </div>
        <div className="ml-auto">
          {/* SearchBar placeholder — Task 5 will replace */}
        </div>
      </nav>
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="text-center text-text-secondary text-xs py-6 border-t border-border">
        WarSpark · Clash of Clans 数据平台
      </footer>
    </div>
  )
}
```

- [ ] **Step 3: Update index.tsx placeholder**

`app/index.tsx`:
```tsx
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  component: () => (
    <div className="max-w-page mx-auto py-12 px-4">
      <h1 className="text-4xl font-bold text-text-primary">WarSpark</h1>
      <p className="text-text-secondary mt-2 text-lg">
        Clash of Clans 数据查询平台
      </p>
    </div>
  ),
})
```

- [ ] **Step 4: Verify — run dev and check Navbar renders**

```bash
pnpm dev
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add design system tokens and root layout with navbar"
```

---

### Task 3: API Layer — types.ts + api.ts + utils.ts

**Files:**
- Create: `warspark-web/lib/types.ts`
- Create: `warspark-web/lib/api.ts`
- Create: `warspark-web/lib/utils.ts`

**Interfaces:**
- Produces: `getClan(tag)`, `getPlayer(tag)`, `getWarLog(tag)`, `getLocations()`, `getClanRanking(locId)`, `searchClans(params)`, `getCurrentWar(tag)`, `getCWLGroup(tag)`, `getCurrentGoldPass()` — all typed
- Produces: `normalizeTag()`, `formatNumber()`, `formatTime()`, `warResultLabel()`, `roleLabel()`

- [ ] **Step 1: Create lib/types.ts**

```ts
// === Common ===
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

export interface Paging {
  cursors: { after?: string; before?: string }
}

export interface PagedList<T> {
  items: T[]
  paging: Paging
}

export interface Label {
  id: number
  name: string
  iconUrls: { small: string; medium: string }
}

export interface Location {
  id: number
  name: string
  countryCode?: string
}

export interface LeagueRef {
  id: number
  name: string
}

// === Clan ===
export interface ClanOverview {
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

export interface ClanDetail {
  clan: ClanOverview
  members: ClanMemberSummary[]
}

export interface ClanMemberSummary {
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

// === Player ===
export interface PlayerOverview {
  tag: string
  name: string
  townHallLevel: number
  townHallWeaponLevel?: number
  expLevel: number
  role: string
  warStars: number
  attackWins: number
  defenseWins: number
  trophies: number
  bestTrophies: number
  warPreference?: string
  clan?: PlayerClanInfo
  league?: LeagueRef
  labels: Label[]
  heroes: HeroLevel[]
  achievements: AchievementProgress[]
}

export interface PlayerClanInfo {
  tag: string
  name: string
  clanLevel: number
  badgeUrls: { small: string; medium: string }
}

export interface HeroLevel {
  name: string
  level: number
  maxLevel: number
  village: string
}

export interface AchievementProgress {
  name: string
  stars: number
  target: number
  value: number
  village: string
  info: string
}

// === War ===
export interface WarLogEntry {
  result: string
  teamSize: number
  endTime: string
  clan: WarLogClan
  opponent: WarLogClan
}

export interface WarLogClan {
  tag: string
  name: string
  clanLevel: number
  stars: number
  destructionPercentage: number
  attacks: number
}

export interface CurrentWar {
  state: string
  teamSize: number
  clan: WarClan
  opponent: WarClan
}

export interface WarClan {
  tag: string
  name: string
  stars: number
  destructionPercentage: number
  clanLevel: number
  attacks: number
}

export interface CWLGroup {
  state: string
  season: string
  clans: CWLClan[]
  rounds: CWLRound[]
}

export interface CWLClan {
  tag: string
  name: string
  clanLevel: number
  members: number
}

export interface CWLRound {
  warTags: string[]
}

// === Rankings ===
export interface ClanRankingEntry {
  tag: string
  name: string
  location: Location
  clanLevel: number
  members: number
  clanPoints: number
  rank: number
  badgeUrls: { small: string; medium: string }
}

export interface PlayerRankingEntry {
  tag: string
  name: string
  expLevel: number
  trophies: number
  rank: number
  clan?: { tag: string; name: string; badgeUrls: { small: string } }
}

// === Battle Log ===
export interface BattleLogEntry {
  battleTime: string
  battleType: string
  attack: boolean
  stars: number
  destructionPercentage: number
  opponentName: string
  opponentTag: string
  opponentTownHallLevel: number
}

export interface BattleLogSummary {
  items: BattleLogEntry[]
}

// === Gold Pass ===
export interface GoldPassSeason {
  startTime: string
  endTime: string
}

// === Capital Raids ===
export interface CapitalRaidSeason {
  state: string
  startTime: string
  endTime: string
  capitalTotalLoot: number
  raidsCompleted: number
  totalAttacks: number
  offensiveReward: number
  defensiveReward: number
  enemyDistrictsDestroyed: number
}

// === League Group ===
export interface PlayerLeagueGroup {
  members: LeagueGroupMember[]
  attackLogs: LeagueBattleLogEntry[]
  defenseLogs: LeagueBattleLogEntry[]
}

export interface LeagueGroupMember {
  playerName: string
  playerTag: string
  clanName: string
  clanTag: string
  leagueTrophies: number
  attackWinCount: number
  attackLoseCount: number
  defenseWinCount: number
  defenseLoseCount: number
}

export interface LeagueBattleLogEntry {
  creationTime: string
  destructionPercentage: number
  opponentName: string
  stars: number
  trophies: number
}

// === Search ===
export interface ClanSearchParams {
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
  labelIds?: string
}
```

- [ ] **Step 2: Create lib/api.ts**

```ts
import type { ApiResponse, GoldPassSeason, PagedList } from './types'
import type {
  ClanDetail, ClanOverview, ClanRankingEntry, ClanSearchParams,
  CurrentWar, CWLGroup, PlayerOverview, PlayerRankingEntry,
  BattleLogSummary, CapitalRaidSeason, PlayerLeagueGroup, Location,
  WarLogEntry,
} from './types'
import { ApiError } from './types'

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

function enc(tag: string) { return encodeURIComponent(tag) }

// Clan
export async function getClan(tag: string) {
  return fetchApi<ClanDetail>(`/clans/${enc(tag)}`)
}

export async function getWarLog(tag: string, limit = 10, after?: string) {
  const params = new URLSearchParams({ limit: String(limit) })
  if (after) params.set('after', after)
  return fetchApi<PagedList<WarLogEntry>>(`/clans/${enc(tag)}/war-log?${params}`)
}

export async function getCapitalRaidSeasons(tag: string) {
  return fetchApi<PagedList<CapitalRaidSeason>>(`/clans/${enc(tag)}/capital-raid-seasons`)
}

// Player
export async function getPlayer(tag: string) {
  return fetchApi<PlayerOverview>(`/players/${enc(tag)}`)
}

export async function getBattleLog(tag: string) {
  return fetchApi<BattleLogSummary>(`/players/${enc(tag)}/battle-log`)
}

export async function getPlayerLeagueGroup(tag: string) {
  return fetchApi<PlayerLeagueGroup>(`/players/${enc(tag)}/league-group`)
}

// Rankings
export async function getLocations() {
  return fetchApi<PagedList<Location>>('/locations')
}

export async function getClanRanking(locationId: string) {
  return fetchApi<PagedList<ClanRankingEntry>>(`/locations/${locationId}/rankings/clans`)
}

export async function getPlayerRanking(locationId: string) {
  return fetchApi<PagedList<PlayerRankingEntry>>(`/locations/${locationId}/rankings/players`)
}

export async function getClanCapitalRanking(locationId: string) {
  return fetchApi<PagedList<ClanRankingEntry>>(`/locations/${locationId}/rankings/clans-capital`)
}

export async function getClanBuilderBaseRanking(locationId: string) {
  return fetchApi<PagedList<ClanRankingEntry>>(`/locations/${locationId}/rankings/clans-builder-base`)
}

export async function getPlayerBuilderBaseRanking(locationId: string) {
  return fetchApi<PagedList<PlayerRankingEntry>>(`/locations/${locationId}/rankings/players-builder-base`)
}

// Search
export async function searchClans(params: ClanSearchParams) {
  const sp = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v != null && v !== '') sp.set(k, String(v))
  }
  return fetchApi<PagedList<ClanOverview>>(`/clans?${sp}`)
}

// War
export async function getCurrentWar(clanTag: string) {
  return fetchApi<CurrentWar>(`/war/current?clan_tag=${enc(clanTag)}`)
}

export async function getCWLGroup(clanTag: string) {
  return fetchApi<CWLGroup>(`/war/cwl?clan_tag=${enc(clanTag)}`)
}

export async function getCWLWar(warTag: string) {
  return fetchApi<CurrentWar>(`/war/cwl/wars/${enc(warTag)}`)
}

// Gold Pass
export async function getCurrentGoldPass() {
  return fetchApi<GoldPassSeason>('/gold-pass/current')
}
```

- [ ] **Step 3: Create lib/utils.ts**

```ts
export function normalizeTag(tag: string): string {
  let t = tag.trim().toUpperCase()
  if (!t.startsWith('#')) t = '#' + t
  return t
}

export function formatNumber(n: number): string {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toLocaleString()
}

export function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getMonth() + 1}-${d.getDate()}`
}

export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

export function daysRemaining(endTime: string): number {
  const diff = new Date(endTime).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

export function warResultLabel(result: string): { text: string; color: string } {
  switch (result) {
    case 'win': return { text: '胜利', color: 'text-green-600' }
    case 'lose': return { text: '失败', color: 'text-red-600' }
    case 'tie': return { text: '平局', color: 'text-yellow-600' }
    default: return { text: result, color: 'text-gray-500' }
  }
}

export function roleLabel(role: string): string {
  const m: Record<string, string> = {
    leader: '首领', coLeader: '副首领', admin: '长老', member: '成员'
  }
  return m[role] ?? role
}

export function warStateLabel(state: string): string {
  const m: Record<string, string> = {
    notInWar: '未参战', preparation: '准备日', inWar: '战斗中', warEnded: '已结束'
  }
  return m[state] ?? state
}

export function leagueIconUrl(league: { id?: number; name?: string }): string | null {
  if (!league?.id) return null
  return `https://api-assets.clashofclans.com/leagues/${league.id}.png`
}
```

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: add API layer — types, fetch wrapper, and utilities"
```

---

### Task 4: UI Component Library

**Files:**
- Create: `warspark-web/components/ui/Button.tsx`
- Create: `warspark-web/components/ui/Card.tsx`
- Create: `warspark-web/components/ui/Input.tsx`
- Create: `warspark-web/components/ui/Select.tsx`
- Create: `warspark-web/components/ui/Badge.tsx`
- Create: `warspark-web/components/ui/Tabs.tsx`
- Create: `warspark-web/components/ui/Skeleton.tsx`
- Create: `warspark-web/components/ui/ErrorCard.tsx`
- Create: `warspark-web/components/ui/EmptyState.tsx`
- Create: `warspark-web/components/ui/Table.tsx`
- Create: `warspark-web/components/ui/SearchBar.tsx`

**Interfaces:**
- Produces: all UI components available for page composition

- [ ] **Step 1: Create Button.tsx**

```tsx
import type { ButtonHTMLAttributes, ReactNode } from 'react'

const variants = {
  primary: 'bg-primary text-white hover:bg-primary-dark shadow-sm',
  secondary: 'bg-white text-primary border border-primary hover:bg-primary/5',
  ghost: 'text-text-secondary hover:text-primary hover:bg-surface-page',
  danger: 'bg-danger text-white hover:bg-red-700',
} as const

const sizes = {
  sm: 'h-8 px-4 text-sm rounded-tag',
  md: 'h-10 px-6 text-sm rounded-btn',
  lg: 'h-12 px-8 text-base rounded-btn',
} as const

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variants
  size?: keyof typeof sizes
  loading?: boolean
  children: ReactNode
}

export function Button({ variant = 'primary', size = 'md', loading, children, className = '', ...props }: ButtonProps) {
  return (
    <button
      className={`inline-flex items-center justify-center gap-2 font-medium transition-colors
        ${variants[variant]} ${sizes[size]}
        ${loading ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'}
        ${className}`}
      disabled={loading}
      {...props}
    >
      {loading && <Spinner />}
      {children}
    </button>
  )
}

function Spinner() {
  return <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
}
```

- [ ] **Step 2: Create Card.tsx**

```tsx
import type { ReactNode } from 'react'

interface CardProps { children: ReactNode; className?: string }
interface CompoundProps { children: ReactNode; className?: string }

export function Card({ children, className = '' }: CardProps) {
  return (
    <div className={`bg-surface-card rounded-card shadow-card ${className}`}>
      {children}
    </div>
  )
}

Card.Header = function CardHeader({ children, className = '' }: CompoundProps) {
  return <div className={`px-6 pt-6 pb-4 ${className}`}>{children}</div>
}

Card.Body = function CardBody({ children, className = '' }: CompoundProps) {
  return <div className={`px-6 pb-6 ${className}`}>{children}</div>
}

Card.Footer = function CardFooter({ children, className = '' }: CompoundProps) {
  return <div className={`px-6 pb-6 flex gap-3 ${className}`}>{children}</div>
}
```

- [ ] **Step 3: Create Input.tsx**

```tsx
import type { InputHTMLAttributes } from 'react'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function Input({ label, error, className = '', ...props }: InputProps) {
  return (
    <div className="flex flex-col gap-1">
      {label && <label className="text-sm font-medium text-text-secondary">{label}</label>}
      <input
        className={`h-10 px-4 rounded-tag border border-border bg-white text-sm
          focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary
          placeholder:text-text-secondary/50 ${error ? 'border-danger' : ''} ${className}`}
        {...props}
      />
      {error && <span className="text-xs text-danger">{error}</span>}
    </div>
  )
}
```

- [ ] **Step 4: Create Select.tsx**

```tsx
import type { SelectHTMLAttributes } from 'react'

interface SelectOption { value: string; label: string }

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  options: SelectOption[]
  placeholder?: string
}

export function Select({ label, options, placeholder, className = '', ...props }: SelectProps) {
  return (
    <div className="flex flex-col gap-1">
      {label && <label className="text-sm font-medium text-text-secondary">{label}</label>}
      <select
        className={`h-10 px-4 rounded-tag border border-border bg-white text-sm
          focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary ${className}`}
        {...props}
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  )
}
```

- [ ] **Step 5: Create Badge.tsx**

```tsx
import type { ReactNode } from 'react'

const colors = {
  green: 'bg-green-100 text-green-700',
  red: 'bg-red-100 text-red-700',
  yellow: 'bg-yellow-100 text-yellow-700',
  blue: 'bg-blue-100 text-blue-700',
  gray: 'bg-gray-100 text-gray-600',
  primary: 'bg-primary/10 text-primary',
} as const

interface BadgeProps {
  color?: keyof typeof colors
  children: ReactNode
}

export function Badge({ color = 'gray', children }: BadgeProps) {
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-tag text-xs font-medium ${colors[color]}`}>
      {children}
    </span>
  )
}
```

- [ ] **Step 6: Create Tabs.tsx**

```tsx
interface Tab { key: string; label: string }

interface TabsProps {
  tabs: Tab[]
  active: string
  onChange: (key: string) => void
}

export function Tabs({ tabs, active, onChange }: TabsProps) {
  return (
    <div className="flex border-b border-border">
      {tabs.map(tab => (
        <button
          key={tab.key}
          onClick={() => onChange(tab.key)}
          className={`px-6 py-3 text-sm font-medium border-b-2 transition-colors
            ${tab.key === active
              ? 'border-primary text-primary'
              : 'border-transparent text-text-secondary hover:text-text-primary'}`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
```

- [ ] **Step 7: Create Skeleton.tsx**

```tsx
export function Skeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse bg-gray-200 rounded ${className}`} />
}

Skeleton.Card = function SkeletonCard() {
  return (
    <div className="bg-surface-card rounded-card p-6 space-y-4">
      <Skeleton className="h-6 w-1/3" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-2/3" />
    </div>
  )
}

Skeleton.Table = function SkeletonTable({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="space-y-3 p-6">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex gap-4">
          {Array.from({ length: cols }).map((_, j) => (
            <Skeleton key={j} className={`h-4 ${j === 0 ? 'w-12' : 'flex-1'}`} />
          ))}
        </div>
      ))}
    </div>
  )
}
```

- [ ] **Step 8: Create ErrorCard.tsx**

```tsx
import { Button } from './Button'

interface ErrorCardProps {
  title?: string
  message?: string
  onRetry?: () => void
}

export function ErrorCard({ title = '加载失败', message = '请稍后重试', onRetry }: ErrorCardProps) {
  return (
    <div className="bg-surface-card rounded-card p-12 text-center space-y-4">
      <div className="text-4xl">😞</div>
      <h3 className="text-lg font-semibold text-text-primary">{title}</h3>
      <p className="text-text-secondary text-sm">{message}</p>
      {onRetry && <Button variant="secondary" size="sm" onClick={onRetry}>重新加载</Button>}
    </div>
  )
}
```

- [ ] **Step 9: Create EmptyState.tsx**

```tsx
interface EmptyStateProps {
  icon?: string
  title?: string
  description?: string
}

export function EmptyState({ icon = '📭', title = '暂无数据', description }: EmptyStateProps) {
  return (
    <div className="py-16 text-center space-y-3">
      <div className="text-4xl">{icon}</div>
      <h3 className="text-lg font-semibold text-text-primary">{title}</h3>
      {description && <p className="text-text-secondary text-sm">{description}</p>}
    </div>
  )
}
```

- [ ] **Step 10: Create Table.tsx**

```tsx
import { Skeleton } from './Skeleton'
import { EmptyState } from './EmptyState'

interface Column<T> {
  key: string
  header: string
  width?: string
  align?: 'left' | 'right' | 'center'
  render?: (row: T) => React.ReactNode
}

interface TableProps<T> {
  columns: Column<T>[]
  data: T[]
  loading?: boolean
  emptyText?: string
  onRowClick?: (row: T) => void
}

export function Table<T extends Record<string, unknown>>({ columns, data, loading, emptyText, onRowClick }: TableProps<T>) {
  if (loading) return <Skeleton.Table rows={5} cols={columns.length} />
  if (data.length === 0) return <EmptyState title={emptyText ?? '暂无数据'} />

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border">
            {columns.map(col => (
              <th key={col.key}
                className={`py-3 px-4 font-medium text-text-secondary text-left tabular-nums
                  ${col.align === 'right' ? 'text-right' : ''}
                  ${col.align === 'center' ? 'text-center' : ''}`}
                style={col.width ? { width: col.width } : undefined}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr key={i}
              className={`border-b border-border/50 ${onRowClick ? 'cursor-pointer hover:bg-surface-page' : ''}`}
              onClick={() => onRowClick?.(row)}
            >
              {columns.map(col => (
                <td key={col.key}
                  className={`py-3 px-4 tabular-nums
                    ${col.align === 'right' ? 'text-right' : ''}
                    ${col.align === 'center' ? 'text-center' : ''}`}
                >
                  {col.render ? col.render(row) : String(row[col.key] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
```

- [ ] **Step 11: Create SearchBar.tsx**

```tsx
import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'

export function SearchBar() {
  const [value, setValue] = useState('')
  const navigate = useNavigate()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const tag = value.trim().toUpperCase()
    if (!tag) return
    const clean = tag.startsWith('#') ? tag : '#' + tag
    // Heuristic: clan tags are 8-9 chars, player tags are 9-10 chars
    // Both start with # and contain alphanumeric chars
    if (clean.startsWith('#') && clean.length >= 4) {
      navigate({ to: `/clans/${encodeURIComponent(clean)}` })
    }
    setValue('')
  }

  return (
    <form onSubmit={handleSubmit} className="relative">
      <input
        type="text"
        value={value}
        onChange={e => setValue(e.target.value)}
        placeholder="输入部落标签 #2PP..."
        className="h-10 w-48 lg:w-64 pl-4 pr-4 rounded-tag border border-border bg-surface-page text-sm
          focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary
          placeholder:text-text-secondary/50"
      />
    </form>
  )
}
```

- [ ] **Step 12: Add SearchBar to __root.tsx Navbar**

Replace the placeholder comment in Navbar:
```tsx
import { SearchBar } from '~/components/ui/SearchBar'

// In Navbar, after the nav links, replace the placeholder:
<div className="ml-auto">
  <SearchBar />
</div>
```

- [ ] **Step 13: Commit**

```bash
git add -A && git commit -m "feat: add UI component library — Button, Card, Input, Select, Badge, Tabs, Skeleton, ErrorCard, EmptyState, Table, SearchBar"
```

---

### Task 5: TanStack Query Hooks

**Files:**
- Create: `warspark-web/lib/queries.ts`

**Interfaces:**
- Produces: `useClan(tag)`, `usePlayer(tag)`, `useWarLog(tag)`, `useRaidSeasons(tag)`, `useBattleLog(tag)`, `useLeagueGroup(tag)`, `useLocations()`, `useRankings(locId, type)`, `useSearchClans(params)`, `useCurrentWar(tag)`, `useCWLGroup(tag)`, `useCurrentGoldPass()` — all returning standard TanStack Query result type

- [ ] **Step 1: Create lib/queries.ts**

```ts
import { useQuery } from '@tanstack/react-query'
import {
  getClan, getPlayer, getWarLog, getCapitalRaidSeasons,
  getBattleLog, getPlayerLeagueGroup, getLocations,
  getClanRanking, getPlayerRanking, getClanCapitalRanking,
  getClanBuilderBaseRanking, getPlayerBuilderBaseRanking,
  searchClans, getCurrentWar, getCWLGroup, getCurrentGoldPass,
} from './api'
import type { ClanSearchParams } from './types'

const STALE = {
  clan: 5 * 60 * 1000,
  player: 5 * 60 * 1000,
  warLog: 5 * 60 * 1000,
  ranking: 10 * 60 * 1000,
  goldPass: 60 * 60 * 1000,
  search: 2 * 60 * 1000,
}

export function useClan(tag: string) {
  return useQuery({
    queryKey: ['clan', tag],
    queryFn: () => getClan(tag),
    staleTime: STALE.clan,
    enabled: !!tag,
  })
}

export function usePlayer(tag: string) {
  return useQuery({
    queryKey: ['player', tag],
    queryFn: () => getPlayer(tag),
    staleTime: STALE.player,
    enabled: !!tag,
  })
}

export function useWarLog(tag: string) {
  return useQuery({
    queryKey: ['warLog', tag],
    queryFn: () => getWarLog(tag),
    staleTime: STALE.warLog,
    enabled: !!tag,
  })
}

export function useRaidSeasons(tag: string) {
  return useQuery({
    queryKey: ['raidSeasons', tag],
    queryFn: () => getCapitalRaidSeasons(tag),
    staleTime: STALE.clan,
    enabled: !!tag,
  })
}

export function useBattleLog(tag: string) {
  return useQuery({
    queryKey: ['battleLog', tag],
    queryFn: () => getBattleLog(tag),
    staleTime: STALE.player,
    enabled: !!tag,
  })
}

export function useLeagueGroup(tag: string) {
  return useQuery({
    queryKey: ['leagueGroup', tag],
    queryFn: () => getPlayerLeagueGroup(tag),
    staleTime: STALE.player,
    enabled: !!tag,
  })
}

export function useLocations() {
  return useQuery({
    queryKey: ['locations'],
    queryFn: () => getLocations(),
    staleTime: STALE.ranking,
  })
}

export type RankingType = 'clans' | 'players' | 'capital' | 'builder-clans' | 'builder-players'

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
    staleTime: STALE.ranking,
    enabled: !!locationId,
  })
}

export function useSearchClans(params: ClanSearchParams) {
  return useQuery({
    queryKey: ['searchClans', params],
    queryFn: () => searchClans(params),
    staleTime: STALE.search,
    enabled: !!(params.name || params.locationId || params.warFrequency),
  })
}

export function useCurrentWar(tag: string) {
  return useQuery({
    queryKey: ['currentWar', tag],
    queryFn: () => getCurrentWar(tag),
    refetchInterval: 30_000,
    staleTime: 30_000,
    enabled: !!tag,
  })
}

export function useCWLGroup(tag: string) {
  return useQuery({
    queryKey: ['cwlGroup', tag],
    queryFn: () => getCWLGroup(tag),
    staleTime: 5 * 60 * 1000,
    enabled: !!tag,
  })
}

export function useCurrentGoldPass() {
  return useQuery({
    queryKey: ['goldPass'],
    queryFn: () => getCurrentGoldPass(),
    staleTime: STALE.goldPass,
  })
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: add TanStack Query hooks for all API endpoints"
```

---

### Task 6: Home Page

**Files:**
- Modify: `warspark-web/app/index.tsx`

**Interfaces:**
- Consumes: `useLocations`, `useRankings` from lib/queries.ts
- Produces: home page with search hero + top rankings preview

- [ ] **Step 1: Rewrite app/index.tsx**

```tsx
import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useLocations, useRankings } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Button } from '~/components/ui/Button'
import { Skeleton } from '~/components/ui/Skeleton'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { formatNumber } from '~/lib/utils'
import type { ClanRankingEntry, PlayerRankingEntry, Location } from '~/lib/types'

export const Route = createFileRoute('/')({
  component: HomePage,
})

function HomePage() {
  return (
    <div>
      <HeroSection />
      <TopRankings />
    </div>
  )
}

function HeroSection() {
  const [value, setValue] = useState('')
  const navigate = useNavigate()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    const tag = value.trim().toUpperCase()
    if (!tag) return
    const clean = tag.startsWith('#') ? tag : '#' + tag
    if (clean.startsWith('#') && clean.length >= 4) {
      navigate({ to: '/clans/$tag', params: { tag: clean } })
    }
    setValue('')
  }

  return (
    <div className="bg-gradient-to-r from-primary to-primary-light py-20 px-4">
      <div className="max-w-page mx-auto text-center">
        <h1 className="text-4xl font-bold text-white mb-4">WarSpark</h1>
        <p className="text-white/80 text-lg mb-8">Clash of Clans 数据查询平台</p>
        <form onSubmit={handleSearch} className="max-w-md mx-auto">
          <div className="flex gap-3">
            <input
              type="text"
              value={value}
              onChange={e => setValue(e.target.value)}
              placeholder="输入部落标签 #2PP..."
              className="flex-1 h-12 px-6 rounded-btn border-none text-sm
                focus:outline-none focus:ring-2 focus:ring-white/30"
            />
            <Button type="submit" size="lg" className="bg-white text-primary hover:bg-white/90">
              搜索
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

function TopRankings() {
  const { data: locData, isLoading: locLoading } = useLocations()
  const globalId = '32000006'

  return (
    <div className="max-w-page mx-auto py-12 px-4">
      <h2 className="text-2xl font-bold text-text-primary mb-8">热门排行榜</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <RankingColumn title="TOP 部落" type="clans" locationId={globalId} linkTo="/leaderboards" />
        <RankingColumn title="TOP 玩家" type="players" locationId={globalId} linkTo="/leaderboards?type=players" />
        <RankingColumn title="TOP 都城" type="capital" locationId={globalId} linkTo="/leaderboards?type=capital" />
      </div>
    </div>
  )
}

function RankingColumn({ title, type, locationId, linkTo }: {
  title: string
  type: 'clans' | 'players' | 'capital'
  locationId: string
  linkTo: string
}) {
  const { data, isLoading, isError, refetch } = useRankings(locationId, type)

  return (
    <Link to={linkTo} className="block">
      <Card className="hover:shadow-lg transition-shadow">
        <Card.Header>
          <h3 className="font-semibold text-text-primary">{title}</h3>
        </Card.Header>
        <Card.Body>
          {isLoading && <Skeleton className="h-32 w-full" />}
          {isError && <ErrorCard title="加载失败" onRetry={() => refetch()} />}
          {data && (
            <div className="space-y-2">
              {(data.items as Array<{ name: string; clanLevel?: number; trophies?: number; clanPoints?: number }>)
                .slice(0, 5).map((item, i) => (
                <div key={i} className="flex items-center gap-3 text-sm">
                  <span className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold text-white
                    ${i === 0 ? 'bg-yellow-500' : i === 1 ? 'bg-gray-400' : i === 2 ? 'bg-amber-600' : 'bg-gray-300 text-gray-500'}`}>
                    {i + 1}
                  </span>
                  <span className="flex-1 truncate">{item.name}</span>
                  {'trophies' in item && <span className="tabular-nums text-text-secondary">{formatNumber(item.trophies!)}</span>}
                  {'clanPoints' in item && <span className="tabular-nums text-text-secondary">{formatNumber(item.clanPoints!)}</span>}
                </div>
              ))}
            </div>
          )}
        </Card.Body>
        <Card.Footer>
          <span className="text-sm text-primary">查看更多 →</span>
        </Card.Footer>
      </Card>
    </Link>
  )
}
```

- [ ] **Step 2: Verify — visit http://localhost:5173**

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: add home page with search hero and top rankings preview"
```

---

### Task 7: Clan Detail Page

**Files:**
- Create: `warspark-web/app/clans.$tag.tsx`

**Interfaces:**
- Consumes: `useClan`, `useWarLog`, `useRaidSeasons` from lib/queries.ts
- Produces: clan detail page with tabs (Overview / War Log / Capital Raids)

- [ ] **Step 1: Create app/clans.$tag.tsx**

```tsx
import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useClan, useWarLog, useRaidSeasons } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Tabs } from '~/components/ui/Tabs'
import { Badge } from '~/components/ui/Badge'
import { Table } from '~/components/ui/Table'
import { Skeleton } from '~/components/ui/Skeleton'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { formatNumber, formatDate, warResultLabel, roleLabel, normalizeTag } from '~/lib/utils'
import type { ClanMemberSummary, WarLogEntry, CapitalRaidSeason } from '~/lib/types'

export const Route = createFileRoute('/clans/$tag')({
  component: ClanDetailPage,
})

const TABS = [
  { key: 'overview', label: '概览' },
  { key: 'warLog', label: '战争日志' },
  { key: 'raids', label: '突袭赛季' },
]

function ClanDetailPage() {
  const { tag } = Route.useParams()
  const normalized = normalizeTag(tag)
  const { data, isLoading, isError, refetch } = useClan(normalized)
  const [activeTab, setActiveTab] = useState('overview')

  if (isError) return <div className="max-w-page mx-auto p-12"><ErrorCard onRetry={() => refetch()} /></div>

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      {isLoading ? (
        <Skeleton.Card />
      ) : data ? (
        <ClanHeader clan={data.clan} />
      ) : null}

      <Card>
        <Tabs tabs={TABS} active={activeTab} onChange={setActiveTab} />
        {activeTab === 'overview' && <OverviewTab tag={normalized} data={data} isLoading={isLoading} />}
        {activeTab === 'warLog' && <WarLogTab tag={normalized} />}
        {activeTab === 'raids' && <RaidsTab tag={normalized} />}
      </Card>
    </div>
  )
}

function ClanHeader({ clan }: { clan: NonNullable<ReturnType<typeof useClan>['data']>['clan'] }) {
  return (
    <Card>
      <Card.Body className="flex items-start gap-4">
        <img src={clan.badgeUrls.medium} alt="" className="w-16 h-16 rounded-lg" />
        <div className="space-y-1 flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold text-text-primary">{clan.name}</h1>
            <span className="text-text-secondary text-sm">{clan.tag}</span>
          </div>
          <div className="flex gap-2 flex-wrap">
            <Badge color="primary">Lv.{clan.clanLevel}</Badge>
            <Badge color="gray">{clan.type === 'open' ? '开放' : '仅限邀请'}</Badge>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mt-2 text-sm text-text-secondary">
            <div>成员: <span className="text-text-primary font-medium">{clan.members}</span></div>
            <div>积分: <span className="text-text-primary font-medium tabular-nums">{formatNumber(clan.clanPoints)}</span></div>
            <div>胜/负/平: <span className="text-text-primary font-medium">{clan.warWins}/{clan.warLosses ?? '-'}/{clan.warTies ?? '-'}</span></div>
            <div>连胜: <span className="text-text-primary font-medium">{clan.warWinStreak}</span></div>
          </div>
        </div>
      </Card.Body>
    </Card>
  )
}

function OverviewTab({ tag, data, isLoading }: { tag: string; data: ReturnType<typeof useClan>['data']; isLoading: boolean }) {
  if (isLoading) return <Skeleton.Table rows={10} cols={6} />
  if (!data) return null

  return (
    <Table
      columns={[
        { key: 'clanRank', header: '#', width: '60px', align: 'center' },
        { key: 'name', header: '名称', render: (row: ClanMemberSummary) => (
          <div>
            <div className="font-medium text-text-primary">{row.name}</div>
            <div className="text-xs text-text-secondary">{row.tag}</div>
          </div>
        )},
        { key: 'role', header: '角色', render: (row: ClanMemberSummary) => roleLabel(row.role) },
        { key: 'expLevel', header: '等级', align: 'center' },
        { key: 'townHallLevel', header: '大本营', align: 'center',
          render: (row: ClanMemberSummary) => `TH${row.townHallLevel}` },
        { key: 'trophies', header: '奖杯', align: 'right',
          render: (row: ClanMemberSummary) => formatNumber(row.trophies) },
        { key: 'donations', header: '捐兵', align: 'right',
          render: (row: ClanMemberSummary) => formatNumber(row.donations) },
      ]}
      data={data.members}
    />
  )
}

function WarLogTab({ tag }: { tag: string }) {
  const { data, isLoading, isError, refetch } = useWarLog(tag)

  if (isError) return <div className="p-6"><ErrorCard onRetry={() => refetch()} /></div>

  return (
    <Table
      loading={isLoading}
      emptyText="暂无战争日志"
      columns={[
        { key: 'endTime', header: '日期', render: (row: WarLogEntry) => formatDate(row.endTime) },
        { key: 'opponent', header: '对手', render: (row: WarLogEntry) => row.opponent.name },
        { key: 'teamSize', header: '规模', align: 'center', render: (row: WarLogEntry) => `${row.teamSize}v${row.teamSize}` },
        { key: 'result', header: '结果', render: (row: WarLogEntry) => {
          const r = warResultLabel(row.result)
          return <Badge color={r.color === 'text-green-600' ? 'green' : r.color === 'text-red-600' ? 'red' : 'yellow'}>{r.text}</Badge>
        }},
        { key: 'stars', header: '星数', align: 'center',
          render: (row: WarLogEntry) => `${row.clan.stars} - ${row.opponent.stars}` },
        { key: 'destruction', header: '摧毁%', align: 'right',
          render: (row: WarLogEntry) => `${row.clan.destructionPercentage.toFixed(1)}%` },
      ]}
      data={data?.items ?? []}
    />
  )
}

function RaidsTab({ tag }: { tag: string }) {
  const { data, isLoading, isError, refetch } = useRaidSeasons(tag)

  if (isError) return <div className="p-6"><ErrorCard onRetry={() => refetch()} /></div>
  if (isLoading) return <div className="p-6"><Skeleton.Table rows={3} cols={5} /></div>
  if (!data?.items.length) return <div className="p-6 text-center text-text-secondary">暂无突袭数据</div>

  return (
    <div className="p-6 grid gap-4 md:grid-cols-2">
      {data.items.map((season, i) => (
        <Card key={i}>
          <Card.Body className="space-y-2">
            <div className="flex items-center justify-between">
              <Badge color={season.state === 'ongoing' ? 'green' : 'gray'}>
                {season.state === 'ongoing' ? '进行中' : '已结束'}
              </Badge>
              <span className="text-xs text-text-secondary">{formatDate(season.startTime)} - {formatDate(season.endTime)}</span>
            </div>
            <div className="grid grid-cols-3 gap-2 text-sm">
              <div>突袭完成<span className="block font-semibold tabular-nums">{season.raidsCompleted}</span></div>
              <div>总攻击<span className="block font-semibold tabular-nums">{season.totalAttacks}</span></div>
              <div>总战利品<span className="block font-semibold tabular-nums">{formatNumber(season.capitalTotalLoot)}</span></div>
            </div>
          </Card.Body>
        </Card>
      ))}
    </div>
  )
}
```

- [ ] **Step 2: Update routeTree.gen.ts to include the new route**

```ts
import { Route as ClanDetailRoute } from './clans.$tag'

// Add to FileRoutesByPath:
'/clans/$tag': { preLoaderRoute: typeof ClanDetailRoute }

// Add to routeTree:
export const routeTree = rootRoute.addChildren([IndexRoute, ClanDetailRoute])
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: add clan detail page with tabs (overview, war log, capital raids)"
```

---

### Task 8: Player Detail Page

**Files:**
- Create: `warspark-web/app/players.$tag.tsx`

- [ ] **Step 1: Create app/players.$tag.tsx**

```tsx
import { createFileRoute } from '@tanstack/react-router'
import { usePlayer, useBattleLog, useLeagueGroup } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Badge } from '~/components/ui/Badge'
import { Table } from '~/components/ui/Table'
import { Skeleton } from '~/components/ui/Skeleton'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { formatNumber, formatDateTime, roleLabel, normalizeTag } from '~/lib/utils'
import type { BattleLogEntry, HeroLevel, LeagueGroupMember, LeagueBattleLogEntry } from '~/lib/types'

export const Route = createFileRoute('/players/$tag')({
  component: PlayerDetailPage,
})

function PlayerDetailPage() {
  const { tag } = Route.useParams()
  const normalized = normalizeTag(tag)
  const { data, isLoading, isError, refetch } = usePlayer(normalized)

  if (isError) return <div className="max-w-page mx-auto p-12"><ErrorCard onRetry={() => refetch()} /></div>
  if (isLoading) return <div className="max-w-page mx-auto py-8 px-4"><Skeleton.Card /></div>
  if (!data) return null

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      <PlayerHeader player={data} />
      <HeroGrid heroes={data.heroes} />
      <BattleLogSection tag={normalized} />
      <LeagueGroupSection tag={normalized} />
    </div>
  )
}

function PlayerHeader({ player }: { player: NonNullable<ReturnType<typeof usePlayer>['data']> }) {
  return (
    <Card>
      <Card.Body className="space-y-3">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold text-text-primary">{player.name}</h1>
          <span className="text-text-secondary text-sm">{player.tag}</span>
          <Badge color="primary">TH{player.townHallLevel}</Badge>
          <Badge color="blue">{roleLabel(player.role)}</Badge>
        </div>
        <div className="grid grid-cols-3 md:grid-cols-6 gap-3 text-sm">
          <Stat label="奖杯" value={formatNumber(player.trophies)} highlight={formatNumber(player.bestTrophies)} />
          <Stat label="战争之星" value={formatNumber(player.warStars)} />
          <Stat label="进攻胜场" value={formatNumber(player.attackWins)} />
          <Stat label="防御胜场" value={formatNumber(player.defenseWins)} />
          <Stat label="经验等级" value={formatNumber(player.expLevel)} />
          {player.clan && <Stat label="部落" value={player.clan.name} />}
        </div>
      </Card.Body>
    </Card>
  )
}

function Stat({ label, value, highlight }: { label: string; value: string; highlight?: string }) {
  return (
    <div>
      <div className="text-text-secondary text-xs">{label}</div>
      <div className="font-semibold text-text-primary tabular-nums">{value}</div>
      {highlight && <div className="text-text-secondary text-xs">最高 {highlight}</div>}
    </div>
  )
}

function HeroGrid({ heroes }: { heroes: HeroLevel[] }) {
  if (!heroes.length) return null
  return (
    <Card>
      <Card.Header><h3 className="font-semibold">英雄</h3></Card.Header>
      <Card.Body>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {heroes.map(h => (
            <div key={h.name} className="flex items-center gap-3 p-3 bg-surface-page rounded-lg">
              <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-sm">
                {h.name.charAt(0)}
              </div>
              <div>
                <div className="text-sm font-medium text-text-primary">{h.name}</div>
                <div className="text-xs text-text-secondary tabular-nums">Lv.{h.level} / {h.maxLevel}</div>
              </div>
            </div>
          ))}
        </div>
      </Card.Body>
    </Card>
  )
}

function BattleLogSection({ tag }: { tag: string }) {
  const { data, isLoading } = useBattleLog(tag)

  return (
    <Card>
      <Card.Header><h3 className="font-semibold">战斗日志</h3></Card.Header>
      <Table
        loading={isLoading}
        emptyText="暂无战斗记录"
        columns={[
          { key: 'battleTime', header: '时间', render: (row: BattleLogEntry) => formatDateTime(row.battleTime) },
          { key: 'type', header: '类型', render: (row: BattleLogEntry) => (
            <Badge color={row.attack ? 'green' : 'red'}>{row.attack ? '进攻' : '防守'}</Badge>
          )},
          { key: 'stars', header: '星数', align: 'center',
            render: (row: BattleLogEntry) => '⭐'.repeat(row.stars) },
          { key: 'destruction', header: '摧毁%', align: 'right',
            render: (row: BattleLogEntry) => `${row.destructionPercentage}%` },
          { key: 'opponent', header: '对手', render: (row: BattleLogEntry) => row.opponentName },
          { key: 'opponentTH', header: 'TH', align: 'center',
            render: (row: BattleLogEntry) => row.opponentTownHallLevel },
        ]}
        data={data?.items ?? []}
      />
    </Card>
  )
}

function LeagueGroupSection({ tag }: { tag: string }) {
  const { data, isLoading } = useLeagueGroup(tag)

  if (isLoading) return <Card><Card.Header><h3 className="font-semibold">联赛分组</h3></Card.Header><Skeleton.Table rows={5} cols={4} /></Card>
  if (!data?.members.length) return null

  return (
    <Card>
      <Card.Header><h3 className="font-semibold">联赛分组</h3></Card.Header>
      <Table
        columns={[
          { key: 'player', header: '玩家', render: (row: LeagueGroupMember) => row.playerName },
          { key: 'clan', header: '部落', render: (row: LeagueGroupMember) => row.clanName },
          { key: 'trophies', header: '奖杯', align: 'right',
            render: (row: LeagueGroupMember) => formatNumber(row.leagueTrophies) },
          { key: 'attack', header: '进攻', align: 'center',
            render: (row: LeagueGroupMember) => `${row.attackWinCount}W ${row.attackLoseCount}L` },
          { key: 'defense', header: '防守', align: 'center',
            render: (row: LeagueGroupMember) => `${row.defenseWinCount}W ${row.defenseLoseCount}L` },
        ]}
        data={data.members}
      />
    </Card>
  )
}
```

- [ ] **Step 2: Update routeTree.gen.ts**

```ts
import { Route as PlayerDetailRoute } from './players.$tag'

'/players/$tag': { preLoaderRoute: typeof PlayerDetailRoute }
// add to routeTree children
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: add player detail page with heroes, battle log, and league group"
```

---

### Task 9: Leaderboards Page

**Files:**
- Create: `warspark-web/app/leaderboards.tsx`

- [ ] **Step 1: Create app/leaderboards.tsx**

```tsx
import { createFileRoute, useSearch } from '@tanstack/react-router'
import { useLocations, useRankings, type RankingType } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Select } from '~/components/ui/Select'
import { Tabs } from '~/components/ui/Tabs'
import { Table } from '~/components/ui/Table'
import { Badge } from '~/components/ui/Badge'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { formatNumber } from '~/lib/utils'
import type { ClanRankingEntry, PlayerRankingEntry } from '~/lib/types'

interface LeaderboardSearch {
  location?: string
  type?: RankingType
}

const RANKING_TABS = [
  { key: 'clans', label: '部落' },
  { key: 'players', label: '玩家' },
  { key: 'capital', label: '都城' },
  { key: 'builder-clans', label: '夜世界部落' },
  { key: 'builder-players', label: '夜世界玩家' },
]

export const Route = createFileRoute('/leaderboards')({
  validateSearch: (search: Record<string, string>): LeaderboardSearch => ({
    location: search.location || '32000006',
    type: (search.type as RankingType) || 'clans',
  }),
  component: LeaderboardsPage,
})

function LeaderboardsPage() {
  const { location: locId, type } = Route.useSearch()
  const navigate = Route.useNavigate()

  const { data: locData } = useLocations()
  const { data, isLoading, isError, refetch } = useRankings(locId ?? '32000006', type ?? 'clans')

  const locationOptions = (locData?.items ?? []).map((l: { id: number; name: string }) => ({
    value: String(l.id), label: l.name,
  }))

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      <div className="flex items-center gap-4">
        <h1 className="text-2xl font-bold text-text-primary">排行榜</h1>
        <div className="w-48">
          <Select
            options={locationOptions}
            value={locId}
            onChange={e => navigate({ search: { location: e.target.value, type } })}
          />
        </div>
      </div>

      <Card>
        <Tabs
          tabs={RANKING_TABS}
          active={type ?? 'clans'}
          onChange={key => navigate({ search: { location: locId, type: key as RankingType } })}
        />
        {isError ? (
          <div className="p-6"><ErrorCard onRetry={() => refetch()} /></div>
        ) : (
          <Table
            loading={isLoading}
            emptyText="暂无排名数据"
            columns={[
              { key: 'rank', header: '#', width: '60px', align: 'center',
                render: (row: ClanRankingEntry) => {
                  const color = row.rank <= 3 ? ['bg-yellow-500', 'bg-gray-400', 'bg-amber-600'][row.rank - 1] : 'bg-gray-300 text-gray-500'
                  return <span className={`w-6 h-6 rounded-full inline-flex items-center justify-center text-xs font-bold text-white ${color}`}>{row.rank}</span>
                }
              },
              { key: 'name', header: '名称', render: (row: ClanRankingEntry) => (
                <div>
                  <div className="font-medium text-text-primary">{row.name}</div>
                  <div className="text-xs text-text-secondary">{row.tag}</div>
                </div>
              )},
              { key: 'clanLevel', header: '等级', align: 'center',
                render: (row: ClanRankingEntry) => row.clanLevel && <Badge color="primary">Lv.{row.clanLevel}</Badge> },
              { key: 'clanPoints', header: '积分', align: 'right',
                render: (row: ClanRankingEntry) => 'clanPoints' in row ? formatNumber(row.clanPoints as number) : '-' },
              { key: 'members', header: '成员', align: 'right',
                render: (row: ClanRankingEntry) => 'members' in row ? `${row.members}` : '-' },
            ]}
            data={data?.items ?? []}
          />
        )}
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: Update routeTree.gen.ts**

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: add leaderboards page with location picker and ranking type tabs"
```

---

### Task 10: Search + Gold Pass Pages

**Files:**
- Create: `warspark-web/app/search.tsx`
- Create: `warspark-web/app/gold-pass.tsx`

- [ ] **Step 1: Create app/search.tsx**

```tsx
import { createFileRoute, useSearch } from '@tanstack/react-router'
import { useSearchClans } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Input } from '~/components/ui/Input'
import { Select } from '~/components/ui/Select'
import { Button } from '~/components/ui/Button'
import { Table } from '~/components/ui/Table'
import { Badge } from '~/components/ui/Badge'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { formatNumber } from '~/lib/utils'
import type { ClanOverview, ClanSearchParams } from '~/lib/types'

const WAR_FREQ_OPTIONS = [
  { value: '', label: '不限' },
  { value: 'always', label: '一直打仗' },
  { value: 'moreThanOncePerWeek', label: '每周多次' },
  { value: 'oncePerWeek', label: '每周一次' },
  { value: 'lessThanOncePerWeek', label: '少于每周一次' },
  { value: 'never', label: '从不' },
]

export const Route = createFileRoute('/search')({
  validateSearch: (s: Record<string, string>): ClanSearchParams => ({
    name: s.name, warFrequency: s.warFrequency, locationId: s.locationId ? Number(s.locationId) : undefined,
    minMembers: s.minMembers ? Number(s.minMembers) : undefined,
    minClanPoints: s.minClanPoints ? Number(s.minClanPoints) : undefined,
    minClanLevel: s.minClanLevel ? Number(s.minClanLevel) : undefined,
  }),
  component: SearchPage,
})

function SearchPage() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data, isLoading, isError, refetch } = useSearchClans(search)

  const setParam = (key: string, value: string | number | undefined) => {
    navigate({ search: { ...search, [key]: value || undefined } })
  }

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      <h1 className="text-2xl font-bold text-text-primary">部落搜索</h1>

      <Card>
        <Card.Body>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Input label="部落名称" value={search.name ?? ''} onChange={e => setParam('name', e.target.value)} placeholder="搜索部落..." />
            <Select label="战争频率" options={WAR_FREQ_OPTIONS} value={search.warFrequency ?? ''} onChange={e => setParam('warFrequency', e.target.value)} />
            <Input label="最少成员" type="number" value={search.minMembers ?? ''} onChange={e => setParam('minMembers', e.target.value ? Number(e.target.value) : undefined)} />
            <Input label="最少等级" type="number" value={search.minClanLevel ?? ''} onChange={e => setParam('minClanLevel', e.target.value ? Number(e.target.value) : undefined)} />
            <Input label="最少积分" type="number" value={search.minClanPoints ?? ''} onChange={e => setParam('minClanPoints', e.target.value ? Number(e.target.value) : undefined)} />
          </div>
          <div className="flex gap-3 mt-4">
            <Button onClick={() => refetch()}>搜索</Button>
            <Button variant="ghost" onClick={() => navigate({ search: {} })}>重置</Button>
          </div>
        </Card.Body>
      </Card>

      {isError && <ErrorCard onRetry={() => refetch()} />}

      {data && data.items.length > 0 && (
        <Card>
          <Card.Header><span className="text-text-secondary text-sm">共 {data.items.length} 个结果</span></Card.Header>
          <Table
            loading={isLoading}
            emptyText="未找到匹配的部落"
            columns={[
              { key: 'badge', header: '', width: '60px',
                render: (row: ClanOverview) => <img src={row.badgeUrls.small} alt="" className="w-8 h-8" /> },
              { key: 'name', header: '部落', render: (row: ClanOverview) => (
                <div>
                  <div className="font-medium text-text-primary">{row.name}</div>
                  <div className="text-xs text-text-secondary">{row.tag}</div>
                </div>
              )},
              { key: 'clanLevel', header: '等级', align: 'center',
                render: (row: ClanOverview) => <Badge color="primary">Lv.{row.clanLevel}</Badge> },
              { key: 'members', header: '成员', align: 'right',
                render: (row: ClanOverview) => `${row.members}` },
              { key: 'clanPoints', header: '积分', align: 'right',
                render: (row: ClanOverview) => formatNumber(row.clanPoints) },
            ]}
            data={data.items}
          />
        </Card>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Create app/gold-pass.tsx**

```tsx
import { createFileRoute } from '@tanstack/react-router'
import { useCurrentGoldPass } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Skeleton } from '~/components/ui/Skeleton'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { daysRemaining } from '~/lib/utils'

export const Route = createFileRoute('/gold-pass')({
  component: GoldPassPage,
})

function GoldPassPage() {
  const { data, isLoading, isError, refetch } = useCurrentGoldPass()

  return (
    <div className="max-w-page mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold text-text-primary mb-6">Gold Pass 当前赛季</h1>
      {isLoading && <Skeleton.Card />}
      {isError && <ErrorCard onRetry={() => refetch()} />}
      {data && (
        <Card>
          <Card.Body className="text-center py-12 space-y-4">
            <div className="text-5xl">🏆</div>
            <h2 className="text-xl font-semibold text-text-primary">
              {new Date(data.startTime).toLocaleDateString('zh-CN')} ～ {new Date(data.endTime).toLocaleDateString('zh-CN')}
            </h2>
            <div className="text-2xl font-bold text-primary tabular-nums">
              剩余 {daysRemaining(data.endTime)} 天
            </div>
          </Card.Body>
        </Card>
      )}
    </div>
  )
}
```

- [ ] **Step 3: Update routeTree.gen.ts**

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: add search page with filters and gold pass page"
```

---

### Task 11: Admin Pages — War Monitor + CWL

**Files:**
- Create: `warspark-web/app/admin/__layout.tsx`
- Create: `warspark-web/app/admin/war.tsx`
- Create: `warspark-web/app/admin/cwl.tsx`

- [ ] **Step 1: Create app/admin/__layout.tsx**

```tsx
import { Outlet, createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/admin')({
  component: AdminLayout,
})

function AdminLayout() {
  return (
    <div>
      <div className="bg-surface-card border-b border-border">
        <div className="max-w-page mx-auto flex gap-6 px-4 h-12 items-center text-sm">
          <Link to="/admin/war" className="text-text-secondary hover:text-primary [&.active]:text-primary font-medium">战争监控</Link>
          <Link to="/admin/cwl" className="text-text-secondary hover:text-primary [&.active]:text-primary font-medium">CWL 联赛</Link>
        </div>
      </div>
      <Outlet />
    </div>
  )
}
```

- [ ] **Step 2: Create app/admin/war.tsx**

```tsx
import { createFileRoute } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useCurrentWar } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Input } from '~/components/ui/Input'
import { Button } from '~/components/ui/Button'
import { Badge } from '~/components/ui/Badge'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { Skeleton } from '~/components/ui/Skeleton'
import { normalizeTag, warStateLabel } from '~/lib/utils'

export const Route = createFileRoute('/admin/war')({
  component: WarMonitorPage,
})

const STORAGE_KEY = 'warspark_admin_war_tags'

function loadTags(): string[] {
  try { return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]') } catch { return [] }
}
function saveTags(tags: string[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(tags))
}

function WarMonitorPage() {
  const [tags, setTags] = useState<string[]>(loadTags)
  const [input, setInput] = useState('')

  const add = () => {
    const tag = normalizeTag(input)
    if (tag && !tags.includes(tag)) {
      const next = [...tags, tag]
      setTags(next)
      saveTags(next)
    }
    setInput('')
  }

  const remove = (tag: string) => {
    const next = tags.filter(t => t !== tag)
    setTags(next)
    saveTags(next)
  }

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      <div className="flex gap-3">
        <Input value={input} onChange={e => setInput(e.target.value)} placeholder="输入部落标签 #ABC" className="w-64" />
        <Button onClick={add}>添加</Button>
      </div>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {tags.map(tag => (
          <WarCard key={tag} tag={tag} onRemove={() => remove(tag)} />
        ))}
      </div>
    </div>
  )
}

function WarCard({ tag, onRemove }: { tag: string; onRemove: () => void }) {
  const { data, isLoading, isError, refetch } = useCurrentWar(tag)

  return (
    <Card>
      <Card.Header className="flex items-center justify-between">
        <span className="font-medium text-sm text-text-primary">{tag}</span>
        <button onClick={onRemove} className="text-text-secondary hover:text-danger text-xs">移除</button>
      </Card.Header>
      <Card.Body>
        {isLoading && <Skeleton className="h-20 w-full" />}
        {isError && <ErrorCard onRetry={() => refetch()} />}
        {data && (
          <div className="space-y-2 text-sm">
            <Badge color={data.state === 'inWar' ? 'green' : data.state === 'preparation' ? 'yellow' : 'gray'}>
              {warStateLabel(data.state)}
            </Badge>
            {data.state === 'inWar' && (
              <>
                <div className="text-center text-lg font-bold tabular-nums">
                  ⭐ {data.clan.stars} - {data.opponent.stars}
                </div>
                <div className="text-center text-text-secondary tabular-nums">
                  💥 {data.clan.destructionPercentage.toFixed(1)}% - {data.opponent.destructionPercentage.toFixed(1)}%
                </div>
              </>
            )}
          </div>
        )}
      </Card.Body>
    </Card>
  )
}
```

- [ ] **Step 3: Create app/admin/cwl.tsx**

```tsx
import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useCWLGroup } from '~/lib/queries'
import { Card } from '~/components/ui/Card'
import { Input } from '~/components/ui/Input'
import { Button } from '~/components/ui/Button'
import { Badge } from '~/components/ui/Badge'
import { ErrorCard } from '~/components/ui/ErrorCard'
import { Skeleton } from '~/components/ui/Skeleton'
import { normalizeTag } from '~/lib/utils'
import type { CWLClan } from '~/lib/types'

export const Route = createFileRoute('/admin/cwl')({
  component: CWLPage,
})

function CWLPage() {
  const [tag, setTag] = useState(() => {
    try { return JSON.parse(localStorage.getItem('warspark_cwl_tag') ?? '""') } catch { return '' }
  })
  const [input, setInput] = useState(tag)

  const handleSearch = () => {
    const t = normalizeTag(input)
    setTag(t)
    localStorage.setItem('warspark_cwl_tag', JSON.stringify(t))
  }

  const { data, isLoading, isError, refetch } = useCWLGroup(tag)

  return (
    <div className="max-w-page mx-auto py-8 px-4 space-y-6">
      <div className="flex gap-3">
        <Input value={input} onChange={e => setInput(e.target.value)} placeholder="输入部落标签 #ABC" className="w-64" />
        <Button onClick={handleSearch}>查询</Button>
      </div>

      {isLoading && <Skeleton.Card />}
      {isError && <ErrorCard onRetry={() => refetch()} />}

      {data && (
        <>
          <Card>
            <Card.Body>
              <div className="flex items-center gap-3">
                <Badge color={data.state === 'ongoing' ? 'green' : 'gray'}>{data.state === 'ongoing' ? '进行中' : '已结束'}</Badge>
                <span className="text-text-secondary text-sm">赛季: {data.season}</span>
              </div>
            </Card.Body>
          </Card>

          <h3 className="font-semibold text-text-primary">参赛部落 ({data.clans.length})</h3>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {data.clans.map((clan: CWLClan) => (
              <Card key={clan.tag}>
                <Card.Body>
                  <div className="space-y-1">
                    <div className="font-medium text-text-primary">{clan.name}</div>
                    <div className="text-xs text-text-secondary">{clan.tag}</div>
                    <Badge color="primary">Lv.{clan.clanLevel}</Badge>
                    <span className="text-sm text-text-secondary ml-2">{clan.members} 成员</span>
                  </div>
                </Card.Body>
              </Card>
            ))}
          </div>

          <h3 className="font-semibold text-text-primary">轮次</h3>
          <div className="space-y-2">
            {data.rounds.map((round, i) => (
              <Card key={i}>
                <Card.Body className="flex items-center justify-between">
                  <span className="font-medium text-text-primary">第 {i + 1} 轮</span>
                  <div className="flex gap-2">
                    {round.warTags.map(wt => (
                      <code key={wt} className="text-xs bg-surface-page px-2 py-1 rounded">{wt}</code>
                    ))}
                  </div>
                </Card.Body>
              </Card>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
```

- [ ] **Step 4: Update routeTree.gen.ts**

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: add admin pages — war monitor and CWL overview"
```

---

### Task 12: Final Verification & Cleanup

- [ ] **Step 1: Run linter**

```bash
pnpm lint
```

- [ ] **Step 2: Run build**

```bash
pnpm build
```

Verify it produces output in `dist/`.

- [ ] **Step 3: Manual test against running backend**

Start backend:
```bash
cd ../WarSpark && go run ./cmd/warspark server --config=configs/config.dev.yaml
```

Start frontend:
```bash
pnpm dev
```

Test each page:
- `/` — Hero section renders, search works
- `/clans/%232PP` — Clan detail with tabs
- `/players/%232PP` — Player detail
- `/leaderboards` — Rankings with location picker
- `/search` — Search form
- `/gold-pass` — Gold pass info
- `/admin/war` — War monitor
- `/admin/cwl` — CWL data

- [ ] **Step 4: Commit any fixes**

- [ ] **Step 5: Final commit**

```bash
git add -A && git commit -m "feat: WarSpark Web — complete initial release

8 pages: home, clan detail, player detail, leaderboards, search,
gold pass, admin war monitor, admin CWL overview.
React 18 + TanStack Start + Tailwind CSS 4 + TanStack Query."
```

