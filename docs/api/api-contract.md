# WarSpark API 契约草案

## 1. 文档目的

本文档定义 WarSpark MVP 的前后端 API 契约草案，覆盖截图找阵、找阵结果、阵型库、视频关联、战争数据、内容入库和内部后台管理接口。

本文档用于开发前对齐，不代表接口已经实现。后续如果接口字段变化，需要同步更新本文档、Swagger 和相关页面规格。

## 2. 通用约定

### 2.1 API 前缀

业务接口统一使用：

```text
/api/v1
```

健康检查和模板既有接口继续按当前项目约定保留。

### 2.2 响应结构

沿用 GinSpark 模板响应结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

约定：

- `code = 0` 表示业务成功。
- `message` 面向前端或调试，不直接作为最终产品文案。
- `data` 保存业务数据。
- 错误响应仍使用相同结构，`data` 可为空。

### 2.3 分页结构

列表接口统一使用：

```json
{
  "items": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 120
  }
}
```

### 2.4 时间与枚举

- 时间字段使用 ISO 8601 字符串。
- 枚举值使用 snake_case。
- ID 字段使用字符串。
- 百分比字段使用数字，如 `96.5`。

### 2.5 认证

- 前台找阵、阵型库、视频和战争情报读取接口默认匿名可用。
- 匿名上传接口必须做文件校验、限流和滥用保护。
- 内部 admin API 必须接入后台鉴权，不暴露为匿名写接口。
- Clash of Clans 官方 API Key 只保存在服务端，不返回前端。

## 3. 通用枚举

### 3.1 找阵任务状态

```text
created
processing
matched
low_confidence
no_result
failed
```

### 3.2 审核状态

```text
reviewed
pending_review
rejected
needs_update
```

### 3.3 质量状态

```text
high_confidence
medium_confidence
low_confidence
unknown
```

### 3.4 链接类型

```text
official_open_layout
source_page
backup
```

### 3.5 链接状态

```text
active
broken
unverified
missing
```

### 3.6 视频关联分组

```text
attack_video
defense_replay
```

### 3.7 视频关联精度

```text
exact
similar
same_th
```

### 3.8 来源类型

```text
user_upload
public_page
official_api
authorized_import
automated_pipeline
manual_entry
search_result
```

## 4. 错误码建议

| 错误码 | 场景 |
| --- | --- |
| `10001` | 参数错误。 |
| `10002` | 未授权或权限不足。 |
| `10003` | 资源不存在。 |
| `10004` | 资源状态不允许当前操作。 |
| `10006` | 内部错误。 |
| `30001` | 图片格式不支持。 |
| `30002` | 图片文件过大。 |
| `30003` | 图片无法读取。 |
| `30004` | 图片存储失败。 |
| `30005` | 找阵处理失败。 |
| `30006` | 匹配服务不可用。 |
| `30007` | 图片尺寸过小。 |
| `30008` | 匿名上传过于频繁。 |
| `31001` | 阵型链接不可用。 |
| `32001` | YouTube 视频不可访问。 |
| `33001` | 官方 API 未配置。 |
| `33002` | 官方 API 请求失败。 |
| `33003` | 无当前战争。 |
| `33004` | Tag 格式错误。 |

## 5. 截图找阵接口

### 5.1 上传截图并创建任务

```text
POST /api/v1/image-search/jobs
Content-Type: multipart/form-data
```

请求字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `image` | file | 是 | 阵型截图。 |
| `target_position` | int | 否 | 敌方序号。 |
| `target_player_name` | string | 否 | 敌方玩家名。 |
| `expected_th` | int | 否 | 用户或战争上下文预期 TH。 |
| `war_target_id` | string | 否 | 战争目标 ID。 |

上传约束：

- 仅允许 `jpg`、`jpeg`、`png`、`webp`。
- 单张图片最大 10MB。
- 图片最短边不低于 512px。
- 服务端移除 EXIF。
- 服务端生成标准化图片，最长边不超过 4096px。
- 匿名上传按 IP 或等效匿名标识限流。

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "job_id": "job_123",
    "search_status": "created",
    "uploaded_image": {
      "image_id": "img_123",
      "image_url": "https://storage.example/uploads/job_123.png",
      "width": 1440,
      "height": 1440,
      "normalized": true,
      "raw_retention_days": 7
    },
    "created_at": "2026-06-05T12:00:00+08:00"
  }
}
```

### 5.2 查询找阵任务状态

```text
GET /api/v1/image-search/jobs/{job_id}
```

响应字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `job_id` | string | 任务 ID。 |
| `search_status` | string | 任务状态。 |
| `detected_th` | int nullable | 识别 TH。 |
| `screenshot_quality` | string nullable | 截图质量。 |
| `buildings_detected` | int nullable | 可见建筑数量。 |
| `error_code` | string nullable | 失败错误码。 |
| `error_message` | string nullable | 失败说明。 |
| `target_context` | object nullable | 战争目标上下文。 |
| `review_queue_id` | string nullable | 进入后台复核队列时返回。 |

### 5.3 查询找阵结果

```text
GET /api/v1/image-search/jobs/{job_id}/results
```

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "job": {
      "job_id": "job_123",
      "search_status": "matched",
      "detected_th": 16,
      "screenshot_quality": "good",
      "buildings_detected": 92
    },
    "match_summary": {
      "candidate_count": 3,
      "best_match_level": "high",
      "low_confidence": false
    },
    "layouts": [],
    "attack_videos": [],
    "defense_replays": [],
    "fallback": null
  }
}
```

结果页聚合中的 `layouts` 使用阵型卡片结构，`attack_videos` 和 `defense_replays` 使用视频卡片结构。

### 5.4 重试找阵任务

```text
POST /api/v1/image-search/jobs/{job_id}/retry
```

用途：

- 从 `failed` 状态重新排队。
- 从 `no_result` 或 `low_confidence` 场景引导用户重新上传时，也可以直接创建新任务。

约束：

- 不做无限自动重试。
- 如果原图不可用，返回资源状态错误。

## 6. 阵型库接口

### 6.1 查询阵型列表

```text
GET /api/v1/layouts
```

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `th_level` | int | TH 筛选。 |
| `layout_type` | string | 阵型类型。 |
| `style_tag` | string | 风格标签。 |
| `source_type` | string | 来源类型。 |
| `review_status` | string | 审核状态。 |
| `quality_status` | string | 质量状态。 |
| `link_status` | string | 链接状态。 |
| `page` | int | 页码。 |
| `page_size` | int | 每页数量。 |

响应中的阵型卡片：

```json
{
  "layout_id": "layout_123",
  "title": "TH16 War Base",
  "th_level": 16,
  "layout_type": "war",
  "style_tags": ["ring", "box"],
  "primary_image_url": "https://storage.example/layouts/layout_123.png",
  "source_type": "public_page",
  "review_status": "reviewed",
  "quality_status": "high_confidence",
  "link_status": "active",
  "updated_at": "2026-06-05T12:00:00+08:00"
}
```

### 6.2 查询阵型详情

```text
GET /api/v1/layouts/{layout_id}
```

响应数据：

```json
{
  "layout_id": "layout_123",
  "title": "TH16 War Base",
  "th_level": 16,
  "layout_type": "war",
  "style_tags": ["ring", "box"],
  "source_type": "public_page",
  "source_url": "https://example.com/base",
  "review_status": "reviewed",
  "quality_status": "high_confidence",
  "images": [],
  "links": [],
  "attack_videos": [],
  "defense_replays": [],
  "similar_layouts": []
}
```

### 6.3 阵型链接结构

```json
{
  "link_id": "link_123",
  "link_type": "official_open_layout",
  "url": "https://link.clashofclans.com/example",
  "link_status": "active",
  "last_checked_at": "2026-06-05T12:00:00+08:00"
}
```

前端约束：

- `active` 才展示打开和复制主动作。
- `broken` 禁用动作。
- `missing` 展示暂无链接状态。

## 7. 视频接口

### 7.1 查询阵型相关视频

```text
GET /api/v1/layouts/{layout_id}/videos
```

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `match_group` | string | `attack_video` 或 `defense_replay`。 |
| `match_type` | string | `exact`、`similar`、`same_th`。 |

视频卡片结构：

```json
{
  "match_id": "match_123",
  "video_id": "video_123",
  "youtube_video_id": "abc123",
  "video_title": "TH16 Attack Replay",
  "channel_name": "Example Channel",
  "timestamp_seconds": 245,
  "youtube_url": "https://www.youtube.com/watch?v=abc123&t=245s",
  "match_group": "attack_video",
  "match_type": "exact",
  "stars": 3,
  "destruction_percent": 100,
  "video_tags": ["ground"],
  "confidence_score": 0.92,
  "review_status": "reviewed"
}
```

### 7.2 视频失效反馈

```text
POST /api/v1/videos/{video_id}/report-broken
```

请求：

```json
{
  "reason": "video_unavailable",
  "source_page": "layout_detail"
}
```

说明：

- MVP 可先记录反馈，不必立即自动隐藏视频。
- Worker 或后台后续更新视频状态。

## 8. 战争数据接口

### 8.1 查询部落概览

```text
GET /api/v1/clans/{clan_tag}
```

用途：

- 查询部落名称、等级、成员数、标签和基础状态。
- 当前战争不可访问时仍可展示基础信息。

### 8.2 查询玩家概览

```text
GET /api/v1/players/{player_tag}
```

用途：

- 查询玩家名称、TH、所属部落和基础信息。
- 支持从玩家 tag 反查可能的战争上下文。

### 8.3 查询当前战争

```text
GET /api/v1/war/current?clan_tag={clan_tag}
```

响应数据：

```json
{
  "war_snapshot_id": "war_123",
  "clan_tag": "#AAA111",
  "opponent_clan_tag": "#BBB222",
  "war_state": "in_war",
  "team_size": 15,
  "clan_stars": 20,
  "opponent_stars": 18,
  "clan_destruction": 88.5,
  "opponent_destruction": 84.2,
  "fetched_at": "2026-06-05T12:00:00+08:00"
}
```

错误状态：

- API 未配置。
- Tag 格式错误。
- 无当前战争。
- 官方 API 请求失败。

### 8.4 查询 CWL 基础结构

```text
GET /api/v1/war/cwl?clan_tag={clan_tag}&season=2026-06
```

用途：

- 获取 CWL 分组、轮次和参与成员的基础结构。
- MVP 只做基础展示，不做完整历史分析。

### 8.5 查询战争成员

```text
GET /api/v1/war/snapshots/{war_snapshot_id}/members
```

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `side` | string | `clan` 或 `opponent`。 |

成员结构：

```json
{
  "member_id": "member_123",
  "side": "opponent",
  "map_position": 5,
  "player_tag": "#PLAYER",
  "player_name": "Player",
  "th_level": 16,
  "attacks_used": 1,
  "best_stars_against": 2,
  "best_destruction_against": 89.5
}
```

### 8.6 创建目标找阵任务

```text
POST /api/v1/war/targets/{target_id}/image-search-jobs
Content-Type: multipart/form-data
```

请求字段：

- `image`

说明：

- 该接口等价于上传截图并携带目标上下文。
- 如果战争目标不存在，返回资源不存在。
- 如果官方 API 不可用，但已有快照和目标存在，仍可创建找阵任务。

## 9. 内部 Admin API

以下接口属于 MVP 内部后台能力，必须鉴权，不进入匿名前台写接口。

### 9.1 查询后台复核队列

```text
GET /api/v1/admin/review-queue
```

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `resource_type` | string | `layout`、`image_search_job`、`layout_link`、`video_match`。 |
| `review_status` | string | 审核状态。 |
| `quality_status` | string | 质量状态。 |
| `page` | int | 页码。 |
| `page_size` | int | 每页数量。 |

### 9.2 创建阵型草稿

```text
POST /api/v1/admin/layouts
```

请求：

```json
{
  "title": "TH16 War Base",
  "th_level": 16,
  "layout_type": "war",
  "style_tags": ["ring"],
  "source_type": "manual_entry",
  "source_url": "https://example.com/source",
  "review_status": "pending_review",
  "quality_status": "medium_confidence",
  "visibility": "hidden"
}
```

### 9.3 添加阵型链接

```text
POST /api/v1/admin/layouts/{layout_id}/links
```

请求：

```json
{
  "link_type": "official_open_layout",
  "url": "https://link.clashofclans.com/example",
  "source_type": "manual_entry",
  "source_url": "https://example.com/source"
}
```

### 9.4 添加视频关联

```text
POST /api/v1/admin/layouts/{layout_id}/video-matches
```

请求：

```json
{
  "youtube_video_id": "abc123",
  "video_title": "TH16 Attack Replay",
  "channel_name": "Example Channel",
  "timestamp_seconds": 245,
  "match_group": "attack_video",
  "match_type": "similar",
  "stars": 3,
  "destruction_percent": 100,
  "confidence_score": 0.8,
  "source_type": "manual_entry",
  "source_url": "https://example.com/source"
}
```

### 9.5 更新审核状态

```text
PATCH /api/v1/admin/review/{resource_type}/{resource_id}
```

请求：

```json
{
  "review_status": "reviewed",
  "quality_status": "high_confidence",
  "visibility": "public",
  "note": "source checked"
}
```

### 9.6 更新链接状态

```text
PATCH /api/v1/admin/layout-links/{link_id}
```

请求：

```json
{
  "link_status": "broken",
  "last_checked_at": "2026-06-05T12:00:00+08:00",
  "note": "cannot open"
}
```

### 9.7 查询后台操作日志

```text
GET /api/v1/admin/audit-logs
```

用途：

- 追踪阵型、链接、视频关联和复核状态的后台修改记录。
- 支撑后续问题回溯。

## 10. 前端状态映射

| 接口状态 | 页面状态 |
| --- | --- |
| 找阵任务 `created` / `processing` | 处理中。 |
| 找阵任务 `matched` | 展示完整结果。 |
| 找阵任务 `low_confidence` | 展示低置信度提示和候选。 |
| 找阵任务 `no_result` | 展示无结果状态和替代入口。 |
| 找阵任务 `failed` | 展示失败和重新上传入口。 |
| 链接 `broken` | 禁用打开和复制。 |
| 视频不可访问 | 展示视频失效状态。 |
| 战争 API 不可用 | 战争数据不可用，找阵入口仍可使用。 |

## 11. API 验收标准

API 契约视为可进入实现时，需要满足：

- 截图上传、任务状态、结果查询接口完整。
- 匿名上传限制、EXIF 清理、尺寸约束和限流要求明确。
- 找阵结果能聚合阵型、进攻视频、防守回放和阵型链接状态。
- 阵型列表和详情接口覆盖页面规格字段。
- 视频接口明确分组和关联精度。
- 战争数据接口覆盖部落、玩家、当前战争、CWL 基础结构和目标绑定。
- 内部 admin API 明确需要鉴权，覆盖复核队列、阵型、链接、视频关联和操作日志。
- 响应结构与 GinSpark 模板保持一致。
