# 05 · API 接口契约

## 1. 通用约定

| 项 | 约定 |
|---|---|
| 版本 | 路径前缀 `/api/v1`，版本只增不改；破坏性变更开 `v2` 并保留 v1 至少 6 个月 |
| 传输 | 仅 HTTPS（自部署可 HTTP，但文档明确警告） |
| 编码 | 请求与响应均 `application/json; charset=utf-8`，流式为 `text/event-stream` |
| 时间 | 一律 RFC3339 带偏移，如 `2026-11-25T14:30:00+08:00` |
| 字段命名 | 响应体 `snake_case`（与数据库、Go struct tag 一致），避免前端做转换 |
| 幂等 | 写操作支持 `Idempotency-Key` 头，24 小时内同键返回首次结果 |
| 追踪 | 响应头 `X-Request-Id` 必回，客户端可透传 |

### 1.1 统一响应信封

非流式接口一律返回：

| 字段 | 类型 | 说明 |
|---|---|---|
| code | int | 0 为成功，非 0 为业务错误码 |
| message | string | 面向用户的中文提示，可直接弹窗 |
| data | object | 业务数据，错误时为 null |
| request_id | string | 追踪 ID |
| server_time | string | RFC3339，用于校正客户端时钟 |

HTTP 状态码同时使用：400 参数错误、401 未认证、403 无权限、404 不存在、409 冲突、429 限流、500 服务端错误、502 上游模型错误。**业务码比 HTTP 码更细**，前端优先读 `code`。

### 1.2 错误码表

| 码 | HTTP | 含义 | 前端处理 |
|---|---|---|---|
| 0 | 200 | 成功 | — |
| 10001 | 400 | 参数校验失败，`data.fields` 给出字段级原因 | 表单标红 |
| 10002 | 400 | 时间超出支持范围（1900–2100） | 提示文案 |
| 10003 | 400 | 出生信息不完整或日期不存在（如农历闰月错配） | 提示文案 |
| 10101 | 401 | 未认证或令牌过期 | 刷新令牌，失败则跳登录 |
| 10102 | 401 | 设备令牌无效 | 重新生成游客令牌 |
| 10201 | 403 | 额度不足 | 引导升级/次日再来 |
| 10202 | 403 | 功能未开放（如分享图未启用） | 置灰 |
| 10301 | 404 | 盘面不存在或已删除 | 回列表 |
| 10401 | 409 | 并发写冲突 | 提示重试 |
| 10501 | 429 | 触发限流，`data.retry_after` 秒 | 倒计时 |
| 20001 | 500 | 排盘引擎内部错误 | 报错并提示已记录 |
| 20002 | 500 | 数据库不可用 | 提示稍后 |
| 30001 | 502 | 模型服务不可用（已尝试全部降级） | 显示"解读暂不可用"，盘面仍可看 |
| 30002 | 504 | 模型超时 | 同上 |
| 30003 | 502 | 模型返回内容不合规（触发输出校验） | 提示重试，记录日志 |
| 40001 | 403 | 管理员权限不足 | — |

## 2. 认证

| 方式 | 头 | 有效期 | 说明 |
|---|---|---|---|
| 用户 | `Authorization: Bearer {access_token}` | 15 分钟 | JWT，claims 含 sub、role、plan |
| 刷新 | `POST /auth/refresh`，refresh token 走 httpOnly Cookie | 7 天 | 轮换式：每次刷新签发新 refresh 并作废旧 token |
| 游客 | `X-Device-Token: {token}` | 长期（前端持久化） | 服务端只存哈希，额度按设备计 |

游客升级：注册或登录时若带 `X-Device-Token`，服务端把该设备的盘面迁移到新用户账号（`charts.guest_device_id` 置空、`user_id` 置新值），迁移上限 20 张。

## 3. 端点清单

### 3.1 元信息

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/meta/calendar?at=&tz=&lon=` | 返回干支、节气前后、月将、旬空，用于首页"今日盘头" |
| GET | `/meta/capabilities` | 三术能力描述与表单字段，前端据此动态渲染输入区 |
| GET | `/meta/health` | 轻量健康检查（DB/Redis/最近一次模型调用成功率） |

### 3.2 认证与账号

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/auth/guest` | 生成游客身份，返回 device_token（仅首次） |
| POST | `/auth/register` | email + password + nickname |
| POST | `/auth/login` | 返回 access + 设置 refresh Cookie |
| POST | `/auth/refresh` | 轮换令牌 |
| POST | `/auth/logout` | 作废 refresh |
| GET | `/auth/me` | 当前用户信息与额度摘要 |
| PATCH | `/auth/me` | 改昵称、改密码、清除已存出生信息 |

### 3.3 排盘

统一入口 `POST /charts/{method}/cast`，`method ∈ {liuyao, daliuren, ziwei}`。这样新增术数零路由改动。

**六爻请求字段**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| mode | enum | 是 | coin / time / manual |
| question | string | 否 | ≤ 200 字 |
| topic | enum | 否 | 显式问事类别：`exam` / `wealth` / `career` / `relationship` / `health` / `travel` / `lost` / `legal` / `family` / `surname` / `general`；为空时按问题关键词分类 |
| gender | enum | 否 | `male` / `female`；感情类用于定取妻财或官鬼，未填则并列两种取法并标 `needs_gender` |
| at | string | 否 | RFC3339；缺省为服务端当前时刻（按 tz 解释）。**装卦需要月建日辰，所以 manual 也必须有时间来源。** |
| tz | string | 否 | 默认 `Asia/Shanghai` |
| lon | float | 否 | 东经度数，默认 120；启用真太阳时必须有值 |
| use_true_solar_time | bool | 否 | 默认 true |
| day_boundary | enum | 否 | `zi_chu` / `zi_zheng`，默认 `zi_chu` |
| coins | int[6][3] | 否 | 仅 coin 模式。每个元素 0（字）/ 1（背），服务端按行求和得背数。 |
| backs | int[6] | 否 | 仅 coin 模式。直接给六爻背数 0–3，与 coins 二选一。 |
| lines | enum[6] | manual 时必填 | `old_yang` / `young_yin` / `young_yang` / `old_yin`，也接受中文 `老阳` / `少阴` / `少阳` / `老阴` |

`mode=coin` 且 `coins`、`backs` 都不传时，服务端用 `crypto/rand` 摇卦，并在响应里返回 `entropy`（`entropy_id` / `backs` / `entropy_proof` / `source` / `created_at`），供用户复核"这一卦不是后台编的"。客户端自带铜钱时不产生 `entropy`，因为此时随机性不在服务端。

`mode=time`（梅花时间起卦）v0.1 尚未实现，返回业务码 10202，不要当作参数错误。

**大六壬请求字段**：`at`、`tz`、`lon`、`question`、`school`（流派开关）。

**紫微请求字段**：`gender`、`calendar_type`、`birth_date`、`birth_time`、`is_leap_month`、`leap_month_policy`、`tz`、`lon`、`question`。

**响应**：`chart_id`、`method`、`engine_version`、`question`、`calendar`、`school`、`payload`、`facts`（服务端默认只回 `weight=core` 的 facts，完整清单 `GET /charts/{id}` 可取）、`warnings`、`entropy`（如为服务端摇卦）、`created_at`。六爻 `payload.use_shen` 返回规则引擎的取用神结果，含 `topic`、`topic_label`、`rule_id`、`confidence`、`primary`、`secondary`、`taboo`、`rationale`、`candidates`；测姓类额外含 `surname_hints`。`facts` 中同步给出「问事类别」「用神」两条 core 事实，测姓类再加「姓氏候选」。

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/charts?method=&cursor=&limit=` | 游标分页，`limit` 默认 20，上限 50 |
| GET | `/charts/{id}` | 完整盘面（本人或公开分享可见） |
| PATCH | `/charts/{id}` | 仅允许改 `title` |
| DELETE | `/charts/{id}` | 软删除 |
| GET | `/charts/{id}/export?format=json` | 导出原始盘面，供第三方复核（P1 加 png） |

### 3.4 解读（SSE）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/charts/{id}/interpret` | 首轮解读，流式 |
| POST | `/charts/{id}/messages` | 多轮追问，流式 |
| GET | `/charts/{id}/messages?conversation_id=` | 拉取历史消息 |
| POST | `/messages/{id}/feedback` | 1–5 星 + 原因 |

请求体（首轮）：`question`、`style`（`plain` 白话 / `classic` 文言夹杂）、`tier_override`（仅管理员可指定）、`max_tokens`（可选，受服务端上限约束）。

请求体（追问，当前实现）：

```json
{
  "message": "应期为什么以出空为节点？",
  "history": [{"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}],
  "style": "plain",
  "chart": { "...": "可选；服务端内存无此盘时，用本地盘面快照重建 facts" }
}
```

- `message` 必填，≤200 字；`history` 只带最近对话，服务端限最近 12 条、每条截断到 2500 字。
- 当前实现只落地 `POST /charts/{id}/messages`；`GET /charts/{id}/messages`、反馈与重连缓冲仍在路线图。
- 追问仍走同一套 SSE `meta / delta / usage / done / error`，但输出不要求四段标题；服务端用 `AuditFollowup` 做事实硬校验，超过 6 轮返回 400。

**SSE 事件契约**

| event | data | 说明 |
|---|---|---|
| `meta` | `{message_id, model, provider}` | 流开始，前端建立占位消息 |
| `delta` | `{text}` | 增量文本，前端追加 |
| `citation` | `{anchor, snippet}` | 一条结论锚点，用于盘面高亮 |
| `usage` | `{prompt_tokens, completion_tokens}` | 结束时给出用量 |
| `done` | `{message_id, finish_reason}` | 正常结束 |
| `error` | `{code, message}` | 出错，连接随后关闭 |
| `ping` | `{}` | 每 15s 一次，防中间设备断连 |

客户端必须处理：网络中断后用 `Last-Event-ID` 重连，服务端从 `sse:{message_id}` 缓冲补发；重连超过 3 次则放弃并提示。

### 3.5 额度与用量

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/usage/quota` | 今日剩余、套餐余额、重置时间 |
| GET | `/usage/records?from=&to=` | 自己的用量明细（不含金额，管理员视图另开） |

### 3.6 分享与计费

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/charts/{id}/share` | 生成分享码与图片，返回短链 |
| GET | `/s/{code}` | 公开访问（无需登录，只读，不含对话） |
| GET | `/plans` | 套餐列表 |
| POST | `/orders` | 下单，需 `Idempotency-Key` |
| POST | `/orders/{id}/notify/{channel}` | 渠道回调，验签，不走 JWT |
| GET | `/subscriptions/me` | 当前订阅 |

### 3.7 管理后台

| 方法 | 路径 | 说明 |
|---|---|---|
| GET / POST / PATCH | `/admin/providers` | provider 增改、启停、权重 |
| GET / POST / PATCH | `/admin/prompts` | 提示词版本管理 |
| GET | `/admin/stats?from=&to=` | 出盘量、解读量、成本、失败率 |
| POST | `/admin/users/{id}/ban` | 封禁 |

管理接口统一要求 `role=admin`，并写 `audit_logs`。

## 4. 限流

| 场景 | 游客 | 免费用户 | Pro |
|---|---|---|---|
| 排盘 | 20 次/日、5 次/分 | 50 次/日、10 次/分 | 不限次数、20 次/分 |
| 解读 | 3 次/日、1 次/分 | 10 次/日、2 次/分 | 50 次/日、3 次/分 |
| 追问 | 0（需注册） | 随会话 10 轮 | 随会话 30 轮 |
| 登录 | 5 次/分/IP | — | — |
| 注册 | 3 次/小时/IP | — | — |

实现：Redis + Lua 原子令牌桶，键 `rl:{subject}:{route}`。超限返回 10501 并给 `retry_after`。

## 5. 前端契约要点

- 所有 GET 列表接口用游标分页（`next_cursor`），不用 offset——盘面表会很大。
- `facts` 的 `anchor` 是 JSON Pointer（如 `/payload/lines/2/najia_zhi`），前端据此定位高亮，不猜结构。
- SSE 用原生 `EventSource` 不够（要 POST + 自定义头），前端用 `fetch` + `ReadableStream` 手工解析，封装成 `useSSE()` composable。
- 错误统一拦截：HTTP 401 → 尝试一次 refresh → 仍失败则清态跳登录；10501 → 弹倒计时。

## 6. OpenAPI 与 SDK

- Gin 侧用 `swaggo` 或手写 `openapi.yaml`。推荐**手写 yaml 作为契约源**，因为它要独立于代码被前端与第三方消费；CI 校验 yaml 与 handler 结构体一致（用 `kin-openapi` 做 schema 校验）。
- 生成 TypeScript 类型：`openapi-typescript` 输出 `apps/web/src/api/types.ts`，前后端字段永不漂移。
- 对外 SDK 暂不做官方维护，文档提供 curl 与 fetch 示例即可。
