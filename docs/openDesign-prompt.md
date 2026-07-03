# WarSpark — openDesign 设计提示词

## 项目概述

WarSpark 是一个 Clash of Clans 数据查询平台，包含 Go REST API 后端 + React TypeScript 前端。前端技术栈：React 18 + TanStack Router + TanStack Query v5 + Tailwind CSS 4 + Vite 6。

设计系统当前令牌：
- 主色 `#3A4DF0` (primary)
- 页面背景 `#F7F9FD` (surface-page)
- 卡片背景 `#FFFFFF` (surface-card)
- 更多见 `frontend/styles/globals.css` 中 `@theme` 定义

---

## 当前已实现页面

### 1. 首页 `/`
- **Hero 区域**：渐变蓝紫背景 + WarSpark 标题 + 搜索栏（部落/玩家单选 + 标签输入框 + 搜索按钮）
- **热门排行榜**：三列卡片网格（TOP 部落 / TOP 玩家 / TOP 都城），各显示前 10 名
- 地区选择器切换数据范围

### 2. 部落详情 `/clans/$tag`
- Banner：大尺寸部落徽章、名称、Tag、等级 Badge、战争联赛、类型/状态
- 统计区块：成员数、积分、战绩（胜/连胜）、战争频率
- 三个 Tab：
  - **成员概览**：表格（排名、联赛图标、成员名+Tag、角色(中文)、等级、TH、奖杯、捐/收兵）
  - **战争日志**：分页表格（日期、对手、规模、结果(绿/红/灰)、星数、摧毁率）
  - **突袭赛季**：卡片网格（状态、起止时间、突袭完成数、总攻击数、总战利品）

### 3. 玩家详情 `/players/$tag`
- Banner：联赛段位图标、玩家名、Tag、Badge 行（TH等级+武器⭐、BH等级、角色、联赛、参战状态）
- 所属部落卡片
- 五个 Tab：
  - **概览**：家园战斗统计（6列）、部落贡献（3列）、夜世界统计、传说联赛统计
  - **军队**：二级 Tab（英雄/装备/兵种/法术），卡片网格 + 家园/夜世界分区，超兵 Badge，装备子项
  - **成就**：成就列表（星级、进度、村庄标识）
  - **战斗日志**：表格（时间、进攻/防守、星数、摧毁率、对手）
  - **联赛分组**：表格（玩家、部落、奖杯、进攻/防守胜负）

### 4. 排行榜 `/leaderboards`
- 地区选择器
- 5 个排名类型 Tab：部落 / 玩家 / 都城 / 夜世界·部落 / 夜世界·玩家
- 表格分页（排名变化值 + 可点击跳转详情）

### 5. 搜索 `/search`
- 部落/玩家搜索切换按钮
- 部落模式：名称、战争频率(下拉5选)、最少成员、最少等级、最少积分筛选
- 玩家模式：标签输入 → 直接跳转玩家页
- 搜索结果表格

### 6. Gold Pass `/gold-pass`
- 当前赛季起止时间 + 剩余天数卡片

### 7. Admin `/admin`（仅占位壳）
- 子导航：战争监控 / CWL 联赛
- 战争监控 `/admin/war`：仅显示标题（规划：多标签输入、实时轮询、状态展示）
- CWL 联赛 `/admin/cwl`：仅显示标题（规划：部落输入、赛季信息、参赛部落、轮次详情）

---

## 已完成的后端 API 端点（已对接前端）

```
GET  /clans/:tag                    → 部落详情 + 成员列表
GET  /clans/:tag/war-log            → 战争日志（游标分页）
GET  /clans/:tag/capital-raid-seasons → 突袭赛季
GET  /players/:tag                  → 玩家概览（含英雄/兵种/法术/成就/传说联赛）
GET  /players/:tag/battle-log       → 战斗日志
GET  /players/:tag/league-group     → 联赛分组
GET  /locations                     → 地区列表
GET  /locations/:id/rankings/clans  → 部落排名
GET  /locations/:id/rankings/players → 玩家排名
GET  /locations/:id/rankings/clans-capital → 都城排名
GET  /locations/:id/rankings/clans-builder-base → 夜世界部落排名
GET  /locations/:id/rankings/players-builder-base → 夜世界玩家排名
GET  /clans?name=&warFrequency=&minMembers=... → 部落搜索
GET  /war/current?clan_tag=          → 当前战争
GET  /war/cwl?clan_tag=              → CWL 联赛组
GET  /war/cwl/wars/:warTag          → CWL 单场战争
GET  /gold-pass/current              → 当前 Gold Pass
GET  /leagues                        → 联赛列表
```

---

## 规划中 / 未实现的功能

### 高优先级

1. **Admin 战争监控** — 多部落标签输入 + localStorage 持久化 + 30s 自动轮询 + 战争状态展示（未参战/准备日/战斗中/已结束）+ 部落 vs 对手详情
2. **Admin CWL 联赛** — 部落标签输入 + 联赛赛季信息 + 参赛部落列表 + 7 轮对战详情 + 点选查看每场战争
3. **导航栏 SearchBar** — 全局搜索框集成到 `__root.tsx` 导航栏（组件已有，未集成）
4. **移动端响应式** — Hamburger 菜单、Tab 横滚、表格横向滚动
5. **联赛元数据浏览** — `/leagues` 列表页 + `/leagues/:id` 详情页 + `/leagues/:id/seasons` 赛季列表 + 赛季排名
6. **标签管理页** — `/labels/clans` + `/labels/players`，按标签筛选部落/玩家

### 中优先级

7. **找阵功能**（图片搜索）：
   - 上传截图页
   - 结果页（相似阵型列表 + 视频 + 防守回放）
   - 阵型库列表页 / 详情页
8. **搜索页 URL 参数同步** — 搜索条件写入 URL search params，支持分享链接
9. **社交分享** — Open Graph 标签订阅、分享卡片预览

### 低优先级 / V2

10. **用户系统** — 注册/登录 + JWT 认证 + 收藏夹 + 个人设置
11. **战争情报页**（C 端）— 面向普通用户的战争分析页面
12. **复盘页** — 战争回放分析
13. **数据对比** — 部落/玩家横向对比
14. **暗色模式**
15. **国际化**

---

## 设计风格参考

当前已建立的设计语言：
- 主色 `#3A4DF0`（蓝色系），hero 区域用渐变 `from-primary via-primary to-primary-light`
- 磨砂玻璃导航栏 `backdrop-blur-lg bg-white/80`
- 卡片白底圆角 `rounded-card`（16px），hover 上浮效果 `card-hover`
- 表格斑马纹行
- 排名前三渐变徽章（金/银/铜）
- Badge 实色背景（green/red/yellow/blue/gray/primary/purple/orange）
- 数字使用 `toLocaleString()` 原始值，不缩写

请在设计中保持此风格一致，如有更好的视觉方案可提出对比。所有新增页面应为移动端优先的响应式设计。
