# 05 · API 接口契约

> 本契约描述 v1 的目标形态。**当前仓库已实现的是无账号的单机版**：六术排盘、盘面读取/列表、历法、能力清单、首轮 SSE 解读与同盘追问；盘面存内存。`/auth/*`、额度、分享、支付、管理后台、Postgres/Redis、`Idempotency-Key` 与 SSE 重连缓冲仍属路线图，调用会 404。下文标注“当前实现”的段落与代码一致，其余为设计目标。

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
| 30001 | 502 | 模型服务不可用、鉴权失败或超时 | 显示"解读暂不可用"，盘面仍可看 |
| 30002 | 502 | 解读未通过盘面事实校验，已整段拦截 | 提示重试，记录日志 |
| 30003 | 503 | 未配置模型，追问不可用（首轮离线摘要仍可查看） | 引导配置模型或另起新盘 |
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
| GET | `/meta/capabilities` | 六术能力描述与表单字段，前端据此动态渲染输入区 |
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

统一入口 `POST /charts/{method}/cast`，`method ∈ {liuyao, daliuren, ziwei, xiaoliuren, meihua, qimen}`。这样新增术数零路由改动。

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

`mode=time`（六爻按时间起卦）v0.1 尚未实现，当前会按参数错误返回 10001；梅花的时间起卦走 `/charts/meihua/cast` 的 `mode=time`。

**大六壬请求字段**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| question | string | 解读时必填 | ≤ 200 字。起课接口允许为空以便先存盘，但首轮解读会返回 10001 “请先填写所问之事” |
| at | string | 否 | RFC3339 带偏移；缺省为服务端当前时刻（按 tz 解释）。月将加时依赖它 |
| tz | string | 否 | 默认 `Asia/Shanghai` |
| lon | float | 否 | 东经度数，默认 120；启用真太阳时必须有值 |
| use_true_solar_time | bool | 否 | 默认 true |
| day_boundary | enum | 否 | `zi_chu` / `zi_zheng`，默认 `zi_chu` |
| school | object | 否 | 流派开关，当前识别 `sanchuan`、`shehai`、`fuyin_zhongmo`、`fanyin_chuchuan`、`guiren`；盘面原样记录 |

**大六壬响应 payload**

| 字段 | 类型 | 说明 |
|---|---|---|
| tian_pan | object[12] | 十二地盘宫。每项含 `earth`（地盘支）、`heaven`（天盘支）、`heaven_wuxing`、`liu_qin`、`tian_jiang`、`shen_sha[]` |
| si_ke | object[4] | 四课：`index`、`name`、`upper`（上神）、`lower`（下神）、`lower_wuxing`、`upper_wuxing`、`relation`（生克关系） |
| san_chuan | object[3] | 初/中/末传：`index`、`name`、`branch`、`palace`、`tian_jiang`、`liu_qin` |
| ke_ti | string | 课体，如 `贼克`、`重审` |
| gate | string | 九宗门判门结果，如 `贼克` |
| month_general | object | 月将：`zhi`、`name`、`after_term`（节气后说明） |
| hour_zhi | string | 占时地支 |
| day_gan / day_zhi | string | 日干、日支；`day_gan` 序列化为中文天干，不是数字 |
| xun_kong | string[2] | 旬空两支 |
| shen_sha | object[] | 全局神煞：`name`、`zhi`、`value`（如驿马在亥） |

**紫微请求字段**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| gender | enum | 是 | `male` / `female`；也接受 `男` / `女` / `乾` / `坤` 并归一化 |
| calendar_type | enum | 否 | `solar`（默认）/ `lunar` |
| birth_date | string | 是 | `YYYY-MM-DD`；农历输入时同样按该格式填农历年月日 |
| birth_time | string | 是 | `HH:MM` 或 `HH:MM:SS`；23:00–23:59 晚子时归次日 |
| is_leap_month | bool | 否 | 仅农历输入有意义；表示 `birth_date` 的月份是闰月 |
| leap_month_policy | enum | 否 | `as_this_month`（默认）/ `as_next_month`，仅闰月时参与命宫月份计算 |
| question | string | 否 | ≤ 200 字。为空时首轮解读自动补“看整体格局”的默认问题 |
| tz | string | 否 | 默认 `Asia/Shanghai` |
| lon | float | 否 | 东经度数，默认 120 |
| use_true_solar_time | bool | 否 | 默认 true。紫微在 `ParseBirth` 内先归一化，构建历法上下文时不会二次校正 |
| day_boundary | enum | 否 | `zi_chu` / `zi_zheng`，默认 `zi_chu` |
| school | object | 否 | 当前记录 `sihua=standard` 与 `geng=tiantong_ke_taiyin_ji` |

**紫微响应 payload**

| 字段 | 类型 | 说明 |
|---|---|---|
| palaces | object[12] | 十二宫：`branch`、`stem`、`stem_branch`、`name`、`major_stars[]`、`minor_stars[]`、`sihua[]`、`daxian_range` |
| ming_gong / shen_gong | string | 命宫、身宫地支 |
| wu_xing_ju | string | 五行局，如 `火六局` |
| ziwei_pos / tianfu_pos | string | 紫微、天府所在宫 |
| sihua_year | object[] | 生年四化：`gan`、`star`、`hua` |
| da_xian | object[12] | 大限：`index`、`name`、`branch`、`start_age`、`end_age`、`range` |
| lunar_date | string | 如 `庚午年四月廿一` |
| solar_date | string | 归一化后的公历时刻，已含真太阳时/晚子时结果 |
| year_gz | string | 生年干支 |

**小六壬请求字段**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| question | string | 解读时必填 | ≤ 200 字；用于近事解读 |
| at | string | 否 | RFC3339；缺省为服务端当前时刻 |
| tz | string | 否 | 默认 `Asia/Shanghai` |
| lon | float | 否 | 东经度数，默认 120 |
| use_true_solar_time | bool | 否 | 默认 true |
| day_boundary | enum | 否 | `zi_chu` / `zi_zheng`，默认 `zi_chu` |
| school | object | 否 | 原样记录；默认 `month_start=daan`、`hour_branch_base=zi=1` |

**小六壬响应 payload**

| 字段 | 类型 | 说明 |
|---|---|---|
| lunar_month / lunar_day | int | 农历月、日 |
| hour_zhi | string | 时支 |
| month_palace / day_palace / final_palace | object | 月宫 / 日宫 / 落宫，含 `index`、`name`、`element`、`deity`、`direction`、`luck`、`meaning` |
| six_palaces | object[6] | 六宫属性完整列表 |
| school | object | 流派开关 |

**梅花请求字段**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| mode | enum | 否 | `time`（默认）/ `numbers` |
| numbers | int[] | numbers 时必填 | 2 或 3 个 1—999999 的正整数；两组取和定动爻，三组第三数定动爻 |
| question | string | 是 | ≤ 200 字 |
| at / tz / lon / use_true_solar_time / day_boundary | — | 否 | 同小六壬；时间卦使用 |

**梅花响应 payload**

| 字段 | 类型 | 说明 |
|---|---|---|
| mode | string | time / numbers |
| numbers | int[] | 报数，时间卦为空 |
| year_zhi / lunar_date / hour_zhi | string | 时间卦的年支、农历日期、时支 |
| upper_number / lower_number / moving_number | int | 上卦数、下卦数、动爻数 |
| ben_gua / hu_gua / bian_gua | object | 本卦 / 互卦 / 变卦，含 `name`、`upper`、`lower`、`lines[6]` |
| moving_line | int | 动爻 1—6 |
| ti_gua / yong_gua | object | 体卦 / 用卦，含 `name`、`element` |
| relation / changed_relation | string | 体用关系 / 变卦关系 |

**奇门请求字段**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| question | string | 解读时必填 | ≤ 200 字 |
| at | string | 否 | RFC3339；缺省为服务端当前时刻，按节气与符头起局 |
| tz / lon / use_true_solar_time / day_boundary | — | 否 | 同上 |
| school | object | 否 | 原样记录；默认 `pan=zhuanpan`、`ju=chaibu_futou`、`qin=ji_kun`、`geju=not_included` |

**奇门响应 payload**

| 字段 | 类型 | 说明 |
|---|---|---|
| dun / ju / yuan / solar_term | string / int | 阴遁或阳遁、局数、上中下元、节气 |
| term_at | string | 节气时刻 |
| year_gz / month_gz / day_gz / hour_gz | string | 四柱 |
| xun_shou / instrument | string | 旬首与旬首六仪 |
| zhi_fu / zhi_fu_palace | string / int | 值符星与落宫 |
| zhi_shi / zhi_shi_palace | string / int | 值使门与落宫 |
| kong / horse / horse_palace | string[2] / string / int | 时空亡、马星与落宫 |
| palaces | object[9] | 九宫：`earth_gan`、`heaven_gan`、`star`、`door`、`god`、`men_po`、`kong`、`horse`、`is_zhi_fu`、`is_zhi_shi` |

**响应**：`chart_id`、`method`、`engine_version`、`question`、`calendar`、`school`、`payload`、`facts`（服务端默认只回 `weight=core` 的 facts，完整清单 `GET /charts/{id}` 可取）、`warnings`、`entropy`（仅六爻服务端摇卦）、`created_at`。六爻 `payload.use_shen` 返回规则引擎的取用神结果，含 `topic`、`topic_label`、`rule_id`、`confidence`、`primary`、`secondary`、`taboo`、`rationale`、`candidates`；测姓类额外含 `surname_hints`。`facts` 中同步给出「问事类别」「用神」两条 core 事实，测姓类再加「姓氏候选」。

> 当前 `chartView` 不单独返回 `input`；归一化后的出生信息在 `payload`（紫微）或 `calendar` 中可见。`calendar` 对紫微记录的是归一化后的公历时刻与节气，不是原始输入。

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

- `question` 为空时先回退到盘面创建时的问题；紫微仍为空则补默认问题“请以这张命盘，看整体格局、当下值得注意的方向。”；大六壬为空直接 10001。
- 服务端按 `chart.method` 选择提示词与审计规则：六爻查卦名/爻位/六亲/世应/动爻；大六壬查课体/三传/天将越界；紫微查星曜/四化越界；小六壬查六宫名称；梅花查 64 卦名与八卦；奇门查九星/八门/八神/阴阳遁/三元/马星/空亡/门迫。六术共用绝对化措辞拦截。
- 默认 `max_tokens` 来自服务端环境变量；紫微最低 2400、大六壬最低 2200、奇门 2600、梅花 2200、小六壬 1800，即使请求传了更小的值也会被抬高，上限仍为 8192。

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
- 追问仍走同一套 SSE `meta / delta / usage / done / error`，但输出不要求四段标题；服务端用 `AuditFollowup` 做事实硬校验。
- 轮次按成功完成的问答计：模型报错或审计拦截时，前端不把这一问计入 6 轮，可在原盘直接重试；服务端按 `history` 里的用户轮次校验，超过 6 轮返回 400。

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
